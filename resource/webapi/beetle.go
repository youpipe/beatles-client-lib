package webapi

import (
	"encoding/json"
	"github.com/giantliao/beatles-client-lib/app/cmd"
	"github.com/giantliao/beatles-client-lib/bootstrap"
	"github.com/giantliao/beatles-client-lib/clientwallet"
	"github.com/giantliao/beatles-client-lib/config"
	"github.com/giantliao/beatles-client-lib/db"
	"github.com/giantliao/beatles-client-lib/miners"
	"github.com/giantliao/beatles-client-lib/ping"
	"github.com/giantliao/beatles-client-lib/streamserver"
	"github.com/giantliao/beatles-mac-client/setting"
	prominers "github.com/giantliao/beatles-protocol/miners"
	"github.com/kprc/libeth/account"
	"github.com/kprc/nbsnetwork/tools"
	"time"
	"log"
	"net/http"
)

var beetleSetMode func(mode int)
var beetleGetMode func() int
var stopBeetle func()

func BeetleInject(setmode func(mode int), getmode func() int,stopbeetle func())  {
	beetleSetMode = setmode
	beetleGetMode = getmode
	stopBeetle = stopbeetle
}

func AddBeetleApi(mux *http.ServeMux)  {
	mux.HandleFunc("/api/beetle/beetleIsStart",beetleIsStart)
	mux.HandleFunc("/api/beetle/startVpn",beetleStartVpn)
	mux.HandleFunc("/api/beetle/stopVpn",beetleStopVpn)
	mux.HandleFunc("/api/beetle/listMiners",listMiners)
	mux.HandleFunc("/api/beetle/refreshMiners",reFreshMiners)
	mux.HandleFunc("/api/beetle/pingAllMiners",pingAllMiners)
	mux.HandleFunc("/api/beetle/setMode",setVPNMode)
	mux.HandleFunc("/api/beetle/getMode",getVPNMode)
	mux.HandleFunc("/api/beetle/pingMiner",pingMiner)
	mux.HandleFunc("/api/beetle/version",beetleVersion)
	mux.HandleFunc("/api/beetle/vpnIsStarted",vpnIsStarted)
	mux.HandleFunc("/api/beetle/stop",beetleStop)
	mux.HandleFunc("/api/beetle/refreshBootstrap",refreshBootstrap)

}

func vpnIsStarted(w http.ResponseWriter , r *http.Request)  {
	if r.Method != "GET"{
		w.WriteHeader(500)
		w.Write([]byte("not a correct method"))
		return
	}

	var status int

	b:=streamserver.StreamServerIsStart()
	if !b{
		status = 0
	}

	w.WriteHeader(200)
	w.Write([]byte(SimpleResponse(status,"",0)))
}

func beetleIsStart(w http.ResponseWriter , r *http.Request)  {
	if r.Method != "GET"{
		w.WriteHeader(500)
		w.Write([]byte("not a correct method"))
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(EmptySuccessString()))
}

type StartBeetleVPN struct {
	Miner string  	`json:"miner"`
}

