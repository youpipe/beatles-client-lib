package licenses

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/giantliao/beatles-client-lib/clientwallet"
	"github.com/giantliao/beatles-client-lib/config"
	"github.com/giantliao/beatles-client-lib/db"
	"github.com/giantliao/beatles-protocol/licenses"
	"github.com/giantliao/beatles-protocol/meta"
	"github.com/kprc/nbsnetwork/tools/httputil"
	"log"
)

type ClientLicenseRenew struct {
	Transaction *common.Hash
	Name        string
	Email       string
	Cell        string
	price       *config.ClientPrice
	license     *licenses.License
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
	tx, err = w.SendToWithNonce(to, ctx.TotalEth, ctx.Nonce, clr.price.Gas)
	if err != nil {
		return err
	}

	clr.Transaction = tx

	txdb := db.GetClientTransactionDb()
	errdb := txdb.Insert(*clr.Transaction, clr.price.Sig)
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
	lr.EthTransaction = *clr.Transaction
	lr.Name = clr.Name
	lr.Email = clr.Email
	lr.Cell = clr.Cell

	w, err := clientwallet.GetWallet()
	if err != nil {
		return nil
	}
	var (
		aesk      []byte
		cipherTxt []byte
	)
	cfg := config.GetCBtlc()

	aesk, err = w.AesKey2(cfg.BeatlesMasterAddr)
	if err != nil {
		return nil
	}
	cipherTxt, err = lr.Marshal(aesk)
	if err != nil {
		return nil
	}

	m := meta.Meta{}
	m.Marshal(cfg.BeatlesMasterAddr.String(), cipherTxt)

	var (
		resp string
		code int
	)

	for i := 0; i < len(cfg.Miners); i++ {
		url := cfg.GetPurchasePath(cfg.Miners[i].Ipv4Addr, cfg.Miners[i].Port-1)

		resp, code, err = httputil.Post(url, m.ContentS, true)
		if err != nil || code != 200 || resp == "" {
			continue
		} else {
			break
		}
	}

	if resp == "" {
		return nil
	}

	clr.license = clr.UnPackResp(aesk, resp)

	cfg.MemLicense = clr.license

	licensedb := db.GetClientLicenseDb()
	errdb := licensedb.Insert(*clr.Transaction, clr.license)
	if errdb != nil {
		log.Println("!!!!import log!!!! save license failed\r\n", clr.Transaction.String(), "\r\n")
		log.Println("license:", clr.license.String())
	}

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
