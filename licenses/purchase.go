package licenses

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/giantliao/beatles-client-lib/clientwallet"
	"github.com/giantliao/beatles-client-lib/coin"
	"github.com/giantliao/beatles-client-lib/config"
	"github.com/giantliao/beatles-client-lib/db"
	"github.com/giantliao/beatles-protocol/licenses"
	"github.com/giantliao/beatles-protocol/meta"
	"github.com/kprc/nbsnetwork/tools/httputil"
	"log"
)

type ClientLicenseRenew struct {
	Transaction *common.Hash        `json:"transaction"`
	Name        string              `json:"name"`
	Email       string              `json:"email"`
	Cell        string              `json:"cell"`
	price       *config.ClientPrice `json:"price"`
	license     *licenses.License   `json:"license"`
}

func NewClientLicenseRenew(price *config.ClientPrice, name, email, cell string) *ClientLicenseRenew {
	return &ClientLicenseRenew{price: price, Name: name, Cell: cell, Email: email}
}

func (clr *ClientLicenseRenew) Buy() error {
	if clr.price == nil || clr.price.Sig == nil {
		return errors.New("please get price")
	}
	w, err := clientwallet.GetWallet()
	if err != nil {
		return err
	}

	cfg := config.GetCBtlc()
	to := common.HexToAddress(cfg.BeatlesEthAddr)
	var tx *common.Hash
	ctx := &clr.price.Sig.Content
	if ctx.PayTyp == licenses.PayTypETH {
		tx, err = w.SendToWithNonce(to, ctx.TotalPrice, ctx.Nonce, clr.price.Gas)
		if err != nil {
			return err
		}
	} else {
		tx, err = coin.GetBTLCoinToken().BtlCoinTransfer(to, ctx.TotalPrice, w.PrivKey())
		if err != nil {
			return err
		}
	}

	clr.Transaction = tx

	txdb := db.GetClientTransactionDb()
	errdb := txdb.Insert(*clr.Transaction, clr.price, clr.Name, clr.Email, clr.Cell)
	if errdb != nil {
		log.Println("!!!!import!!!, transaction insert into db failed:", errdb)
		log.Println("failed transaction is,", clr.Transaction.String(), "\r\n", clr.price.Sig.String())

		return errdb
	}

	return nil
}

func (clr *ClientLicenseRenew) GetLicense() *licenses.License {
	if clr.price == nil || clr.price.Sig == nil || clr.Transaction == nil {
		log.Println("please invoke buy first")
		return nil
	}

	lr := &licenses.LicenseRenew{}
	lr.TXSig = *clr.price.Sig
	lr.TxStr = clr.Transaction.String()
	lr.Name = clr.Name
	lr.Email = clr.Email
	lr.Cell = clr.Cell

	w, err := clientwallet.GetWallet()
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	var (
		aesk      []byte
		cipherTxt []byte
	)
	cfg := config.GetCBtlc()

	aesk, err = w.AesKey2(cfg.BeatlesMasterAddr)
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	cipherTxt, err = lr.Marshal(aesk)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	m := meta.Meta{}
	m.Marshal(w.BtlAddress().String(), cipherTxt)

	var (
		resp string
		code int
	)

	flag := false
	for i := 0; i < len(cfg.Miners); i++ {
		url := cfg.GetPurchasePath(cfg.Miners[i].Ipv4Addr, cfg.Miners[i].Port-1)

		resp, code, err = httputil.Post(url, m.ContentS, true)
		if err != nil || code != 200 || resp == "" {
			continue
		} else {
			flag = true
			break
		}
	}

	if !flag {
		log.Println("no response")
		return nil
	}

	clr.license = clr.UnPackResp(aesk, resp)

	if clr.price.Sig.Content.Receiver == w.BtlAddress() {
		cfg.MemLicense = clr.license

		licensedb := db.GetClientLicenseDb()
		errdb := licensedb.Insert(*clr.Transaction, clr.license)
		if errdb != nil {
			log.Println("!!!!import log!!!! save license failed\r\n", clr.Transaction.String(), "\r\n")
			log.Println("license:", clr.license.String())
		}
	} else {
		log.Printf("license for %s: \r\n%s\r\n", clr.license.Content.Receiver, clr.license.String())
	}

	tdb := db.GetClientTransactionDb()
	tdb.Use(clr.Transaction)

	return clr.license
}

func (clr *ClientLicenseRenew) UnPackResp(key []byte, respstr string) *licenses.License {
	m := meta.Meta{ContentS: respstr}
	_, cipherTxt, err := m.UnMarshal()
	if err != nil || len(cipherTxt) == 0 {
		return nil
	}

	nps := &licenses.License{}

	if err = nps.UnMarshal(key, cipherTxt); err != nil {
		return nil
	}

	return nps
}
