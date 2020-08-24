package clientwallet

import (
	"errors"
	"github.com/giantliao/beatles-client-lib/config"
	"github.com/kprc/libeth/wallet"
	"github.com/kprc/nbsnetwork/tools"
)

var (
	beatlesClientWallet wallet.WalletIntf
)

func GetWallet() (wallet.WalletIntf, error) {
	if beatlesClientWallet == nil {
		return nil, errors.New("no wallet in memory, please load it first")
	}

	return beatlesClientWallet, nil

}

func newWallet(auth, savePath, remoteEth string) wallet.WalletIntf {
	w := wallet.CreateWallet(savePath, remoteEth)

	if w == nil {
		return nil
	}

	w.Save(auth)

	return w

}

func LoadWallet(auth string) error {
	cfg := config.GetCBtlc()

	if !tools.FileExists(cfg.GetWalletSavePath()) {
		beatlesClientWallet = newWallet(auth, cfg.GetWalletSavePath(), cfg.EthAccPoint)
		if beatlesClientWallet == nil {
			return errors.New("create wallet failed")
		}
	} else {
		var err error
		beatlesClientWallet, err = wallet.RecoverWallet(cfg.GetWalletSavePath(), cfg.EthAccPoint, auth)
		if err != nil {
			return errors.New("load wallet failed: " + err.Error())
		}
	}

	return nil
}
