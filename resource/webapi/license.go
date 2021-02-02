package webapi

import (
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/giantliao/beatles-client-lib/config"
	"github.com/giantliao/beatles-client-lib/db"
	"github.com/giantliao/beatles-client-lib/licenses"
	prolic "github.com/giantliao/beatles-protocol/licenses"
	"github.com/kprc/libeth/account"
	"github.com/kprc/nbsnetwork/tools"
	"net/http"
)

func AddLicenseApi(mux *http.ServeMux)  {
	mux.HandleFunc("/api/license/isApplied",licenseIsApplied)
	mux.HandleFunc("/api/license/showLicense",showLicense)
	mux.HandleFunc("/api/license/queryPrice",queryPrice)
	mux.HandleFunc("/api/license/buyLicense",buyLicense)
	mux.HandleFunc("/api/license/renewLicense",renewLicense)
	mux.HandleFunc("/api/license/refrehLicense",refrehLicense)
	mux.HandleFunc("/api/license/showLog",showLog)
}

func licenseIsApplied(w http.ResponseWriter , r *http.Request)  {
	if r.Method != "GET"{
		w.WriteHeader(500)
		w.Write([]byte("not a correct method"))
		return
	}

	ldb:=db.GetClientLicenseDb()
	cli:=ldb.FindNewestLicense()
	if cli == nil{
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,"not found license",1)))
		return
	}

	now:=tools.GetNowMsTime()

	if cli.License.Content.ExpireTime < now{
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,"license is expired",2)))
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(SimpleResponse(0,"",0)))

}

func showLicense(w http.ResponseWriter , r *http.Request)   {
	if r.Method != "GET"{
		w.WriteHeader(500)
		w.Write([]byte("not a correct method"))
	}

	ldb:=db.GetClientLicenseDb()
	cli:=ldb.FindNewestLicense()
	if cli == nil{
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,"not found license",1)))
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(Respponse(0,"",0,cli.License)))

}

type BeetleQueryPrice struct {
	Month int 	`json:"month"`
	PayTyp int  `json:"pay_typ"`
	Receiver string `json:"receiver"`
}

func checkQueryPrice(qr *BeetleQueryPrice) error  {
	if qr == nil{
		return errors.New("parameter error")
	}
	if qr.Month <1 || qr.Month > 36{
		return errors.New("month must between 1 ~ 36")
	}

	if qr.PayTyp != prolic.PayTypETH && qr.PayTyp != prolic.PayTypBTLC {
		return errors.New("pay type error")
	}

	if qr.Receiver == ""{
		return nil
	}

	id:=account.BeatleAddress(qr.Receiver)
	if !id.IsValid(){
		return errors.New("id is not correct")
	}
	return nil
}

func queryPirceResult(qr *BeetleQueryPrice) (*config.ClientPrice,error)  {
	var cp *licenses.CurrentPrice
	var err error
	cp, err = licenses.NewCurrentPrice(int64(qr.Month), qr.PayTyp, account.BeatleAddress(qr.Receiver))
	if err != nil {
		return nil,err
	}

	np := cp.Get()
	if np == nil {
		return nil,errors.New("get price failed")
	}

	config.GetCBtlc().MemPrice = np

	return np,nil
}

func queryPrice(w http.ResponseWriter , r *http.Request)  {
	pbytes,err:=ReadReq(r)
	if err!=nil{
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	qr:=&BeetleQueryPrice{}
	err = json.Unmarshal(pbytes,qr)
	if err != nil{
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,err.Error(),1)))
		return
	}

	err = checkQueryPrice(qr)
	if err!=nil{
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,err.Error(),2)))
		return
	}

	var np *config.ClientPrice
	np,err = queryPirceResult(qr)
	if err!=nil{
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,err.Error(),3)))
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(Respponse(0,"",0,np)))

}

type BeetleBuyLicense struct {
	Name string		`json:"name"`
	Email string	`json:"email"`
	Cell string		`json:"cell"`
}

