package main

import (
	"encoding/json"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/nagohak/chat-app/models"
)

const PubSubGeneralChannel = "general"

type WsServer struct {
	clients        map[*Client]bool
	register       chan *Client
	unregister     chan *Client
	broadcast      chan []byte
	rooms          map[*Room]bool
	users          []models.User
	roomRepository models.RoomRepository
	userRepository models.UserRepository
	redis          *redis.Client
}

func NewWsServer(roomRepository models.RoomRepository, userRepository models.UserRepository, redis *redis.Client) *WsServer {
	s := &WsServer{
		clients:        make(map[*Client]bool),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		broadcast:      make(chan []byte),
		rooms:          make(map[*Room]bool),
		roomRepository: roomRepository,
		userRepository: userRepository,
		redis:          redis,
	}

	users, err := userRepository.GetAllUsers()
	if err != nil {
		log.Fatalln(err)
	}
	s.users = users

	return s
}

func (server *WsServer) Run() {
	go server.listPubSubChannel()

	for {
		select {
		case client := <-server.register:
			server.registerClient(client)
		case client := <-server.unregister:
			server.unregisterClient(client)
		case message := <-server.broadcast:
			server.broadcastToClients(message)
		}
	}
}

func (server *WsServer) publishClientJoined(client *Client) {
	message := &Message{
		Action: UserJoinedAction,
		Sender: client,
	}

	if err := server.redis.Publish(ctx, PubSubGeneralChannel, message.encode()).Err(); err != nil {
		log.Println(err)
	}
}

func (server *WsServer) publicClientLeft(client *Client) {
	message := &Message{
		Action: UserLeftAction,
		Sender: client,
	}

	if err := server.redis.Publish(ctx, PubSubGeneralChannel, message.encode()).Err(); err != nil {
		log.Println(err)
	}
}

func (server *WsServer) listPubSubChannel() {
	pubsub := server.redis.Subscribe(ctx, PubSubGeneralChannel)
	ch := pubsub.Channel()

	var message Message
	for msg := range ch {
		if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
			log.Printf("Error on unmarshal JSON message %s\n", err)
			return
		}
	}

	switch message.Action {
	case UserJoinedAction:
		server.handleUserJoined(message)
	case UserLeftAction:
		server.handleUserLeft(message)
	case JoinRoomPrivateAction:
		server.handleUserJoinPrivate(message)
	}
}

func (server *WsServer) handleUserJoined(message Message) {
	server.users = append(server.users, message.Sender)
	server.broadcastToClients(message.encode())
}

func (server *WsServer) handleUserLeft(message Message) {
	for i, user := range server.users {
		if user.GetID() == message.Sender.GetID() {
			server.users[i] = server.users[len(server.users)-1]
			server.users = server.users[:len(server.users)-1]
			break // only remove the first occurence
		}
	}
	server.broadcastToClients(message.encode())
}

func (server *WsServer) handleUserJoinPrivate(message Message) {
	// Find client for given user, if found add the user to the room.
	// Expect multiple clients for one user now.
	targetClients := server.findClientsByID(message.Message)
	for _, client := range targetClients {
		client.joinRoom(message.Target.GetName(), message.Sender)
	}
}

func (server *WsServer) FindUserById(ID string) models.User {
	var found models.User
	for _, user := range server.users {
		if user.GetID() == ID {
			found = user
			break
		}

	}

	return found
}

func (server *WsServer) findRoomByID(ID string) *Room {
	var found *Room
	for room := range server.rooms {
		if room.GetId() == ID {
			found = room
			break
		}
	}

	return found
}

// func (server *WsServer) findClientByID(ID string) *Client {
// 	var found *Client
// 	for client := range server.clients {
// 		if client.ID.String() == ID {
// 			found = client
// 			break
// 		}
// 	}

// 	return found
// }

func (server *WsServer) findClientsByID(ID string) []*Client {
	var found []*Client
	for client := range server.clients {
		if client.ID.String() == ID {
			found = append(found, client)
		}
	}

	return found
}

func (server *WsServer) findRoomByName(name string) *Room {
	var found *Room
	for room := range server.rooms {
		if room.GetName() == name {
			found = room
			break
		}
	}

	if found == nil {
		found = server.runRoomFromRepository(name)
	}

	return found
}

func (server *WsServer) runRoomFromRepository(name string) *Room {
	var r *Room

	dbRoom, err := server.roomRepository.FindRoomByName(name)
	if err != nil {
		log.Println(err)
		return nil
	}
	if dbRoom != nil {
		r = NewRoom(dbRoom.GetName(), dbRoom.GetPrivate(), server.redis)
		r.ID, _ = uuid.Parse(dbRoom.GetId())

		go r.RunRoom()
		server.rooms[r] = true
	}

	return r
}

func (server *WsServer) createRoom(name string, private bool) *Room {
	r := NewRoom(name, private, server.redis)

	err := server.roomRepository.AddRoom(r)
	if err != nil {
		log.Println(err)
	}

	go r.RunRoom()
	server.rooms[r] = true

	return r
}

func (server *WsServer) registerClient(client *Client) {
	if user := server.FindUserById(client.GetID()); user == nil {
		err := server.userRepository.AddUser(client)
		if err != nil {
			log.Println(err)
		}
	}

	server.publishClientJoined(client)

	server.listOnlineClients(client)
	server.clients[client] = true
}

func (server *WsServer) unregisterClient(client *Client) {
	delete(server.clients, client)

	// err := server.userRepository.RemoveUser(client)
	// if err != nil {
	// 	log.Println(err)
	// }
	server.publicClientLeft(client)
}

func (server *WsServer) broadcastToClients(message []byte) {
	for client := range server.clients {
		client.send <- message
	}
}

// func (server *WsServer) notifiyClientJoined(client *Client) {
// 	message := &Message{
// 		Action: UserJoinedAction,
// 		Sender: client,
// 	}

// 	server.broadcastToClients(message.encode())
// }

// func (server *WsServer) notifiyClientLeft(client *Client) {
// 	message := &Message{
// 		Action: UserLeftAction,
// 		Sender: client,
// 	}

// 	server.broadcastToClients(message.encode())
// }

func (server *WsServer) listOnlineClients(client *Client) {

	// Find unique users istead pf returning all users
	var uniqueUsers = make(map[string]bool)

	for _, user := range server.users {
		if ok := uniqueUsers[user.GetID()]; !ok {
			message := &Message{
				Action: UserJoinedAction,
				Sender: user,
			}
			uniqueUsers[user.GetID()] = true
			client.send <- message.encode()
		}
	}
}
