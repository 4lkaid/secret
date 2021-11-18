package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"golang.org/x/net/websocket"
)

// PORT 默认端口
const PORT = "9527"

// ClientConn 客户端链接
type ClientConn struct {
	websocket *websocket.Conn
	clientIP  string
}

//go:embed index.html
var view embed.FS

// ActiveClients 活动客户端
var ActiveClients = make(map[ClientConn]int)

func send(activeClients map[ClientConn]int, sendMessage string) {
	var err error
	for acs := range activeClients {
		if err = websocket.Message.Send(acs.websocket, sendMessage); err != nil {
			log.Println("消息发送失败", acs.clientIP, err.Error())
		}
	}
}

func chat(ws *websocket.Conn) {
	var err error
	var clientMessage string
	var message string
	client := ws.Request().RemoteAddr
	sockCli := ClientConn{ws, client}
	ActiveClients[sockCli] = 0
	count := len(ActiveClients)
	message = client + " 进入聊天室, 当前在线人数 " + strconv.Itoa(count)
	send(ActiveClients, message)
	log.Println(message)
	for {
		if err = websocket.Message.Receive(ws, &clientMessage); err != nil {
			delete(ActiveClients, sockCli)
			count = len(ActiveClients)
			message = client + " 离开了聊天室, 当前在线人数 " + strconv.Itoa(count)
			send(ActiveClients, message)
			log.Println(message)
			return
		}
		if clientMessage != "" {
			clientMessage = sockCli.clientIP + " Said: " + clientMessage
			send(ActiveClients, clientMessage)
		}
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFS(view, "index.html")
	t.Execute(w, nil)
}

func main() {
	http.HandleFunc("/", index)
	http.Handle("/websocket", websocket.Handler(chat))
	if err := http.ListenAndServe(":"+PORT, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
