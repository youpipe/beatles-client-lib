package cmdservice

import (
	"github.com/giantliao/beatles-client-lib/resource/pacserver"
	"github.com/giantliao/beatles-client-lib/streamserver"
	"github.com/giantliao/beatles-mac-client/setting"
	"google.golang.org/grpc"
	"sync"

	"net"

	"errors"
	"google.golang.org/grpc/reflection"
	"log"

	"github.com/giantliao/beatles-client-lib/app/cmdpb"
	"github.com/giantliao/beatles-client-lib/app/cmdservice/api"
	"github.com/giantliao/beatles-client-lib/config"
)

type cmdServer struct {
	localaddr  string
	grpcServer *grpc.Server
}

type CmdServerInter interface {
	StartCmdService()
	StopCmdService()
}

var (
	cmdServerInst     CmdServerInter
	cmdServerInstLock sync.Mutex
)

func GetCmdServerInst() CmdServerInter {
	if cmdServerInst == nil {
		cmdServerInstLock.Lock()
		defer cmdServerInstLock.Unlock()
		if cmdServerInst == nil {
			cfg := config.GetCBtlc()
			cmdServerInst = &cmdServer{localaddr: cfg.CmdListenPort}
		}
	}

	return cmdServerInst
}

func (cs *cmdServer) checklocaladdress() error {
	if cs.localaddr == "" {
		return errors.New("No Server Listen address")
	}

	return nil
}

func (cs *cmdServer) StartCmdService() {
	if err := cs.checklocaladdress(); err != nil {
		log.Fatal("Start Cmd Service Failed", err)
		return
	}

	lis, err := net.Listen("tcp", cs.localaddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	cs.grpcServer = grpc.NewServer()

	cmdpb.RegisterDefaultcmdsrvServer(cs.grpcServer, &api.CmdDefaultServer{stop})
	cmdpb.RegisterStringopsrvServer(cs.grpcServer, &api.CmdStringOPSrv{})

	reflection.Register(cs.grpcServer)
	log.Println("Commamd line server will start at", cs.localaddr)
	if err := cs.grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}

func (cs *cmdServer) StopCmdService() {
	config.GetCBtlc().Save()
	//server.DNSServerStop()
	//dohserver.GetDohDaemonServer().ShutDown()
	//mem.MemStateStop()
	pacserver.StopWebDaemon()
	streamserver.StopStreamserver()
	setting.ClearProxy()
	cs.grpcServer.Stop()
	log.Println("Command line server stoped")
}

var (
	flag         bool
	stopflagLock sync.Mutex
)

func stop() {

	if flag {

		return
	}

	if !flag {
		stopflagLock.Lock()
		defer stopflagLock.Unlock()

		if flag {
			return
		}

		flag = true
	}

	//httpservice.StopWebDaemon()
	GetCmdServerInst().StopCmdService()

}
