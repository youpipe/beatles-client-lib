package miners

import (
	"errors"
	"github.com/giantliao/beatles-client-lib/clientwallet"
	"github.com/giantliao/beatles-client-lib/config"
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

	aesk, err = w.AesKey2(w.BtlAddress())
	if err != nil {
		return err
	}

	cipherTxt, err = cm.license.Marshal(aesk)
	if err != nil {
		return err
	}

	m := meta.Meta{}

	m.Marshal(w.BtlAddress().String(), cipherTxt)

	cfg := config.GetCBtlc()

	var (
		resp string
		code int
	)

	for i := 0; i < len(cfg.Miners); i++ {
		url := cfg.GetListMinerPath(cfg.Miners[i].Ipv4Addr, cfg.Miners[i].Port)
		resp, code, err = httputil.Post(url, m.ContentS, true)
		if err != nil || code != 200 || resp == "" {
			continue
		} else {
			break
		}
	}

	if resp == "" {
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

	if err = ms.UnMarshal(key, m.Content); err != nil {
		return nil
	}

	return ms
}
