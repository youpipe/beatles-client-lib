package api

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/giantliao/beatles-client-lib/app/cmdcommon"
	"github.com/giantliao/beatles-client-lib/app/cmdpb"
	"github.com/giantliao/beatles-client-lib/bootstrap"
	"github.com/giantliao/beatles-client-lib/clientwallet"
	"github.com/giantliao/beatles-client-lib/config"
	"github.com/giantliao/beatles-client-lib/db"
	"github.com/giantliao/beatles-client-lib/licenses"
	"github.com/giantliao/beatles-client-lib/resource/pacserver"
	"github.com/giantliao/beatles-client-lib/streamserver"
	"strconv"

	"time"
)

type CmdStringOPSrv struct {
}

func (cso *CmdStringOPSrv) StringOpDo(cxt context.Context, so *cmdpb.StringOP) (*cmdpb.DefaultResp, error) {
	msg := ""
	switch so.Op {
	case cmdcommon.CMD_START:
		msg = cso.start(so.Param[0])
	case cmdcommon.CMD_SHOW_ETH_PRICE:
		msg = cso.ethPrice(so.Param[0])
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
	default:
		return encapResp("Command Not Found"), nil
	}

	return encapResp(msg), nil
}

func (cso *CmdStringOPSrv) start(passwd string) string {
	cfg := config.GetCBtlc()

	if len(cfg.Miners) == 0 {
		err := bootstrap.UpdateBootstrap()
		if err != nil {
			return err.Error()
		}
	}

	err := clientwallet.LoadWallet(passwd)
	if err != nil {
		return err.Error()
	}
	cfg.Save()

	//go streamserver.StartStreamServer()

	return "client ready"
}

func (cso *CmdStringOPSrv) ethPrice(month string) string {
	ms, err := strconv.Atoi(month)
	if err != nil {
		return err.Error()
	}
	if ms <= 0 {
		return "month must large than 1"
	}

	var cp *licenses.CurrentPrice
	cp, err = licenses.NewCurrentPrice(int64(ms))
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
		if cti.Used {
			return "transaction is used"
		}
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
		tdb.Iterator()
		for {
			k, v, e := tdb.Next()
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
	if idx >= len(cfg.Miners) {
		return "miner not exists"
	}

	go streamserver.StartStreamServer(idx)
	go pacserver.StartWebDaemon()

	return "start vpn success"
}

func int64time2string(t int64) string {
	tm := time.Unix(t/1000, 0)
	return tm.Format("2006-01-02 15:04:05")
}
