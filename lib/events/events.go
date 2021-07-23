package events

import (
	"encoding/json"
	"fmt"
)

type event int

const (
	ACCEPT_CHAT_REQUEST event = 1
	END_CONVERSATION          = 2
	VOLUNTEER_RECONNECT       = 3
	VOLUNTEER_LOGIN           = 4
	STUDENT_RECONNECT         = 5
	REQUEST_FOR_CHAT          = 6
	SEND_MESSAGE              = 7
)

type Eventer interface {
	Event() event
}

func (e event) Event() event {
	return e
}

type Message struct {
	Payload Payload `json:"payload"`
	Event   event   `json:"event"`
}

type Payload interface {
	Print()
}

type AcceptChatPayload struct {
	UserID int `json:"userId"`
}

func (p AcceptChatPayload) Print() {
	bytes, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bytes))
}

type VolunteerReconnectPayload struct {
	Email string `json:"email"`
}

func (p VolunteerReconnectPayload) Print() {
	bytes, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bytes))
}

type StudentReconnectPayload struct {
	UserID int `json:"userId"`
}

func (p StudentReconnectPayload) Print() {
	bytes, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bytes))
}

type SendMessagePayload struct {
	Message string `json:"message"`
}

func (p SendMessagePayload) Print() {
	fmt.Println(p.Message)
}

type VolunteerLoginPayload struct {
	Email string `json:"email"`
}

func (p VolunteerLoginPayload) Print() {
	fmt.Println(p.Email)
}
