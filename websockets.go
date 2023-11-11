package sigeaserver2

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
)

type WsMsg map[string]interface{}

type websocketConn struct {
	ws           *websocket.Conn
	receiverChan chan WsMsg
	sendChan     chan WsMsg
	handler      WebSocketHandler
}

func (s *websocketConn) writeLoop() {
	for {
		msg := <-s.sendChan
		data, _ := json.Marshal(msg)
		s.ws.WriteMessage(websocket.TextMessage, []byte(data))
	}
}

func (s *websocketConn) readLoop() {
	for {
		_, data, err := s.ws.ReadMessage()

		if err != nil {
			s.ws.Close()
			break
		}

		var m interface{}
		json.Unmarshal(data, &m)
		//log.Printf("Lida mensagem web socket [%v][%v]\n", string(data), m)
		var msg WsMsg
		msg = m.(map[string]interface{})
		s.receiverChan <- msg
	}
}

type WebSocketHandler interface {
	WebSocketConnect(receiveChan chan WsMsg, sendChan chan WsMsg)
}

type WebSocketInfo struct {
	name    string
	handler WebSocketHandler
}

var upgrader = &websocket.Upgrader{}

func (i *WebSocketInfo) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	//fmt.Printf("Fazendo upgrade para ws [%v]\n", err)
	if err != nil {
		return
	}
	defer ws.Close()

	receiveChan := make(chan WsMsg)
	sendChan := make(chan WsMsg)

	wsConn := websocketConn{ws, receiveChan, sendChan, i.handler}

	go wsConn.writeLoop()
	go i.handler.WebSocketConnect(receiveChan, sendChan)
	wsConn.readLoop()
}

func (s *SigeaServer) AddWebSocket(caminho string, handler WebSocketHandler) {
	s.webSocketInfos = append(s.webSocketInfos, WebSocketInfo{caminho, handler})
}
