package licenses

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/giantliao/beatles-client-lib/clientwallet"
	"github.com/giantliao/beatles-client-lib/config"
	"github.com/giantliao/beatles-protocol/licenses"
	"github.com/giantliao/beatles-protocol/meta"
	"github.com/giantliao/beatles-protocol/miners"
	"github.com/kprc/libeth/account"
	"github.com/kprc/nbsnetwork/tools/httputil"
)

type CurrentPrice struct {
	miners             []*miners.Miner
	licenseBeatlesAddr account.BeatleAddress
	nonce              uint64
	gas                float64
	fee                float64
	selfBeatlesAddr    account.BeatleAddress
	selfEthAddr        common.Address
	month              int64
}

func NewCurrentPrice(month int64) (*CurrentPrice, error) {
	cp := &CurrentPrice{month: month}

	if err := cp.init(); err != nil {
		return nil, err
	}

	return cp, nil
}

func (cp *CurrentPrice) init() error {
	cfg := config.GetCBtlc()

	cp.miners = cfg.Miners
	cp.licenseBeatlesAddr = cfg.BeatlesMasterAddr

	w, err := clientwallet.GetWallet()
	if err != nil {
		return err
	}
	cp.selfBeatlesAddr = w.BtlAddress()
	cp.selfEthAddr = w.Address()

	cp.nonce, err = w.Nonce()
	if err != nil {
		return err
	}

	cp.gas, cp.fee, err = w.Gas()
	if err != nil {
		return err
	}

	return nil

}

func (cp *CurrentPrice) Get() *config.ClientPrice {
	if cp.miners == nil {
		return nil
	}

	np := &licenses.NoncePrice{}
	np.Nonce = cp.nonce
	np.Month = cp.month
	np.Receiver = cp.selfBeatlesAddr
	np.EthAddr = cp.selfEthAddr

	w, err := clientwallet.GetWallet()
	if err != nil {
		return nil
	}
	var (
		aesk      []byte
		cipherTxt []byte
	)

	aesk, err = w.AesKey2(cp.licenseBeatlesAddr)
	if err != nil {
		return nil
	}
	cipherTxt, err = np.Marshal(aesk)
	if err != nil {
		return nil
	}

	m := &meta.Meta{}
	m.Marshal(cp.selfBeatlesAddr.String(), cipherTxt)

	var (
		resp string
		code int
	)

	cfg := config.GetCBtlc()

	flag := false
	//todo... post to random miner
	for i := 0; i < len(cp.miners); i++ {
		url := cfg.GetNoncePriceWebPath(cp.miners[i].Ipv4Addr, cp.miners[i].Port-1)
		resp, code, err = httputil.Post(url, m.ContentS, true)
		if err != nil || code != 200 || resp == "" {
			continue
		} else {
			flag = true
			break
		}
	}
	if !flag {
		return nil
	}

	sig := cp.UnPackResp(aesk, resp)

	return &config.ClientPrice{Sig: sig, Gas: cp.gas, EstimatedFee: cp.fee}

}

func (cp *CurrentPrice) UnPackResp(key []byte, respstr string) *licenses.NoncePriceSig {
	m := meta.Meta{ContentS: respstr}
	peerAddr, cipherTxt, err := m.UnMarshal()
	if err != nil || peerAddr != cp.licenseBeatlesAddr.String() || len(cipherTxt) == 0 {
		return nil
	}

	nps := &licenses.NoncePriceSig{}

	if err = nps.UnMarshal(key, cipherTxt); err != nil {
		return nil
	}

	return nps
}
