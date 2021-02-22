package main

import (
	"net/http"
	"webSocket/controller"
)

func main() {

	http.HandleFunc("/ws", controller.WSHandler)
	http.ListenAndServe("127.0.0.1:7777", nil)

}