func beetleStartVpn(w http.ResponseWriter , r *http.Request)  {

	pbytes,err:=ReadReq(r)
	if err!=nil{
		w.WriteHeader(500)
		w.Write([]byte("read parameter error"))
		return
	}

	if _,err:=clientwallet.GetWallet();err!=nil{
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,"wallet not opened",1)))
		return
	}

	cfg:=config.GetCBtlc()
	if len(cfg.Miners) == 0{
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,"no miner to connect",2)))
		return
	}

	ldb:=db.GetClientLicenseDb()
	cli:=ldb.FindNewestLicense()
	if cli == nil{
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,"not found license",3)))
		return
	}

	now:=tools.GetNowMsTime()

	if cli.License.Content.ExpireTime < now{
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,"license is expired",4)))
		return
	}

	sbv:=&StartBeetleVPN{}

	err=json.Unmarshal(pbytes,sbv)
	if err!=nil{
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,"unmarshal json error",5)))
		return
	}

	if streamserver.StreamServerIsStart(){
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,"vpn have started",6)))
		return
	}


	find := false
	minerIdx := 0

	currentMiner:=account.BeatleAddress(sbv.Miner)
	if currentMiner == ""{
		currentMiner = cfg.CurrentMiner
	}

	for i := 0; i < len(cfg.Miners); i++ {
		if currentMiner == cfg.Miners[i].MinerId {
			find = true
			minerIdx = i
			break
		}
	}
	if !find {
		currentMiner = cfg.Miners[0].MinerId
	}

	cfg.CurrentMiner = currentMiner

	go streamserver.StartStreamServer(minerIdx,nil,streamserver.Handshake,nil)

	setting.SetProxy(cfg.VPNMode)

	cfg.Save()

	log.Println("start vpn success")

	w.WriteHeader(200)
	w.Write([]byte(Respponse(0,"",0,cfg.Miners[minerIdx])))
}

func beetleStopVpn(w http.ResponseWriter , r *http.Request)  {
	if r.Method != "POST"{
		w.WriteHeader(500)
		w.Write([]byte("not a correct method"))
		return
	}

	if !streamserver.StreamServerIsStart() {
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(0,"",0)))
		return
	}


	streamserver.StopStreamserver()
	//pacserver.StopWebDaemon()

	setting.ClearProxy()

	w.WriteHeader(200)
	w.Write([]byte(SimpleResponse(0,"",0)))

}

type AllBeetleMiners struct {
	CurrentMiner string				`json:"current_miner"`
	Miners []*prominers.Miner		`json:"miners"`
}

func listMiners(w http.ResponseWriter, r *http.Request)  {
	if r.Method != "GET"{
		w.WriteHeader(500)
		w.Write([]byte("not a correct method"))
		return
	}

	cfg:=config.GetCBtlc()

	am:=&AllBeetleMiners{
		CurrentMiner: cfg.CurrentMiner.String(),
		Miners: cfg.Miners,
	}

	w.WriteHeader(200)
	w.Write([]byte(Respponse(0,"",0,am)))
}

func beetleStop(w http.ResponseWriter, r *http.Request)  {
	if r.Method != "POST"{
		w.WriteHeader(500)
		w.Write([]byte("not a correct method"))
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(SimpleResponse(0,"",0)))

	time.Sleep(time.Second)

	stopBeetle()
}

func refreshBootstrap(w http.ResponseWriter, r *http.Request)  {
	if r.Method != "POST"{
		w.WriteHeader(500)
		w.Write([]byte("not a correct method"))
		return
	}

	err := bootstrap.UpdateBootstrap()
	if err != nil {
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,err.Error(),1)))
		return
	}


	w.WriteHeader(200)
	w.Write([]byte(SimpleResponse(0,"",0)))

}

func reFreshMiners(w http.ResponseWriter, r *http.Request)  {
	if r.Method != "POST"{
		w.WriteHeader(500)
		w.Write([]byte("not a correct method"))
		return
	}

	flushMachine := miners.NewClientMiners()
	if flushMachine == nil {
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,"may be you are no license",1)))
		return
	}

	if err := flushMachine.FlushMiners(); err != nil {
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,err.Error(),2)))
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(SimpleResponse(0,"",0)))

}

type BeetlePingMiner struct {
	Miner string `json:"miner"`
}

type BeetlePingResult struct {
	Miner *prominers.Miner `json:"miner"`
	TimeInterval int `json:"time_interval"`
}

type BeetlePingAllResult struct {
	AllResult []*BeetlePingResult `json:"all_result"`
}

