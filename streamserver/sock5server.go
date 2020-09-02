package streamserver

import (
	"errors"
	"log"
	"sync"
)

var (
	streamserver         *StreamServer
	streamserverlock     sync.Mutex
	streamServerFlag     bool
	streamServerFlagLock sync.Mutex
)

func GetStreamServer() *StreamServer {
	if streamserver == nil {
		streamserverlock.Lock()
		defer streamserverlock.Unlock()
		if streamserver == nil {
			streamserver = NewStreamServer(0)
		}
	}
	return streamserver
}

func GetStreamServerWithIdx(idx int) *StreamServer {
	if streamserver == nil {
		streamserverlock.Lock()
		defer streamserverlock.Unlock()
		if streamserver == nil {
			streamserver = NewStreamServer(idx)
		}
	}
	return streamserver
}

func DestroyStreamServer() {
	streamserver = nil
}

func StartStreamServer(idx int) error {

	if !streamServerFlag {
		streamServerFlagLock.Lock()

		if !streamServerFlag {
			streamServerFlag = true
		} else {
			streamServerFlagLock.Unlock()
			return errors.New("vpn have started")
		}

		streamServerFlagLock.Unlock()
	} else {
		return errors.New("vpn have started")
	}

	GetStreamServerWithIdx(idx).StartServer()

	streamServerFlag = false

	return nil
}

func StopStreamserver() {
	if !streamServerFlag {
		log.Println("vpn not start")
		return
	}
	streamServerFlagLock.Lock()
	defer streamServerFlagLock.Unlock()
	if !streamServerFlag {
		return
	}
	streamServerFlag = false

	GetStreamServer().StopServer()
	DestroyStreamServer()

}
