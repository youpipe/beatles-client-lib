package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/giantliao/beatles-client-lib/app/cmdcommon"
	"github.com/giantliao/beatles-client-lib/app/cmdpb"
	"github.com/giantliao/beatles-client-lib/config"
	"github.com/giantliao/beatles-client-lib/db"
	"github.com/giantliao/beatles-client-lib/licenses"
	"github.com/giantliao/beatles-client-lib/ping"
	"github.com/giantliao/beatles-client-lib/streamserver"
	"github.com/giantliao/beatles-mac-client/setting"
	prolic "github.com/giantliao/beatles-protocol/licenses"
	"github.com/kprc/libeth/account"
	log "github.com/sirupsen/logrus"
	"strconv"

	"time"
)

type CmdStringOPSrv struct {
}

func (cso *CmdStringOPSrv) StringOpDo(cxt context.Context, so *cmdpb.StringOP) (*cmdpb.DefaultResp, error) {
	msg := ""
	switch so.Op {
	case cmdcommon.CMD_SHOW_ETH_PRICE:
		msg = cso.ethPrice(so.Param[0], so.Param[1], so.Param[2])
	case cmdcommon.CMD_ETH_BUY:
		msg = cso.ethBuy(so.Param[0], so.Param[1], so.Param[2])
	case cmdcommon.CMD_ETH_RENEW_LICENSE:
		msg = cso.ethBuyLicense(so.Param[0])
	case cmdcommon.CMD_SHOW_ETH_TX:
		msg = cso.ethTx(so.Param[0])
	case cmdcommon.CMD_SHOW_LICENSE:
		msg = cso.licenseShow(so.Param[0])
	case cmdcommon.CMD_START_VPN:
		msg = cso.startVpn(so.Param[0])
	case cmdcommon.CMD_VPN_MODE:
		msg = cso.setMode(so.Param[0])
	case cmdcommon.CMD_PING:
		msg = cso.pingminer(so.Param[0])
	default:
		return encapResp("Command Not Found"), nil
	}

	return encapResp(msg), nil
}

func (cso *CmdStringOPSrv) ethPrice(month string, typ string, addr string) string {
	ms, err := strconv.Atoi(month)
	if err != nil {
		return err.Error()
	}
	if ms <= 0 {
		return "month must large than 1"
	}
	var paytyp int
	paytyp, err = strconv.Atoi(typ)
	if err != nil {
		return err.Error()
	}
	if paytyp != prolic.PayTypETH && paytyp != prolic.PayTypBTLC {
		return "pay type error"
	}

	if addr != "" {
		if !account.BeatleAddress(addr).IsValid() {
			return "not a correct receiver address"
		}
	}

	var cp *licenses.CurrentPrice
	cp, err = licenses.NewCurrentPrice(int64(ms), paytyp, account.BeatleAddress(addr))
	if err != nil {
		return err.Error()
	}

	np := cp.Get()
	if np == nil {
		return "get price failed"
	}

	config.GetCBtlc().MemPrice = np

	var j []byte
	j, err = json.MarshalIndent(*np, " ", "\t")
	if err != nil {
		return err.Error()
	}

	return string(j)

}

func (cso *CmdStringOPSrv) ethBuy(name, email, cell string) string {
	cfg := config.GetCBtlc()
	if cfg.MemPrice == nil {
		return "please get price first"
	}
	lr := licenses.NewClientLicenseRenew(cfg.MemPrice, name, email, cell)

	err := lr.Buy()
	if err != nil {
		return err.Error()
	}

	log.Println(lr.String())

	tdb := db.GetClientTransactionDb()
	if v := tdb.Find(*lr.Transaction); v != nil {
		return v.String()
	} else {
		return "buy license info not in db"
	}
}

func (cso *CmdStringOPSrv) ethBuyLicense(tx string) string {
	tdb := db.GetClientTransactionDb()

	var (
		cti *db.ClientTranstionItem
		err error
	)

	if tx == "" {
		cti, err = tdb.FindLatest()
		if err != nil {
			return err.Error()
		}
	} else {
		cti = tdb.Find(common.HexToHash(tx))
		if cti == nil {
			return "not found transaction"
		}
	}
	if cti.Used {
		return "transaction is used"
	}

	clr := licenses.NewClientLicenseRenew(cti.Price, cti.Name, cti.Email, cti.Cell)
	clr.Transaction = &cti.Tx

	l := clr.GetLicense()
	if l == nil {
		return "something wrong, get license failed"
	}

	j, _ := json.MarshalIndent(*l, " ", "\t")

	return string(j)

}