func pingAllMiners(w http.ResponseWriter, r *http.Request)  {
	if r.Method != "POST"{
		w.WriteHeader(500)
		w.Write([]byte("not a correct method"))
		return
	}

	cfg:=config.GetCBtlc()
	for i:=0;i<len(cfg.Miners);i++{
		tv,err:=ping.Ping(cfg.Miners[i].Ipv4Addr,cfg.Miners[i].Port)
		if err!=nil{
			continue
		}
		config.AddPingTestResult(cfg.Miners[i].MinerId,tv)
	}

	allr := &BeetlePingAllResult{}

	for i:=0;i<len(cfg.Miners);i++{
		pr:=&BeetlePingResult{
			Miner: cfg.Miners[i],
		}

		tv,err:=config.GetPingTestResult(cfg.Miners[i].MinerId)
		if err != nil{
			tv = -1
		}

		pr.TimeInterval = int(tv)

		allr.AllResult = append(allr.AllResult, pr)
	}



	w.WriteHeader(200)
	w.Write([]byte(Respponse(0,"",0,allr)))
}


func pingMiner(w http.ResponseWriter, r *http.Request)  {

	pbytes,err:=ReadReq(r)
	if err!=nil{
		w.WriteHeader(500)
		w.Write([]byte("error: read failed"))
		return
	}

	pm := &BeetlePingMiner{}
	err = json.Unmarshal(pbytes,pm)
	if err!=nil{
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,"unmarshal json error",1)))
		return
	}

	cfg:=config.GetCBtlc()

	idx:=-1

	for i:=0;i<len(cfg.Miners);i++{
		if cfg.Miners[i].MinerId == account.BeatleAddress(pm.Miner){
			idx = i
			break
		}
	}

	if idx < 0{
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,"miner not found",2)))
		return
	}
	var tv int64
	tv,err = ping.Ping(cfg.Miners[idx].Ipv4Addr,cfg.Miners[idx].Port)
	if err!=nil{
		log.Println(err.Error())
		tv = -1
	}else{
		config.AddPingTestResult(account.BeatleAddress(pm.Miner),tv)
	}

	rv:=&BeetlePingResult{
		Miner: cfg.Miners[idx],
		TimeInterval: int(tv),
	}

	w.WriteHeader(200)
	w.Write([]byte(Respponse(0,"",0,rv)))
}

type VPNMode struct {
	Mode int `json:"mode"`
}

func setVPNMode(w http.ResponseWriter, r *http.Request)  {

	pbytes,err:=ReadReq(r)
	if err!=nil{
		w.WriteHeader(500)
		w.Write([]byte("error: read failed"))
		return
	}

	vm:=&VPNMode{}

	err = json.Unmarshal(pbytes,vm)
	if err!=nil{
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,"unmarshal json error",1)))
		return
	}
	if vm.Mode !=0 && vm.Mode != 1{
		w.WriteHeader(200)
		w.Write([]byte(SimpleResponse(1,"mode must 0 or 1",2)))
		return
	}
	beetleSetMode(vm.Mode)

	w.WriteHeader(200)
	w.Write([]byte(SimpleResponse(0,"",0)))
}

func getVPNMode(w http.ResponseWriter, r *http.Request)  {
	if r.Method != "GET"{
		w.WriteHeader(500)
		w.Write([]byte("not a correct method"))
		return
	}

	vm:=&VPNMode{
		Mode: beetleGetMode(),
	}

	w.WriteHeader(200)
	w.Write([]byte(Respponse(0,"",0,vm)))

}

type BeetleVersion struct {
	Version   string	`json:"version"`
	Build     string		`json:"build"`
	BuildTime string	`json:"build_time"`
}

func beetleVersion(w http.ResponseWriter, r *http.Request)  {
	if r.Method != "GET"{
		w.WriteHeader(500)
		w.Write([]byte("not a correct method"))
		return
	}

	bv:=&BeetleVersion{
		Version: cmd.CmdVersion,
		Build: cmd.CmdBuild,
		BuildTime: cmd.CmdBuildTime,
	}


	w.WriteHeader(200)
	w.Write([]byte(Respponse(0,"",0,bv)))

}

