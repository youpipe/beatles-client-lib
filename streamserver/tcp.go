package streamserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kprc/libeth/account"
	"syscall"
	"time"

	"github.com/giantliao/beatles-client-lib/clientwallet"
	"github.com/giantliao/beatles-client-lib/config"
	"github.com/giantliao/beatles-client-lib/db"
	"github.com/giantliao/beatles-protocol/stream"
	"github.com/kprc/libeth/wallet"
	"log"
	"net"
	"strconv"
	"sync"
)

type StreamServer struct {
	addr       string
	remoteAddr string
	quit       chan struct{}
	lis        net.Listener
	session    map[string]net.Conn
	wg         sync.WaitGroup
	w          wallet.WalletIntf
	aesKey     [32]byte
	minerId    account.BeatleAddress
	protect func(fd int32) bool
	getTarget func(conn net.Conn) (string,error)
	removeSession func(conn net.Conn)
}

type CloseConn struct {
	net.Conn
	isClosed bool
}

func (cc *CloseConn) Close() error {
	if !cc.isClosed {
		cc.isClosed = true
		return cc.Conn.Close()
	}

	return nil
}

type CloseListener struct {
	net.Listener
	isClosed bool
}

func (cl *CloseListener) Close() error {
	if !cl.isClosed {
		cl.isClosed = true
		return cl.Listener.Close()
	}
	return nil
}

func NewStreamServer(idx int,protect func(fd int32) bool, getTarget func(conn net.Conn) (string,error), removeSession func(conn net.Conn)) *StreamServer {

	cfg := config.GetCBtlc()

	addr := ":" + strconv.Itoa(cfg.StreamServerPort)

	if len(cfg.Miners) <= idx {
		log.Println("no miners in your beatles client")
		return nil
	}

	m := cfg.Miners[idx]
	remoteAddr := m.Ipv4Addr + ":" + strconv.Itoa(m.Port)

	ss := &StreamServer{
		addr: addr,
		remoteAddr: remoteAddr,
		protect: protect,
		getTarget: getTarget,
		removeSession: removeSession,
	}
	ss.quit = make(chan struct{})
	ss.session = make(map[string]net.Conn)
	ss.minerId = m.MinerId

	return ss
}

func (ss *StreamServer) StartServer() error {
	cfg := config.GetCBtlc()

	if cfg.BeatlesMasterAddr == "" {
		return errors.New("no beatles master address")
	}

	w, err := clientwallet.GetWallet()
	if err != nil {
		log.Println(err.Error())
		return err
	}
	ss.w = w

	var key []byte
	key, err = ss.w.AesKey2(ss.minerId)
	if err != nil {
		return err
	}

	copy(ss.aesKey[:], key)

	var lis net.Listener

	lis, err = net.Listen("tcp", ss.addr)
	if err != nil {
		log.Println("failed to listen on %s" + ss.addr + " : " + err.Error())
	}

	ss.lis = &CloseListener{Listener: lis}
	defer ss.lis.Close()

	log.Println("Stream Server start at ", ss.addr)

	ss.wg.Add(1)
	go ss.serve()

	ss.wg.Wait()

	return nil
}

//
func (ss *StreamServer) serve() {
	defer ss.wg.Done()

	for {
		conn, err := ss.lis.Accept()
		if err != nil {
			select {
			case <-ss.quit:
				return
			default:
				log.Println("accept error", err)
			}
		} else {
			ss.wg.Add(1)
			go func() {
				cc := &CloseConn{Conn: conn}
				ss.handleConn(cc)
			}()
		}

	}
}

