package response

import (
	"encoding/json"
	"wstest/lib/student"
	"wstest/lib/volunteer"

	"github.com/lainio/err2"
)

const (
	CHAT_REQUEST_ACCEPTED      = 1
	CHAT_MESSAGE               = 2
	PARTY_HAS_RECONNECT        = 3
	PARTY_HAS_DISCONNECT       = 4
	PARTY_HAS_TIMEOUT          = 5
	PARTY_HAS_END_CONVERSATION = 6
	DASHBOARD_STATUS_UPDATE    = 7
	CHAT_REQUEST_REPLY         = 8
)

type message struct {
	Type int `json:"type"`
}

type chatMessage struct {
	Type    int    `json:"type"`
	Message string `json:"message"`
}

type dashboardStatusUpdate struct {
	Type       int                              `json:"type"`
	Volunteers []volunteer.VolunteerStateUpdate `json:"volunteers"`
	Students   []student.StudentStateUpdate     `json:"students"`
}

type chatRequestReply struct {
	Type   int `json:"type"`
	UserId int `json:"userId"`
}

func ChatRequestAcceptedFactory() []byte {
	m := message{
		Type: CHAT_REQUEST_ACCEPTED,
	}
	r, err := json.Marshal(m)
	err2.Check(err)

	return r
}

func ChatMessageFactory(messageContent string) []byte {
	m := chatMessage{
		Type:    CHAT_MESSAGE,
		Message: messageContent,
	}
	r, err := json.Marshal(m)
	err2.Check(err)

	return r
}

func PartyHasReconnectFactory() []byte {
	m := message{
		Type: PARTY_HAS_RECONNECT,
	}
	r, err := json.Marshal(m)
	err2.Check(err)

	return r
}

func PartyHasDisconnectFactory() []byte {
	m := message{
		Type: PARTY_HAS_DISCONNECT,
	}
	r, err := json.Marshal(m)
	err2.Check(err)

	return r
}

func PartyHasTimeoutFactory() []byte {
	m := message{
		Type: PARTY_HAS_TIMEOUT,
	}
	r, err := json.Marshal(m)
	err2.Check(err)

	return r
}

func PartyHasEndConversationFactory() []byte {
	m := message{
		Type: PARTY_HAS_END_CONVERSATION,
	}
	r, err := json.Marshal(m)
	err2.Check(err)

	return r
}

func DashboardStatusUpdate(volunteers []volunteer.VolunteerStateUpdate, students []student.StudentStateUpdate) []byte {
	m := dashboardStatusUpdate{
		Type:       DASHBOARD_STATUS_UPDATE,
		Volunteers: volunteers,
		Students:   students,
	}
	r, err := json.Marshal(m)
	err2.Check(err)

	return r
}

func ChatRequestReply(userId int) []byte {
	m := chatRequestReply{
		Type:   CHAT_REQUEST_REPLY,
		UserId: userId,
	}

	r, err := json.Marshal(m)
	err2.Check(err)
	return r
}
