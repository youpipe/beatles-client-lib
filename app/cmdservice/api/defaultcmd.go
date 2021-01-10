package api

import (
	"context"
	"encoding/json"
	"github.com/giantliao/beatles-client-lib/app/cmdcommon"
	"github.com/giantliao/beatles-client-lib/app/cmdpb"
	"github.com/giantliao/beatles-client-lib/clientwallet"
	"github.com/giantliao/beatles-client-lib/config"
	"github.com/giantliao/beatles-client-lib/miners"
	"github.com/giantliao/beatles-client-lib/streamserver"
	"github.com/giantliao/beatles-mac-client/setting"

	"log"
	"strconv"
	"time"
)

type CmdDefaultServer struct {
	Stop func()
}

func (cds *CmdDefaultServer) DefaultCmdDo(ctx context.Context,
	request *cmdpb.DefaultRequest) (*cmdpb.DefaultResp, error) {

	msg := ""

	switch request.Reqid {
	case cmdcommon.CMD_STOP:
		msg = cds.stop()
	case cmdcommon.CMD_CONFIG_SHOW:
		msg = cds.configShow()
	case cmdcommon.CMD_ETH_BALANCE:
		msg = cds.ehtBalance()
	case cmdcommon.CMD_MINER_SHOW:
		msg = cds.showAllMiners()
	case cmdcommon.CMD_MINER_FLUSH:
		msg = cds.flushMiner()
	case cmdcommon.CMD_STOP_VPN:
		msg = cds.stopVpn()
	case cmdcommon.CMD_WALLET_SHOW:
		msg = cds.showWallet()
	}

	if msg == "" {
		msg = "No Results"
	}

	resp := &cmdpb.DefaultResp{}
	resp.Message = msg

	return resp, nil

}

func (cds *CmdDefaultServer) stop() string {

	go func() {
		time.Sleep(time.Second * 2)
		cds.Stop()
	}()

	return "beatles client stopped"
}

func encapResp(msg string) *cmdpb.DefaultResp {
	resp := &cmdpb.DefaultResp{}
	resp.Message = msg

	return resp
}

func (cds *CmdDefaultServer) configShow() string {
	cfg := config.GetCBtlc()

	bapc, err := json.MarshalIndent(*cfg, "", "\t")
	if err != nil {
		return "Internal error"
	}

	return string(bapc)
}

func (cds *CmdDefaultServer) ehtBalance() string {
	w, err := clientwallet.GetWallet()
	if err != nil {
		return err.Error()
	}
	var b float64
	b, err = w.BalanceOf(true)
	if err != nil {
		return err.Error()
	}

	msg := "Eth Address: " + w.AccountString()
	msg += "\r\nBeatles Address: " + w.BtlAddress().String()

	return msg + "\r\nEth Balance: " + strconv.FormatFloat(b, 'f', -1, 64)
}

func (cds *CmdDefaultServer) showAllMiners() string {
	cfg := config.GetCBtlc()

	if len(cfg.Miners) == 0 {
		return "no miner"
	}

	msg := ""

	for i:=0;i<len(cfg.Miners);i++{
		msg += cfg.Miners[i].String()
		msg += "\r\n"
	}

	return msg
}

func (cds *CmdDefaultServer) flushMiner() string {
	flushMachine := miners.NewClientMiners()
	if flushMachine == nil {
		return "may be you are no license"
	}

	if err := flushMachine.FlushMiners(); err != nil {
		return err.Error()
	}

	return "flush miners success, miner count: " + strconv.Itoa(len(config.GetCBtlc().Miners))
}

func (cds *CmdDefaultServer) stopVpn() string {

	log.Println("begin to stop vpn")

	if !streamserver.StreamServerIsStart(){
		return "vpn not started"
	}

	streamserver.StopStreamserver()
	//pacserver.StopWebDaemon()

	setting.ClearProxy()

	return "vpn stopped, disconnected from miner: "+ config.GetCBtlc().CurrentMiner.String()
}

func (cds *CmdDefaultServer) showWallet() string {
	if _, err := clientwallet.GetWallet(); err != nil {
		return err.Error()
	} else {
		var s string
		if s, err = clientwallet.ShowWallet(); err != nil {
			return err.Error()
		}

		return s
	}
}