func (ss *StreamServer) RemoteHandShake(conn net.Conn) (net.Conn, error) {
	s := &stream.StreamConn{Conn: conn}
	b := stream.NewStreamBuf()

	cs, err := stream.NewCipherConn(s, ss.aesKey)
	var n int

	var sh []byte
	iv := cs.(*stream.CipherConn).GetIV()
	sh = append(sh, iv[:]...)
	sh = append(sh, []byte(ss.w.BtlAddress().String())...)

	n, err = s.Write(sh)
	if err != nil || n != len(sh) {
		return nil, errors.New("write license failure")
	}

	n, err = cs.Read(b)
	if err != nil {
		log.Println("read first handshake")
		return nil, err
	}

	if b[0] == '0' {
		return cs, nil
	}

	if b[0] != '1' {
		return nil, errors.New("peer is not a server")
	}

	ldb := db.GetClientLicenseDb()
	cli := ldb.FindNewestLicense()
	if cli == nil {
		return nil, errors.New("no license")
	}

	log.Println("license", cli.String())

	j, _ := json.Marshal(*cli.License)
	n, err = cs.Write(j)
	if err != nil || n != len(j) {
		return nil, errors.New("write license failure")
	}

	n, err = cs.Read(b)
	if err != nil {
		log.Println("read second handshake")
		return nil, err
	}

	if b[0] == '0' {
		return cs, nil
	}

	return nil, errors.New("hand shake with remote server failure")

}

func (ss *StreamServer) dialTcpWithSaver(addr string) (net.Conn, error) {
	d := &net.Dialer{
		Timeout: time.Second * 2,
		Control: func(network, address string, c syscall.RawConn) error {
			if ss.protect != nil {
				return c.Control(func(fd uintptr) {
					ss.protect(int32(fd))
				})
			}
			return nil
		},
	}

	conn, err := d.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (ss *StreamServer) handleConn(conn net.Conn) {
	defer ss.wg.Done()
	defer func() {
		if ss.removeSession != nil{
			ss.removeSession(conn)
		}
		conn.Close()
	}()


	conn.(*CloseConn).Conn.(*net.TCPConn).SetKeepAlive(true)

	var (
		tgts string
		err error
		rc  net.Conn
		rcs net.Conn
	)

	if ss.protect == nil{
		rc, err = net.Dial("tcp", ss.remoteAddr)
		if err != nil {
			log.Println("failed to connect to server ", ss.remoteAddr, err)
			return
		}
	}else{
		rc,err = ss.dialTcpWithSaver(ss.remoteAddr)
		if err != nil {
			log.Println("failed to connect to server ", ss.remoteAddr, err)
			return
		}
	}

	defer rc.Close()
	rc.(*net.TCPConn).SetKeepAlive(true)

	rcs, err = ss.RemoteHandShake(rc)
	if err != nil {
		log.Println("handshake with remote failed, ", ss.remoteAddr, err)
		return
	}

	if tgts, err = ss.getTarget(conn); err != nil {
		return
	}

	if _, err = rcs.Write(ParseAddr(tgts)); err != nil {
		log.Println("failed to send target address: ", err)
		return
	}

	log.Println("proxy ", conn.RemoteAddr(), "<->", ss.remoteAddr, "<->", tgts)

	err = relay2(conn, rcs)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return
		}
		log.Println("relay error:", err)
	}
}

func relay2(left, right net.Conn) error {
	defer func() {
		right.SetDeadline(time.Now())
		left.SetDeadline(time.Now())
	}()

	go func() {
		defer func() {
			right.SetDeadline(time.Now())
			left.SetDeadline(time.Now())
		}()
		for {
			buf := stream.NewStreamBuf()
			n, err := left.Read(buf)
			if err != nil {
				fmt.Println("left->right read err", err)
				return
			}
			var nw int
			nw, err = right.Write(buf[:n])
			if n != nw || err != nil {
				fmt.Println("left->right write err", err, n, nw)
				return
			}
		}
	}()
	for {
		buf := stream.NewStreamBuf()
		n, err := right.Read(buf)
		if err != nil {
			fmt.Println("right->left read err", err)
			return err
		}
		var nw int
		nw, err = left.Write(buf[:n])
		if n != nw || err != nil {
			fmt.Println("left->right write err", err, n, nw)
			return err
		}
	}

	return nil
}

func (ss *StreamServer) StopServer() {
	close(ss.quit)
	ss.lis.Close()
}
