package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/nagohak/chat-app/config"
)

type Room struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Private    bool      `json:"private"`
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
	redis      *config.RedisClient
}

const welcomeMessage = "%s joined the room"
const leavedMessage = "%s leaved the room"

var ctx = context.Background()

func NewRoom(name string, private bool, redis *config.RedisClient) *Room {
	return &Room{
		ID:         uuid.New(),
		Name:       name,
		Private:    private,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message),
		redis:      redis,
	}
}

func (r *Room) RunRoom() {
	go r.subscribeToRoomMessages()

	for {
		select {
		case client := <-r.register:
			r.registerClientInRoom(client)
		case client := <-r.unregister:
			r.unregisterClientInRoom(client)
		case message := <-r.broadcast:
			r.publishRoomMessage(message.encode())
		}
	}
}

func (r *Room) registerClientInRoom(client *Client) {
	// send welcome message first then new user won't see his own message
	if !r.Private {
		r.notifyClientJoinedRoom(client)
	}
	r.clients[client] = true
}

func (r *Room) unregisterClientInRoom(client *Client) {
	delete(r.clients, client)
	r.notifyClientLeavedRoom(client)
}

func (r *Room) broadcastToClientsInRoom(message []byte) {
	for client := range r.clients {
		client.send <- message
	}
}

func (r *Room) notifyClientJoinedRoom(client *Client) {
	message := &Message{
		Action:  SendMessageAction,
		Target:  r,
		Message: fmt.Sprintf(welcomeMessage, client.GetName()),
	}

	r.publishRoomMessage(message.encode())
}

func (r *Room) notifyClientLeavedRoom(client *Client) {
	message := &Message{
		Action:  SendMessageAction,
		Target:  r,
		Message: fmt.Sprintf(leavedMessage, client.GetName()),
	}

	r.publishRoomMessage(message.encode())
}

func (r *Room) publishRoomMessage(message []byte) {
	err := r.redis.Publish(ctx, r.GetName(), message).Err()

	if err != nil {
		log.Println(err)
	}
}

func (r *Room) subscribeToRoomMessages() {
	pubsub := r.redis.Subscribe(ctx, r.GetName())

	ch := pubsub.Channel()

	for msg := range ch {
		r.broadcastToClientsInRoom([]byte(msg.Payload))
	}
}

func (r *Room) GetId() string {
	return r.ID.String()
}

func (r *Room) GetName() string {
	return r.Name
}

func (r *Room) GetPrivate() bool {
	return r.Private
}
