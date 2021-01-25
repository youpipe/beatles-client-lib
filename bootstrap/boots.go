package bootstrap

import (
	"errors"
	"github.com/giantliao/beatles-client-lib/config"
	"github.com/giantliao/beatles-protocol/meta"
	"github.com/giantliao/beatles-protocol/miners"
	"github.com/giantliao/beatles-protocol/token"
	"github.com/kprc/libgithub"
	"log"
)

func DownloadBootstrap() (content string, err error) {
	cfg := config.GetCBtlc()

	for i := 0; i < len(cfg.GithubAddress); i++ {
		content, err = downloadBootstrap(cfg.GithubAddress[i])
		if err == nil {
			return
		}
	}

	return "", errors.New("Can't download from github")
}

func downloadBootstrap(ap *miners.GithubDownLoadPoint) (content string, err error) {
	gc := libgithub.NewGithubClient(token.TokenRevert(ap.ReadToken), ap.Owner, ap.Repository, ap.Path, "", "")

	content, _, err = gc.GetContent()
	if err != nil {
		log.Println("download bootstrap failed from : ", ap.String(), err)
		return "", err
	}
	return
}

func UpdateBootstrap() error {
	contents, err := DownloadBootstrap()
	if err != nil {
		return err
	}

	m := &meta.Meta{ContentS: contents}
	var (
		ciphertxt []byte
	)

	_, ciphertxt, err = m.UnMarshal()
	if err != nil {
		return err
	}

	btms := &miners.BootsTrapMiners{}

	err = btms.UnMarshal(miners.SecKey(), ciphertxt)
	if err != nil {
		return err
	}

	cfg := config.GetCBtlc()
	cfg.BeatlesEthAddr = btms.BeatlesEthAddr
	cfg.BeatlesMasterAddr = btms.BeatlesMasterAddr
	cfg.BeatlesTrxAddr = btms.BeatlesTrxAddr
	cfg.EthAccPoint = btms.EthAccPoint
	cfg.TrxAccPoint = btms.TrxAccPoint
	cfg.BTLCoinAddr = btms.BTLCoinAddr
	cfg.BTLCoinPrice = btms.BTLCPrice
	cfg.BTLCAccessPoint = btms.BtlcAccPoint

	//update bootstrap
	cfg.GithubAddress = btms.NextDownloadPoint

	var boots []*miners.Miner
	//update miners
	for i:=0;i<len(btms.Boots);i++{
		flag:=false
		for j:=0;j<len(cfg.Miners);j++{
			if btms.Boots[i].MinerId == cfg.Miners[j].MinerId{
				cfg.Miners[j] = btms.Boots[i]
				flag = true
				break
			}
		}
		if !flag{
			boots = append(boots,btms.Boots[i])
		}
	}

	for i:=0;i<len(boots);i++{
		cfg.Miners = append(cfg.Miners,boots[i])
	}

	cfg.Save()

	return nil
}
