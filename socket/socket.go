package socket

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

var Upgrader = websocket.Upgrader{
	// 允许所有CORS跨域请求
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 消息，这里根据自己需求添加属性，比如：接收者，发送者
type WSMessage struct {
	messageType int
	data        []byte
}

// 客户端连接
type WSConnection struct {
	WSSocket  *websocket.Conn // 底层websocket
	InChan    chan *WSMessage // 读队列
	OutChan   chan *WSMessage // 写队列
	Mutex     sync.Mutex      // 避免重复关闭channel
	IsClosed  bool
	CloseChan chan struct{} // 关闭通知的channel
}

func (wsConn *WSConnection) WsReadLoop() {
	for {
		// 读一个message
		msgType, data, err := wsConn.WSSocket.ReadMessage()
		if err != nil {
			goto closed
		}
		req := &WSMessage{
			msgType,
			data,
		}
		// 放入请求队列
		select {
		case wsConn.InChan <- req:
			fmt.Print("读取消息")
		case <-wsConn.CloseChan:
			goto closed
		}
	}
closed:
	wsConn.wsClose()

}

func (wsConn *WSConnection) WsWriteLoop() {
	for {
		select {
		// 取一个应答
		case msg := <-wsConn.OutChan:
			// 写给websocket
			if err := wsConn.WSSocket.WriteMessage(msg.messageType, msg.data); err != nil {
				goto closed
			}
		case <-wsConn.CloseChan:
			goto closed
		}
	}
closed:
	wsConn.wsClose()
}

func (wsConn *WSConnection) ProcLoop() {
	// 启动一个gouroutine发送心跳
	go func() {
		for {
			time.Sleep(10 * time.Second)
			if err := wsConn.wsWrite(websocket.TextMessage, []byte("heartbeat from server")); err != nil {
				fmt.Println("heartbeat fail")
				wsConn.wsClose()
				break
			}
		}
	}()
}
func (wsConn *WSConnection) wsWrite(messageType int, data []byte) error {
	select {
	case wsConn.OutChan <- &WSMessage{messageType, data}:
	case <-wsConn.CloseChan:
		return errors.New("websocket closed")
	}
	return nil
}

func (wsConn *WSConnection) wsRead() (*WSMessage, error) {
	select {
	case msg := <-wsConn.InChan:
		return msg, nil
	case <-wsConn.CloseChan:
	}
	return nil, errors.New("websocket closed")
}

func (wsConn *WSConnection) wsClose() {
	wsConn.WSSocket.Close()
	wsConn.Mutex.Lock()
	defer wsConn.Mutex.Unlock()
	if !wsConn.IsClosed {
		wsConn.IsClosed = true
		close(wsConn.CloseChan)
	}
}
