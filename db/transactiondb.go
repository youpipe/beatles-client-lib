package db

import (
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/giantliao/beatles-client-lib/config"
	"github.com/kprc/nbsnetwork/db"
	"github.com/kprc/nbsnetwork/tools"
	"sync"
)

type ClientTransactionDb struct {
	db.NbsDbInter
	dbLock sync.Mutex
	cursor *db.DBCusor
}

type ClientTranstionItem struct {
	Tx         common.Hash         `json:"-"`
	Price      *config.ClientPrice `json:"price"`
	Name       string              `json:"name"`
	Email      string              `json:"email"`
	Cell       string              `json:"cell"`
	Used       bool                `json:"used"`
	CreateTime int64               `json:"create_time"`
	UpdateTime int64               `json:"update_time"`
}

func (cti *ClientTranstionItem) String() string {
	msg := "key: " + cti.Tx.String() + "\r\n"

	j, _ := json.MarshalIndent(*cti, " ", "\t")

	msg += string(j)

	return msg
}

var (
	clientTransactionStore     *ClientTransactionDb
	clientTransactionStoreLock sync.Mutex
)

func newClientTransactionDb() *ClientTransactionDb {
	cfg := config.GetCBtlc()
	db := db.NewFileDb(cfg.GetTransactionDBPath()).Load()

	return &ClientTransactionDb{NbsDbInter: db}
}

func GetClientTransactionDb() *ClientTransactionDb {
	if clientTransactionStore == nil {
		clientTransactionStoreLock.Lock()
		defer clientTransactionStoreLock.Unlock()

		if clientTransactionStore == nil {
			clientTransactionStore = newClientTransactionDb()
		}
	}
	return clientTransactionStore
}

func (ctdb *ClientTransactionDb) Insert(txid common.Hash, price *config.ClientPrice, name, email, cell string) error {
	ctdb.dbLock.Lock()
	defer ctdb.dbLock.Unlock()

	now := tools.GetNowMsTime()

	if _, err := ctdb.NbsDbInter.Find(txid.String()); err != nil {
		ct := &ClientTranstionItem{Price: price, Name: name, Cell: cell, Email: email}
		ct.CreateTime = now
		ct.UpdateTime = now

		j, _ := json.Marshal(*ct)
		ctdb.NbsDbInter.Insert(txid.String(), string(j))

	} else {
		return errors.New("key is existed, row id is " + txid.String())
	}

	return nil

}

func (ctdb *ClientTransactionDb) Find(txid common.Hash) *ClientTranstionItem {
	ctdb.dbLock.Lock()
	defer ctdb.dbLock.Unlock()

	if v, err := ctdb.NbsDbInter.Find(txid.String()); err != nil {
		return nil
	} else {
		ci := &ClientTranstionItem{}

		err = json.Unmarshal([]byte(v), ci)

		if err != nil {
			return nil
		}

		ci.Tx = txid

		return ci
	}
}

func (ctdb *ClientTransactionDb) Save() {
	ctdb.dbLock.Lock()
	defer ctdb.dbLock.Unlock()

	ctdb.NbsDbInter.Save()
}

func (ctdb *ClientTransactionDb) Iterator() {
	ctdb.dbLock.Lock()
	defer ctdb.dbLock.Unlock()

	ctdb.cursor = ctdb.NbsDbInter.DBIterator()
}

func (ctdb *ClientTransactionDb) Next() (txid *common.Hash, ci *ClientTranstionItem, err error) {
	if ctdb.cursor == nil {
		return nil, nil, errors.New("initialize cursor first")
	}
	ctdb.dbLock.Lock()
	k, v := ctdb.cursor.Next()
	if k == "" {
		ctdb.dbLock.Unlock()
		return nil, nil, errors.New("no transaction in list")
	}
	ctdb.dbLock.Unlock()

	ci = &ClientTranstionItem{}
	if err := json.Unmarshal([]byte(v), ci); err != nil {
		return nil, nil, err
	}

	id := common.HexToHash(k)
	txid = &id

	ci.Tx = id

	return
}

func (ctdb *ClientTransactionDb) Use(tx *common.Hash) error {
	ctdb.dbLock.Lock()
	defer ctdb.dbLock.Unlock()

	if v, err := ctdb.NbsDbInter.Find(tx.String()); err != nil {
		return err
	} else {
		ci := &ClientTranstionItem{}

		err = json.Unmarshal([]byte(v), ci)

		if err != nil {
			return err
		}
		if ci.Used {
			return errors.New("transaction is used")
		}

		ci.Used = true

		j, _ := json.Marshal(*ci)

		ctdb.NbsDbInter.Update(tx.String(), string(j))

		return nil
	}
}

func (ctdb *ClientTransactionDb) FindLatest() (*ClientTranstionItem, error) {
	ctdb.Iterator()

	var latest *ClientTranstionItem

	for {
		k, v, err := ctdb.Next()
		if k == nil || err != nil {
			break
		}

		if v.Used {
			continue
		}

		if latest == nil {
			latest = v
		} else {
			if latest.UpdateTime < v.UpdateTime {
				latest = v
			}
		}
	}

	if latest == nil {
		return nil, errors.New("no record in db")
	}

	return latest, nil

}
