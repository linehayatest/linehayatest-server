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
)

type message struct {
	Event int `json:"event"`
}

type chatMessage struct {
	message
	Message string `json:"message"`
}

type dashboardStatusUpdate struct {
	message
	volunteers []volunteer.VolunteerStateUpdate `json:"volunteers"`
	students   []student.StudentStateUpdate     `json:"students"`
}

func ChatRequestAcceptedFactory() []byte {
	m := message{
		Event: CHAT_REQUEST_ACCEPTED,
	}
	r, err := json.Marshal(m)
	err2.Check(err)

	return r
}

func ChatMessageFactory(messageContent string) []byte {
	m := chatMessage{
		message: message{
			Event: CHAT_MESSAGE,
		},
		Message: messageContent,
	}
	r, err := json.Marshal(m)
	err2.Check(err)

	return r
}

func PartyHasReconnectFactory() []byte {
	m := message{
		Event: PARTY_HAS_RECONNECT,
	}
	r, err := json.Marshal(m)
	err2.Check(err)

	return r
}

func PartyHasDisconnectFactory() []byte {
	m := message{
		Event: PARTY_HAS_DISCONNECT,
	}
	r, err := json.Marshal(m)
	err2.Check(err)

	return r
}

func PartyHasTimeoutFactory() []byte {
	m := message{
		Event: PARTY_HAS_TIMEOUT,
	}
	r, err := json.Marshal(m)
	err2.Check(err)

	return r
}

func PartyHasEndConversationFactory() []byte {
	m := message{
		Event: PARTY_HAS_END_CONVERSATION,
	}
	r, err := json.Marshal(m)
	err2.Check(err)

	return r
}

func DashboardStatusUpdate(volunteers []volunteer.VolunteerStateUpdate, students []student.StudentStateUpdate) []byte {
	m := dashboardStatusUpdate{
		volunteers: volunteers,
		students:   students,
		message: message{
			Event: DASHBOARD_STATUS_UPDATE,
		},
	}
	r, err := json.Marshal(m)
	err2.Check(err)

	return r
}
