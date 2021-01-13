package coin

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/giantliao/beatles-client-lib/config"
	"github.com/giantliao/beatles-master/contract"
	"github.com/kprc/libeth/wallet"
	"sync"

	"math/big"
)


type BTLCoinToken struct {
	ethAccessPoint string
	coinAddr string
}

var gBTLCoinTokenInst *BTLCoinToken
var gBTLCoinTokenLock sync.Mutex

func GetBTLCoinToken() *BTLCoinToken {
	if gBTLCoinTokenInst != nil{
		return gBTLCoinTokenInst
	}

	gBTLCoinTokenLock.Lock()
	defer gBTLCoinTokenLock.Unlock()

	if gBTLCoinTokenInst != nil{
		return gBTLCoinTokenInst
	}

	cfg:=config.GetCBtlc()

	gBTLCoinTokenInst = &BTLCoinToken{
		ethAccessPoint: cfg.BTLCAccessPoint,
		coinAddr: cfg.BTLCoinAddr,
	}

	return gBTLCoinTokenInst

}

func (bcw *BTLCoinToken)BtlCoinBalance(addr common.Address) (*big.Int,error)  {
	ec, err := ethclient.Dial(bcw.ethAccessPoint)
	if err != nil {
		return nil,err
	}
	defer ec.Close()
	var btlc *contract.BtlCoin
	btlc,err=contract.NewBtlCoin(common.HexToAddress(bcw.coinAddr),ec)
	if err!=nil{
		return nil, err
	}
	return btlc.BalanceOf(nil,addr)
}

func (bcw *BTLCoinToken)BtlCoinTransfer(toAddr common.Address, tokenNum float64, key *ecdsa.PrivateKey) (hashptr *common.Hash,err error) {
	ec, err := ethclient.Dial(bcw.ethAccessPoint)
	if err != nil {
		return nil,err
	}
	defer ec.Close()
	var btlc *contract.BtlCoin
	btlc,err = contract.NewBtlCoin(common.HexToAddress(bcw.coinAddr),ec)
	if err!=nil {
		return nil, err
	}

	opts:=bind.NewKeyedTransactor(key)
	val:=wallet.BalanceEth(tokenNum)

	var tx *types.Transaction

	tx,err = btlc.Transfer(opts,toAddr,val)
	if err!=nil{
		fmt.Println("BTLCoin Transer error",err.Error())
		return nil,err
	}

	hash := tx.Hash()

	return &hash,nil
}

//func TransferERCToken(target string, tokenNo float64, key *ecdsa.PrivateKey) (string, error) {
//
//	t, err := config.SysEthConfig.NewTokenClient()
//	if err != nil {
//		fmt.Println("[TransferERCToken]: tokenConn err:", err.Error())
//		return "", err
//	}
//	defer t.Close()
//
//	opts := bind.NewKeyedTransactor(key)
//	val := util.BalanceEth(tokenNo)
//
//	fmt.Printf("\n----->%.2f", util.BalanceHuman(val))
//
//	tx, err := t.Transfer(opts, common.HexToAddress(target), val)
//	if err != nil {
//		fmt.Println("[TransferERCToken]: Transfer err:", err.Error())
//		return "", err
//	}
//
//	return tx.Hash().Hex(), nil
//}