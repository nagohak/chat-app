package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	ws "github.com/gorilla/websocket"
	"github.com/nagohak/chat-app/auth"
	"github.com/nagohak/chat-app/models"
)

const (
	// Max wait time when writing message to peer
	writeWait = 10 * time.Second

	// Max time till next pong from peer
	pongWait = 60 * time.Second

	// Send ping interval. Must be less then pong wait time
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 10000
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

var (
	newline = []byte{'\n'}
	// space   = []byte{' '}
)

// Client represents the websocket client at the server
type Client struct {
	// The actual websocket connection.
	conn     *ws.Conn
	wsServer *WsServer
	send     chan []byte
	rooms    map[*Room]bool
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
}

func newClient(conn *ws.Conn, wsServer *WsServer, ID string, name string) *Client {
	client := &Client{
		conn:     conn,
		wsServer: wsServer,
		rooms:    make(map[*Room]bool),
		send:     make(chan []byte, 256),
		// ID:       uuid.New(),
		Name: name,
	}

	client.ID, _ = uuid.Parse(ID)
	return client
}

// ServeWs handles websocket requests from clients requests.
func ServeWs(wsServer *WsServer, w http.ResponseWriter, r *http.Request) {

	// name, ok := r.URL.Query()["name"]

	// if !ok || len(name[0]) < 1 {
	// 	log.Println("Url Param 'name' is missing")
	// 	return
	// }

	userCtxValue := r.Context().Value(auth.UserContextKey)
	if userCtxValue == nil {
		log.Println("Not autheticated")
		return
	}

	user := userCtxValue.(models.User)

	upgrader.CheckOrigin = func(r *http.Request) bool { return true } // for test purposes only
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := newClient(conn, wsServer, user.GetID(), user.GetName())

	go client.writePump()
	go client.readPump()

	wsServer.register <- client
}

func (client *Client) GetID() string {
	return client.ID.String()
}

func (client *Client) GetName() string {
	return client.Name
}

func (client *Client) disconnect() {
	client.wsServer.unregister <- client
	for r := range client.rooms {
		r.unregister <- client
	}
	close(client.send)
	client.conn.Close()
}

func (client *Client) handleNewMessage(jsonMessage []byte) {

	var message Message
	if err := json.Unmarshal(jsonMessage, &message); err != nil {
		log.Printf("Error on unmarshal JSON message: %s\n", err)
	}

	message.Sender = client

	switch message.Action {
	case SendMessageAction:
		roomID := message.Target.GetId()
		if room := client.wsServer.findRoomByID(roomID); room != nil {
			room.broadcast <- &message
		}
	case JoinRoomAction:
		client.handleJoinRoomMessage(message)
	case LeaveRoomAction:
		client.handleLeaveRoomMessage(message)
	case JoinRoomPrivateAction:
		client.handleJoinRoomPrivateMessage(message)
	}
}

func (client *Client) handleJoinRoomMessage(message Message) {
	roomName := message.Message

	client.joinRoom(roomName, nil)
}

func (client *Client) handleLeaveRoomMessage(message Message) {
	room := client.wsServer.findRoomByID(message.Message)
	if room == nil {
		return
	}

	delete(client.rooms, room)

	room.unregister <- client
}

func (client *Client) handleJoinRoomPrivateMessage(message Message) {
	target := client.wsServer.FindUserById(message.Message)
	if target == nil {
		return
	}

	roomName := message.Message + client.GetID()

	joinedRoom := client.joinRoom(roomName, target)

	if joinedRoom != nil {
		client.inviteTargetUser(target, joinedRoom)
	}
}

func (client *Client) joinRoom(roomName string, sender models.User) *Room {
	room := client.wsServer.findRoomByName(roomName)
	if room == nil {
		room = client.wsServer.createRoom(roomName, sender != nil)
	}

	if sender == nil && room.Private {
		return nil
	}

	if !client.isInRoom(room) {
		client.rooms[room] = true
		room.register <- client

		client.notifyRoomJoined(room, sender)
	}

	return room
}

func (client *Client) inviteTargetUser(target models.User, room *Room) {
	message := &Message{
		Action:  JoinRoomPrivateAction,
		Message: target.GetID(),
		Target:  room,
		Sender:  client,
	}

	if err := client.wsServer.redis.Publish(ctx, PubSubGeneralChannel, message.encode()).Err(); err != nil {
		log.Println(err)
	}
}

func (client *Client) isInRoom(room *Room) bool {
	if _, ok := client.rooms[room]; ok {
		return true
	}

	return false
}

func (client *Client) notifyRoomJoined(room *Room, sender models.User) {
	message := Message{
		Action: RoomJoinedAction,
		Target: room,
		Sender: sender,
	}

	client.send <- message.encode()
}

func (client *Client) readPump() {
	defer func() {
		client.disconnect()
	}()

	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error { client.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, jsonMessage, err := client.conn.ReadMessage()
		if err != nil {
			if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
				log.Printf("unexpected close error: %v\n", err)
			}
			break

		}

		client.handleNewMessage(jsonMessage)
	}

}

func (client *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				client.conn.WriteMessage(ws.CloseMessage, []byte{})
				return
			}

			w, err := client.conn.NextWriter(ws.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Attach queued chat messages to the current websocker message.
			n := len(client.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-client.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(ws.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
