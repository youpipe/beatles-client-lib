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
	"strconv"

	"time"
)

type CmdStringOPSrv struct {
}

func (cso *CmdStringOPSrv) StringOpDo(cxt context.Context, so *cmdpb.StringOP) (*cmdpb.DefaultResp, error) {
	msg := ""
	switch so.Op {
	case cmdcommon.CMD_RUN:
		msg = cso.run(so.Param[0])
	case cmdcommon.CMD_SHOW_ETH_PRICE:
		msg = cso.ethPrice(so.Param[0])
	case cmdcommon.CMD_ETH_BUY:
		msg = cso.ethBuy(so.Param[0], so.Param[1], so.Param[2])
	case cmdcommon.CMD_ETH_RENEW_LICENSE:
		msg = cso.ethBuyLicense(so.Param[0])
	default:
		return encapResp("Command Not Found"), nil
	}

	return encapResp(msg), nil
}

func (cso *CmdStringOPSrv) run(passwd string) string {

	err := clientwallet.LoadWallet(passwd)
	if err != nil {
		return err.Error()
	}

	cfg := config.GetCBtlc()

	if len(cfg.Miners) == 0 {
		err := bootstrap.UpdateBootstrap()
		if err != nil {
			return err.Error()
		}
	}
	cfg.Save()

	return "vpn started"
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

//func createAccount(passwd string) string {
//	err := chatcrypt.GenEd25519KeyAndSave(passwd)
//	if err != nil {
//		return "create account failed"
//	}
//
//	chatcrypt.LoadKey(passwd)
//
//	addr := address.ToAddress(config.GetCCC().PubKey).String()
//
//	return "Address: " + addr
//}
//
//func loadAccount(passwd string) string {
//
//	chatcrypt.LoadKey(passwd)
//
//	addr := address.ToAddress(config.GetCCC().PubKey).String()
//
//	return "load account success! \r\nAddress: " + addr
//}
//
//func regUser(alias string, timeInterval string) string {
//	cfg := config.GetCCC()
//	if cfg.PrivKey == nil {
//		return "Please load account first"
//	}
//
//	tv, err := strconv.Atoi(timeInterval)
//	if err != nil {
//		return err.Error()
//	}
//
//	if err = chatmeta.RegChat(alias, tv); err != nil {
//		return err.Error()
//	}
//
//	msg := "Registered success"
//	msg += fmt.Sprintf("Name:%-30s ExpireTime:%-30s",
//		cfg.SP.SignText.AliasName,
//		int64time2string(cfg.SP.SignText.ExpireTime))
//
//	return msg
//}
//
//func addFriend(addr string) string {
//	cfg := config.GetCCC()
//	if cfg.SP == nil {
//		return "Please register first"
//	}
//
//	if err := chatmeta.AddFriend(address.ChatAddress(addr)); err != nil {
//		return err.Error()
//	}
//
//	return "Add " + addr + " friend success"
//}
//
//func delFriend(addr string) string {
//	cfg := config.GetCCC()
//	if cfg.SP == nil {
//		return "Please register first"
//	}
//
//	if err := chatmeta.DelFriend(address.ChatAddress(addr)); err != nil {
//		return err.Error()
//	}
//
//	return "Del " + addr + " friend success"
//}
//
//func createGroup(name string) string {
//	cfg := config.GetCCC()
//
//	if cfg.SP == nil {
//		return "Please register first"
//	}
//
//	if err := chatmeta.CreateGroup(name); err != nil {
//		return err.Error()
//	}
//
//	return "Create group " + name + " success"
//}
//
//func delGroup(gid string) string {
//	cfg := config.GetCCC()
//
//	if cfg.SP == nil {
//		return "Please register first"
//	}
//	if !groupid.GrpID(gid).IsValid() {
//		return "not a group id"
//	}
//
//	if err := chatmeta.DelGroup(groupid.GrpID(gid)); err != nil {
//		return err.Error()
//	}
//
//	return "Delete group " + gid + " success"
//}
//
//func joinGroup(groupId string, userid string) string {
//	cfg := config.GetCCC()
//	if cfg.SP == nil {
//		return "Please register first"
//	}
//
//	if err := chatmeta.JoinGroup(groupid.GrpID(groupId), userid); err != nil {
//		return err.Error()
//	}
//
//	return "Join group success"
//
//}
//
//func quitGroup(groupId string, userid string) string {
//	cfg := config.GetCCC()
//	if cfg.SP == nil {
//		return "Please register first"
//	}
//
//	if err := chatmeta.QuitGroup(groupid.GrpID(groupId), userid); err != nil {
//		return err.Error()
//	}
//
//	return "Quit group success"
//
//}

func int64time2string(t int64) string {
	tm := time.Unix(t/1000, 0)
	return tm.Format("2006-01-02 15:04:05")
}