func (cso *CmdStringOPSrv) ethTx(used string) string {
	msg := ""
	if u, err := strconv.ParseBool(used); err != nil {
		return err.Error()
	} else {
		tdb := db.GetClientTransactionDb()
		cursor:=tdb.Iterator()
		for {
			k, v, e := tdb.Next(cursor)
			if k == nil || e != nil {
				break
			}
			if !u && v.Used {
				continue
			}
			if msg != "" {
				msg += "\r\n"
			}
			msg += "===================================\r\n"
			msg += v.String()
		}
	}

	if msg == "" {
		return "no tx in db"
	} else {
		msg += "\r\n===================================="
	}

	return msg
}

func (cso *CmdStringOPSrv) licenseShow(history string) string {
	msg := ""
	if h, err := strconv.ParseBool(history); err != nil {
		return err.Error()
	} else {
		ldb := db.GetClientLicenseDb()
		if h {

			ldb.Iterator()
			for {
				k, v, e := ldb.Next()
				if k == nil || e != nil {
					break
				}
				if msg != "" {
					msg += "\r\n"
				}
				msg += "===================================\r\n"
				msg += v.String()
			}
		} else {
			if cli := ldb.FindNewestLicense(); cli != nil {
				msg = "===================================\r\n"
				msg += cli.String()
			}
		}
	}

	if msg == "" {
		return "no tx in db"
	} else {
		msg += "\r\n===================================="
	}

	return msg
}

func (cso *CmdStringOPSrv) startVpn(m string) string {
	idx, err := strconv.Atoi(m)
	if err != nil {
		return err.Error()
	}

	cfg := config.GetCBtlc()
	if idx >= len(cfg.Miners) || idx < -1 {
		return "miner not exists"
	}

	if streamserver.StreamServerIsStart() {
		return "server is started"
	}

	if idx == -1 {
		for i := 0; i < len(cfg.Miners); i++ {
			if cfg.Miners[i].MinerId == cfg.CurrentMiner {
				idx = i
				break
			}
		}
	}

	if idx == -1 {
		idx = 0
	}

	cfg.CurrentMiner = cfg.Miners[idx].MinerId

	cfg.Save()

	go streamserver.StartStreamServer(idx,nil,streamserver.Handshake,nil)

	setting.SetProxy(cfg.VPNMode)

	mode := "global"

	if cfg.VPNMode == 0 {
		mode = "pac"
	}

	return "start vpn success, miner ip: " + cfg.Miners[idx].Ipv4Addr + " vpnmode: " + mode
}

func (cso *CmdStringOPSrv) setMode(v string) string {
	//now:=tools.GetNowMsTime()
	m, err := strconv.Atoi(v)
	if err != nil {
		return err.Error()
	}

	if m != 1 && m != 0 {
		return "not a correct mode"
	}

	cfg := config.GetCBtlc()
	oldm := cfg.VPNMode
	cfg.VPNMode = m

	if oldm == m {
		return "nothing to do"
	}
	//log.Print("time:",tools.GetNowMsTime() - now)
	setting.SetProxy(m)
	//log.Print("time:",tools.GetNowMsTime() - now)
	cfg.Save()
	//log.Print("time:",tools.GetNowMsTime() - now)
	return "set mode success"
}

func (cso *CmdStringOPSrv)pingminer(minerid string) string  {
	if minerid != ""{
		if !account.IsValidID(minerid){
			return "miner id not correct"
		}
		cfg:=config.GetCBtlc()

		idx:=-1

		for i:=0;i<len(cfg.Miners);i++{
			if cfg.Miners[i].MinerId == account.BeatleAddress(minerid){
				idx = i
				break
			}
		}

		if idx < 0{
			return "miner not found"
		}

		tv,err:=ping.Ping(cfg.Miners[idx].Ipv4Addr,cfg.Miners[idx].Port)
		if err!=nil{
			log.Println("=====>",err.Error())
			return "ping failed"
		}

		config.AddPingTestResult(account.BeatleAddress(minerid),tv)

		return "time delay:"+strconv.FormatInt(tv,10)
	}else{
		cfg:=config.GetCBtlc()
		for i:=0;i<len(cfg.Miners);i++{
			tv,err:=ping.Ping(cfg.Miners[i].Ipv4Addr,cfg.Miners[i].Port)
			if err!=nil{
				continue
			}
			config.AddPingTestResult(cfg.Miners[i].MinerId,tv)
		}

		msg:=""

		for k,v:=range config.PingTestResult{
			if msg !=""{
				msg +="\r\n"
			}
			msg += fmt.Sprintf("%s: %d",k,v)
		}
		if msg == ""{
			return "ping failed"
		}
		return msg
	}
}


func int64time2string(t int64) string {
	tm := time.Unix(t/1000, 0)
	return tm.Format("2006-01-02 15:04:05")
}
