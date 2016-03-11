package gomc

import (
	"fmt"
	"github.com/gorilla/websocket"
	"time"
	"log"
	"net/http"
)

type WebSocketHub struct {

	Connections map[*WebSocketConnection]bool
	Broadcast chan []byte
	Register chan *WebSocketConnection
	Unregister chan *WebSocketConnection
}

type WebSocketConnection struct {
	Ws *websocket.Conn
	Send chan []byte
}

var WebSocketHubs = make(map[string]WebSocketHub)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)



func (h *WebSocketHub) Run() {
	for {
		select {
		case c := <-h.Register:
			h.Connections[c] = true
		case c := <-h.Unregister:
			if _, ok := h.Connections[c]; ok {
				delete(h.Connections, c)
				close(c.Send)
			}
		case m := <-h.Broadcast:
			for c := range h.Connections {
				select {
				case c.Send <- m:
				default:
					close(c.Send)
					delete(h.Connections, c)
				}
			}
		}
	}
}

func (c *WebSocketConnection) ReadPump(h WebSocketHub) {
	defer func() {
		h.Unregister <- c
		c.Ws.Close()
	}()
	c.Ws.SetReadLimit(maxMessageSize)
	c.Ws.SetReadDeadline(time.Now().Add(pongWait))
	c.Ws.SetPongHandler(func(string) error { c.Ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.Ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		h.Broadcast <- message
	}
}

func (c *WebSocketConnection) Write(mt int, payload []byte) error {
	c.Ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.Ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *WebSocketConnection) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.Write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func WebSocketPush(channel string, message string) error{

	var err error
	messageBytes := []byte(message)
	if(len(WebSocketHubs[channel].Connections) > 0){
		WebSocketHubs[channel].Broadcast <- messageBytes
	}

	return err
}

func WebSocketRegister(channel string, w http.ResponseWriter, r *http.Request) error{

	var err error
	if(len(WebSocketHubs[channel].Connections) == 0){
        var h = WebSocketHub{
            Broadcast:   make(chan []byte),
            Register:    make(chan *WebSocketConnection),
            Unregister:  make(chan *WebSocketConnection),
            Connections: make(map[*WebSocketConnection]bool),
        }
        WebSocketHubs[channel] = h
    }
    
    for hubKey, hubValue := range WebSocketHubs {
        if(hubKey == channel){
            go hubValue.Run()
        }
    }
    ws, err := Upgrader.Upgrade(w, r, nil)
    fmt.Println(err)
    if err != nil {
        log.Println(err)
        return err
    }
    c := &WebSocketConnection{Send: make(chan []byte, 256), Ws: ws}
    
    //hubs["conn_1"].register <- c
    WebSocketHubs[channel].Register <- c
    go c.WritePump()

	return err
}