func buyLicenseResult(bl *BeetleBuyLicense) (*db.ClientTranstionItem, error) {
	cfg := config.GetCBtlc()
	lr := licenses.NewClientLicenseRenew(cfg.MemPrice, bl.Name, bl.Email, bl.Cell)

	err := lr.Buy()
	if err != nil {
		return nil,err
	}

	tdb := db.GetClientTransactionDb()
	if v := tdb.Find(*lr.Transaction); v != nil {
		//j,_:=json.Marshal(*lr)
		v.Tx = *lr.Transaction
		return v,nil
	} else {
		return nil,errors.New("buy license info not in db")
	}
}

func buyLicense(w http.ResponseWriter , r *http.Request)  {
	pbytes,err:=ReadReq(r)
	if err!=nil{
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	bl:=&BeetleBuyLicense{}
	err = json.Unmarshal(pbytes,bl)
	if err != nil{
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,err.Error(),1)))
		return
	}

	cfg := config.GetCBtlc()
	if cfg.MemPrice == nil {
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,"please get price first",2)))
		return
	}

	var item *db.ClientTranstionItem
	item,err = buyLicenseResult(bl)
	if err!=nil{
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,err.Error(),3)))
	}

	w.WriteHeader(200)
	w.Write([]byte(Respponse(0,"",0,item)))

}

type RenewLicense struct {
	Tx string `json:"tx"`
}

func findRenewLicenseTx(rl *RenewLicense) (*db.ClientTranstionItem,error)  {
	tdb := db.GetClientTransactionDb()
	var (
		cti *db.ClientTranstionItem
		err error
	)

	if rl.Tx == "" {
		cti, err = tdb.FindLatest()
		if err != nil {
			return nil,err
		}
	} else {
		cti = tdb.Find(common.HexToHash(rl.Tx))
		if cti == nil {
			return nil,errors.New("not found transaction")
		}
	}
	if cti.Used {
		return nil,errors.New("transaction is used")
	}

	return cti,nil
}

func renewLicense(w http.ResponseWriter , r *http.Request)  {
	pbytes,err:=ReadReq(r)
	if err!=nil{
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	rl:=&RenewLicense{}
	err = json.Unmarshal(pbytes,rl)
	if err != nil{
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,err.Error(),1)))
		return
	}

	cti,err:=findRenewLicenseTx(rl)
	if err!=nil{
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,err.Error(),2)))
		return
	}
	clr := licenses.NewClientLicenseRenew(cti.Price, cti.Name, cti.Email, cti.Cell)
	clr.Transaction = &cti.Tx

	l := clr.GetLicense()
	if l == nil {
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,"something wrong, get license failed",3)))
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(Respponse(0,"",0,l)))
}
func refrehLicense(w http.ResponseWriter , r *http.Request)  {
	if r.Method != "POST"{
		w.WriteHeader(500)
		w.Write([]byte("not a correct method"))
		return
	}

	cfl := licenses.ClientFreshLicense{}

	if err := cfl.FreshLicense(); err != nil {
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,err.Error(),1)))
	}

	w.WriteHeader(200)
	w.Write([]byte(Respponse(0,"",0,&cfl.Flr.License)))

	return

}

type BeetleLicenseLog struct {
	TxStr string  			`json:"tx_str"`
	*db.ClientTranstionItem
}

type BeetleNewLicenseLog struct {
	Logs []*BeetleLicenseLog	`json:"logs"`
}

func showLog(w http.ResponseWriter , r *http.Request)  {
	if r.Method != "POST"{
		w.WriteHeader(500)
		w.Write([]byte("not a correct method"))
		return
	}

	bnll:=&BeetleNewLicenseLog{}

	tdb := db.GetClientTransactionDb()
	cursor:=tdb.Iterator()

	for {
		k, v, e := tdb.Next(cursor)
		if k == nil || e != nil {
			break
		}

		ti:=&BeetleLicenseLog{}
		ti.TxStr = k.String()
		ti.ClientTranstionItem = v

		bnll.Logs = append(bnll.Logs,ti)
	}

	w.WriteHeader(200)
	w.Write([]byte(Respponse(0,"",0,bnll)))
}