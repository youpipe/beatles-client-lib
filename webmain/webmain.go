package webmain

import (
	"fmt"
	"github.com/giantliao/beatles-client-lib/bootstrap"
	"github.com/giantliao/beatles-client-lib/config"
	"github.com/giantliao/beatles-client-lib/resource/pacserver"
	"github.com/giantliao/beatles-client-lib/resource/webapi"
	"github.com/giantliao/beatles-client-lib/streamserver"
	"log"
	"mime"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)



var builtinTypesLower = map[string]string{
	".css":  "text/css; charset=utf-8",
	".gif":  "image/gif",
	".htm":  "text/html; charset=utf-8",
	".html": "text/html; charset=utf-8",
	".jpeg": "image/jpeg",
	".jpg":  "image/jpeg",
	".js":   "text/javascript; charset=utf-8",
	".json": "application/json",
	".mjs":  "text/javascript; charset=utf-8",
	".pdf":  "application/pdf",
	".png":  "image/png",
	".svg":  "image/svg+xml",
	".wasm": "application/wasm",
	".webp": "image/webp",
	".xml":  "text/xml; charset=utf-8",
}

type Proxy interface {
	SetProxy(int)
	ClearProxy()
}

var stop chan os.Signal

func init()  {
	stop = make(chan os.Signal,1)
}


var currentPorxy Proxy

func setMode(mode int)  {
	currentPorxy.SetProxy(mode)
	cfg:=config.GetCBtlc()
	if cfg.VPNMode != mode{
		cfg.VPNMode = mode
		cfg.Save()
	}

}

func clearProxy()  {
	currentPorxy.ClearProxy()
}


func getMode() int  {
	cfg:=config.GetCBtlc()
	return cfg.VPNMode
}


func dealExtensionType() {
	for k, v := range builtinTypesLower {
		if err := mime.AddExtensionType(k, v); err != nil {
			log.Println("add mime error", err.Error())
			continue
		}
	}
}

func OpenBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

}

func StartWEBService(proxy Proxy)  {

	currentPorxy = proxy

	dealExtensionType()


	cfg:=config.GetCBtlc()
	if len(cfg.Miners) == 0 {
		err := bootstrap.UpdateBootstrap()
		if err != nil {
			log.Println(err.Error())
			return
		}

		if len(cfg.Miners) == 0{
			panic("no miners")
		}

	}

	go pacserver.StartWebDaemon(
		webapi.AddBeetleApi,
		webapi.AddLicenseApi,
		webapi.AddWalletApi)

	webapi.BeetleInject(clearProxy,setMode,getMode,sig2Beetle)

	OpenBrowser(cfg.GetVPNPage())

	go bootstrap.StartTimer()


	//defer close(stop)

	signal.Notify(stop,
		syscall.SIGKILL,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	s:=<-stop

	fmt.Println("get signal:",s)
	log.Println("get signal:",s)

	StopAll()

}

var flag = false
var stopflagLock sync.Mutex

func sig2Beetle()  {
	stop <- syscall.SIGQUIT
}

func StopAll()  {
	if flag {
		return
	}

	if !flag {
		stopflagLock.Lock()
		defer stopflagLock.Unlock()

		if flag {
			return
		}

		flag = true
	}

	config.GetCBtlc().Save()

	log.Println("begin to clear proxy setting")
	//setting.ClearProxy()
	currentPorxy.ClearProxy()

	bootstrap.StopTimer()

	time.Sleep(time.Second * 2)

	log.Println("begin to stop pac server")
	pacserver.StopWebDaemon()
	log.Println("begin to stop stream server")
	streamserver.StopStreamserver()
}