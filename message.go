package main

import (
	"encoding/json"
	"log"

	"github.com/nagohak/chat-app/models"
)

const SendMessageAction = "send-message"
const JoinRoomAction = "join-room"
const LeaveRoomAction = "leave-room"
const UserJoinedAction = "user-join"
const UserLeftAction = "user-left"
const JoinRoomPrivateAction = "join-room-private"
const RoomJoinedAction = "room-joined"

type Message struct {
	Action  string      `json:"action"`
	Message string      `json:"message"`
	Target  *Room       `json:"target"`
	Sender  models.User `json:"sender"`
}

func (m *Message) UnmarshalJSON(data []byte) error {
	type Alias Message
	msg := &struct {
		Sender Client `json:"sender"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}

	m.Sender = &msg.Sender
	return nil
}

func (m *Message) encode() []byte {
	json, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
	}

	return json
}
