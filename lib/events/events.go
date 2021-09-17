package events

type Event int

const (
	VOLUNTEER_ACCEPT_CHAT_REQUEST Event = 1
	VOLUNTEER_LOGIN                     = 2
	VOLUNTEER_RECONNECT                 = 3
	STUDENT_REQUEST_FOR_CHAT            = 4
	STUDENT_RECONNECT                   = 5
	SEND_MESSAGE                        = 6
	END_CONVERSATION                    = 7
	STUDENT_REQUEST_FOR_CALL            = 8
	VOLUNTEER_ACCEPT_CALL               = 9
	END_CALL                            = 10
)

const VOLUNTEER_TYPE = "volunteer"
const STUDENT_TYPE = "student"

type Eventer interface {
	Event() Event
}

func (e Event) Event() Event {
	return e
}

type Message struct {
	Type Event `json:"type"`
	// used to identify user
	Metadata struct {
		// either "volunteer" or "student"
		UserType string `json:"type"`
		// if "student", should be userId (else "email")
		Identity string `json:"identity"`
	} `json:"metadata"`
}

type AcceptChatPayload struct {
	Payload struct {
		UserID int `json:"userId"`
	} `json:"payload"`
}

type VolunteerReconnectPayload struct {
	Payload struct {
		Email string `json:"email"`
	} `json:"payload"`
}

type SendMessagePayload struct {
	Payload struct {
		Message string `json:"message"`
	} `json:"payload"`
}

type RequestCallPayload struct {
	Payload struct {
		PeerID string `json:"peerId"`
	} `json:"payload"`
}

type AcceptCallPayload AcceptChatPayload

type EndCallPayload VolunteerReconnectPayload
