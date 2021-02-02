package webapi

import (
	"encoding/json"
	"github.com/giantliao/beatles-client-lib/clientwallet"
	"net/http"
)

func AddWalletApi(mux *http.ServeMux)  {
	mux.HandleFunc("/api/wallet/isCreate",walletIsCreate)
	mux.HandleFunc("/api/wallet/isOpen",walletIsOpen)
	mux.HandleFunc("/api/wallet/openWallet",openWallet)
	mux.HandleFunc("/api/wallet/balance",getBalance)
	mux.HandleFunc("/api/wallet/address",getWalletAddress)
}

func walletIsCreate(w http.ResponseWriter, r *http.Request)  {
	if r.Method != "GET"{
		w.WriteHeader(500)
		w.Write([]byte("not a correct method"))
		return
	}

	b := clientwallet.IsWalletCreate()

	status  := 0

	if b{
		status = 1
	}

	w.WriteHeader(200)
	w.Write([]byte(SimpleResponse(status,"",0)))
}

func walletIsOpen(w http.ResponseWriter, r *http.Request)  {
	if r.Method != "GET"{
		w.WriteHeader(500)
		w.Write([]byte("not a correct method"))
		return
	}

	status := 0

	if _,err:=clientwallet.GetWallet();err==nil{
		status = 1
	}

	w.WriteHeader(200)
	w.Write([]byte(SimpleResponse(status,"",0)))
}

type OpenWallet struct {
	Auth string `json:"auth"`
	Typ  int    `json:"typ"`		//0 eth, 1 btl address, 2 trx
}

func openWallet(w http.ResponseWriter, r *http.Request)  {
	if reqBytes,err:=ReadReq(r);err!=nil{
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}else{
		ow:=&OpenWallet{}

		err:=json.Unmarshal(reqBytes,ow)
		if err!=nil{
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		if err:=clientwallet.LoadWallet(ow.Auth);err!=nil{
			w.WriteHeader(200)
			w.Write([]byte(SimpleResponse(1,"password not correct",1)))
			return
		}

		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(0,"success",0)))
		return
	}
}


func getBalance(w http.ResponseWriter, r *http.Request)  {
	if r.Method != "GET"{
		w.WriteHeader(500)
		w.Write([]byte("not a correct method"))
		return
	}

	wi,err:=clientwallet.GetBalance()
	if err!=nil{
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,err.Error(),1)))
		return
	}

	resp:=Respponse(0,"",0,wi)

	w.WriteHeader(200)
	w.Write([]byte(resp))
}

func getWalletAddress(w http.ResponseWriter, r *http.Request)  {
	if r.Method != "GET"{
		w.WriteHeader(500)
		w.Write([]byte("not a correct method"))
		return
	}

	wb,err:=clientwallet.GetWalletInfo()
	if err!=nil{
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,err.Error(),1)))
		return
	}

	resp:=Respponse(0,"",0,wb)

	w.WriteHeader(200)
	w.Write([]byte(resp))
}