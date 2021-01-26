package ping

import (
	"crypto/rand"
	"errors"
	"github.com/giantliao/beatles-client-lib/clientwallet"
	"github.com/giantliao/beatles-client-lib/config"
	"github.com/giantliao/beatles-protocol/meta"
	"github.com/kprc/nbsnetwork/tools"
	"github.com/kprc/nbsnetwork/tools/httputil"
	"log"
)

func Ping(ip string,port int) (int64,error) {
	w, err := clientwallet.GetWallet()
	if err != nil {
		return 0,err
	}

	cfg := config.GetCBtlc()

	buf:=make([]byte,32)
	_,err = rand.Read(buf)
	if err!=nil{
		return 0,err
	}

	m := meta.Meta{}

	m.Marshal(w.BtlAddress().String(), buf)

	url:=cfg.GetPingPath(ip,port-1)

	hp:=&httputil.HttpPost{Protect: config.ProtectFD,Blog: true,DialTimeout: 2,ConnTimeout: 2}

	time1:=tools.GetNowMsTime()
	var (
		resp string
		code int
	)
	resp, code, err = hp.ProtectPost(url, m.ContentS)
	if err != nil || code != 200 || resp == "" {
		if err!=nil{
			log.Println("--------->",err.Error())
		}

		return 0,errors.New("post failed")
	}

	time2:=tools.GetNowMsTime()

	return (time2-time1)/3,nil

}

