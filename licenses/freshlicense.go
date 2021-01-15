package licenses

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/giantliao/beatles-client-lib/clientwallet"
	"github.com/giantliao/beatles-client-lib/config"
	"github.com/giantliao/beatles-client-lib/db"
	"github.com/giantliao/beatles-protocol/licenses"
	"github.com/giantliao/beatles-protocol/meta"
	"github.com/kprc/nbsnetwork/tools/httputil"
	"github.com/pkg/errors"
	"log"
)

type ClientFreshLicense struct {
	req *licenses.FreshLicenseReq
	Flr *licenses.FreshLicensResult
}

func (cfl *ClientFreshLicense) FreshLicense() error {
	w, err := clientwallet.GetWallet()
	if err != nil {
		return err
	}

	var (
		aesk      []byte
		cipherTxt []byte
	)

	cfg := config.GetCBtlc()
	aesk, err = w.AesKey2(cfg.BeatlesMasterAddr)
	if err != nil {
		return err
	}

	if cfl.req == nil {
		cfl.req = &licenses.FreshLicenseReq{
			Receiver: w.BtlAddress(),
		}
	}

	cipherTxt, err = cfl.req.Marshal(aesk)
	if err != nil {
		return err
	}
	m := &meta.Meta{}
	m.Marshal(w.BtlAddress().String(), cipherTxt)

	var (
		resp string
		code int
	)

	flag := false
	for i := 0; i < len(cfg.Miners); i++ {
		url := cfg.GetFreshLicensePath(cfg.Miners[i].Ipv4Addr, cfg.Miners[i].Port-1)
		resp, code, err = httputil.Post(url, m.ContentS, true)

		if err != nil || code != 200 || resp == "" {
			continue
		} else {
			flag = true
			break
		}

	}
	if !flag {
		return errors.New("no license")
	}

	cfl.Flr = cfl.unPackResp(aesk, resp)
	if cfl.Flr == nil {
		return errors.New("no license")
	}

	cfg.MemLicense = &cfl.Flr.License

	log.Println("=========>", cfl.Flr.License.String())

	hash := common.HexToHash(cfl.Flr.TxStr)

	licensedb := db.GetClientLicenseDb()
	errdb := licensedb.Insert(hash, cfg.MemLicense)
	if errdb != nil {
		log.Println("fresh license success, but save failed\r\n", cfl.Flr.TxStr, "\r\n")
	}

	tdb := db.GetClientTransactionDb()
	tdb.Use(&hash)

	return nil
}

func (cfl *ClientFreshLicense) unPackResp(key []byte, respstr string) *licenses.FreshLicensResult {
	m := &meta.Meta{ContentS: respstr}
	_, cipherTxt, err := m.UnMarshal()
	if err != nil || len(cipherTxt) == 0 {
		return nil
	}

	flr := &licenses.FreshLicensResult{}

	if err = flr.UnMarshal(key, cipherTxt); err != nil {
		return nil
	}

	return flr

}
