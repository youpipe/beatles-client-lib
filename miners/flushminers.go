package miners

import (
	"errors"
	"github.com/giantliao/beatles-client-lib/clientwallet"
	"github.com/giantliao/beatles-client-lib/config"
	"github.com/giantliao/beatles-client-lib/db"
	"github.com/giantliao/beatles-protocol/licenses"
	"github.com/giantliao/beatles-protocol/meta"
	"github.com/giantliao/beatles-protocol/miners"
	"github.com/kprc/nbsnetwork/tools/httputil"
)

type ClientMiners struct {
	license *licenses.License
	miners  []*miners.Miner
}

func NewClientMiners() *ClientMiners {
	cm := &ClientMiners{}
	cfg := config.GetCBtlc()

	if cfg.MemLicense == nil {
		ldb := db.GetClientLicenseDb()
		cli := ldb.FindNewestLicense()
		if cli == nil {
			return nil
		}
		cfg.MemLicense = cli.License
	}

	cm.license = cfg.MemLicense

	if cm.license == nil {
		return nil
	}

	return cm
}

func (cm *ClientMiners) FlushMiners() error {
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

	cipherTxt, err = cm.license.Marshal(aesk)
	if err != nil {
		return err
	}

	m := meta.Meta{}

	m.Marshal(w.BtlAddress().String(), cipherTxt)

	var (
		resp string
		code int
	)

	flag := false

	for i := 0; i < len(cfg.Miners); i++ {
		url := cfg.GetListMinerPath(cfg.Miners[i].Ipv4Addr, cfg.Miners[i].Port-1)
		resp, code, err = httputil.Post(url, m.ContentS, true)
		if err != nil || code != 200 || resp == "" {
			continue
		} else {
			flag = true
			break
		}
	}

	if !flag {
		return errors.New("no miners")
	}

	if ms := cm.UnPackResp(aesk, resp); ms == nil {
		return errors.New("unpack miners failed")
	} else {
		for i := 0; i < len(ms.Ms); i++ {
			cm.miners = append(cm.miners, &ms.Ms[i])
		}
	}

	cfg.Miners = cm.miners

	return nil

}

func (cm *ClientMiners) UnPackResp(key []byte, respstr string) *miners.BestMiners {
	m := meta.Meta{ContentS: respstr}
	_, cipherTxt, err := m.UnMarshal()
	if err != nil || len(cipherTxt) == 0 {
		return nil
	}

	ms := &miners.BestMiners{}

	if err = ms.UnMarshal(key, cipherTxt); err != nil {
		return nil
	}

	return ms
}
