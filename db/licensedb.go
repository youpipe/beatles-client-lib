package db

import (
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/giantliao/beatles-client-lib/config"
	"github.com/giantliao/beatles-protocol/licenses"
	"github.com/kprc/nbsnetwork/db"
	"github.com/kprc/nbsnetwork/tools"
	"sync"
)

type ClientLicenseDb struct {
	db.NbsDbInter
	dbLock sync.Mutex
	cursor *db.DBCusor
}

type ClientLicenseItem struct {
	Tx         common.Hash       `json:"-"`
	License    *licenses.License `json:"license"`
	CreateTime int64             `json:"create_time"`
	UpdateTime int64             `json:"update_time"`
}

var (
	clientLicenseStore     *ClientLicenseDb
	clientLicenseStoreLock sync.Mutex
)

func newClientLicenseDb() *ClientLicenseDb {
	cfg := config.GetCBtlc()
	db := db.NewFileDb(cfg.GetLicenseDBPath())

	return &ClientLicenseDb{NbsDbInter: db}
}

func GetClientLicenseDb() *ClientLicenseDb {
	if clientLicenseStore == nil {
		clientLicenseStoreLock.Lock()
		defer clientLicenseStoreLock.Unlock()
		if clientLicenseStore == nil {
			clientLicenseStore = newClientLicenseDb()
		}
	}
	return clientLicenseStore
}

func (cldb *ClientLicenseDb) Insert(txid common.Hash, license *licenses.License) error {
	cldb.dbLock.Lock()
	defer cldb.dbLock.Unlock()

	now := tools.GetNowMsTime()

	if _, err := cldb.NbsDbInter.Find(txid.String()); err != nil {
		ci := &ClientLicenseItem{License: license}
		ci.CreateTime = now
		ci.UpdateTime = now

		j, _ := json.Marshal(*ci)
		cldb.NbsDbInter.Insert(txid.String(), string(j))

	} else {
		return errors.New("key is existed, row id is " + txid.String())
	}

	return nil

}

func (cldb *ClientLicenseDb) Find(txid common.Hash) *ClientLicenseItem {
	cldb.dbLock.Lock()
	defer cldb.dbLock.Unlock()

	if v, err := cldb.NbsDbInter.Find(txid.String()); err != nil {
		return nil
	} else {
		ci := &ClientLicenseItem{}

		err = json.Unmarshal([]byte(v), ci)

		if err != nil {
			return nil
		}

		ci.Tx = txid

		return ci
	}

}

func (cldb *ClientLicenseDb) Save() {
	cldb.dbLock.Lock()
	defer cldb.dbLock.Unlock()

	cldb.NbsDbInter.Save()
}

func (cldb *ClientLicenseDb) Iterator() {
	cldb.dbLock.Lock()
	defer cldb.dbLock.Unlock()

	cldb.cursor = cldb.NbsDbInter.DBIterator()
}

func (cldb *ClientLicenseDb) Next() (txid *common.Hash, ci *ClientLicenseItem, err error) {
	if cldb.cursor == nil {
		return nil, nil, errors.New("initialize cursor first")
	}
	cldb.dbLock.Lock()
	k, v := cldb.cursor.Next()
	if k == "" {
		cldb.dbLock.Unlock()
		return nil, nil, errors.New("no license in list")
	}
	cldb.dbLock.Unlock()

	ci = &ClientLicenseItem{}
	if err := json.Unmarshal([]byte(v), ci); err != nil {
		return nil, nil, err
	}

	id := common.HexToHash(k)
	txid = &id

	ci.Tx = *txid

	return
}
