package streamserver

import (
	"errors"
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

func StartStreamServer(idx int) error {
	defer func() {
		streamServerFlag = false
	}()

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

	return GetStreamServerWithIdx(idx).StartServer()
}

func StopStreamserver() {
	if !streamServerFlag {
		return
	}
	streamServerFlagLock.Lock()
	defer streamServerFlagLock.Unlock()
	if !streamServerFlag {
		return
	}
	streamServerFlag = false

	GetStreamServer().StopServer()

}
