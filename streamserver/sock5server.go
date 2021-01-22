package streamserver

import (
	"errors"
	"log"
	"net"
	"sync"
)

var (
	streamserver         *StreamServer
	streamserverlock     sync.Mutex
	streamServerFlag     bool
	streamServerFlagLock sync.Mutex
)

func GetStreamServer() *StreamServer {
	//if streamserver == nil {
	//	streamserverlock.Lock()
	//	defer streamserverlock.Unlock()
	//	if streamserver == nil {
	//		streamserver = NewStreamServer(0)
	//	}
	//}
	return streamserver
}

func GetStreamServerWithIdx(idx int,protect func(fd int32) bool, getTarget func(conn net.Conn) (string,error), removeSession func(conn net.Conn)) *StreamServer {
	if streamserver == nil {
		streamserverlock.Lock()
		defer streamserverlock.Unlock()
		if streamserver == nil {
			streamserver = NewStreamServer(idx,protect,getTarget,removeSession)
		}
	}
	return streamserver
}

func DestroyStreamServer() {
	streamserver = nil
}

func isStart() bool {
	streamServerFlagLock.Lock()
	defer streamServerFlagLock.Unlock()

	if streamServerFlag {
		return true
	}

	streamServerFlag = true

	return false

}
func StreamServerIsStart() bool {
	return streamServerFlag
}
func StartStreamServer(idx int,protect func(fd int32) bool, getTarget func(conn net.Conn) (string,error), removeSession func(conn net.Conn)) error {

	if streamServerFlag || isStart() {
		return errors.New("vpn have started")
	}

	log.Println("begin start vpn...")

	GetStreamServerWithIdx(idx,protect,getTarget,removeSession).StartServer()

	return nil
}

func StopStreamserver() {
	//log.Println("1",streamServerFlag)
	if !streamServerFlag {
		log.Println("vpn not start")
		return
	}
	streamServerFlagLock.Lock()
	defer streamServerFlagLock.Unlock()
	if !streamServerFlag {
		return
	}

	//log.Println("begin stop vpn ")
	streamServerFlag = false
	//log.Println("2",streamServerFlag)

	GetStreamServer().StopServer()
	DestroyStreamServer()

}
