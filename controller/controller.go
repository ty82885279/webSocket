package controller

import (
	"fmt"
	"net/http"
	"webSocket/socket"
)

func WSHandler(resp http.ResponseWriter, req *http.Request) {
	// 升级连接为websocket
	wsSocket, err := socket.Upgrader.Upgrade(resp, req, nil)
	if err != nil {
		return
	}
	wsSocket.SetCloseHandler(func(code int, text string) error {
		fmt.Println("连接关闭")
		return nil
	})
	wsConn := &socket.WSConnection{
		WSSocket:  wsSocket,
		InChan:    make(chan *socket.WSMessage, 1000),
		OutChan:   make(chan *socket.WSMessage, 1000),
		CloseChan: make(chan struct{}),
		IsClosed:  false,
	}
	// 心跳携程
	go wsConn.ProcLoop()
	// 读消息协程
	go wsConn.WsReadLoop()
	// 写消息协程
	go wsConn.WsWriteLoop()
}
