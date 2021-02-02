package clientwallet

import (
	"github.com/giantliao/beatles-client-lib/coin"
	"github.com/giantliao/beatles-client-lib/config"
	"github.com/kprc/libeth/wallet"
	"math/big"
)

type BeetleBalance struct {
	Eth float64		`json:"eth"`
	BtlcGas float64	`json:"btlc_gas"`
	Btlc float64	`json:"btlc"`
}

type WalletInfo struct {
	EthAddr string		`json:"eth_addr"`
	BeetleAddr string	`json:"beetle_addr"`
	TrxAddr string		`json:"trx_addr"`
}

func GetWalletInfo() (*WalletInfo,error) {
	w,err:=GetWallet()
	if err!=nil{
		return nil, err
	}

	wi:=&WalletInfo{
		EthAddr: w.AccountString(),
		BeetleAddr: w.BtlAddress().String(),
		TrxAddr: "",
	}

	return wi,nil
}

func GetBalance() (*BeetleBalance,error) {
	w,err:=GetWallet()
	if err!=nil{
		return nil, err
	}

	b:=&BeetleBalance{}

	b.Eth,err = w.BalanceOf(true)
	if err!=nil{
		return nil,err
	}
	b.BtlcGas, err=w.BalanceOfGas(config.GetCBtlc().BTLCAccessPoint)
	if err!=nil{
		return nil,err
	}

	var btlc *big.Int

	btlc, err = coin.GetBTLCoinToken().BtlCoinBalance(w.Address())
	if err!=nil{
		return nil,err
	}
	b.Btlc = wallet.BalanceHuman(btlc)

	return b,nil

}