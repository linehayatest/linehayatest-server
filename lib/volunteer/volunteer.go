package volunteer

import (
	"fmt"
	"net"

	"wstest/lib/response"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/looplab/fsm"
)

type Err string

const (
	CONN_SET_FAILURE Err = "Unable to set connection, volunteer with specified email not found"
)

type volunteerState string

const (
	FREE            volunteerState = "free"
	CHAT_ACTIVE     volunteerState = "chat-active"
	CHAT_DISCONNECT volunteerState = "chat-disconnect"
	CALL_ACTIVE     volunteerState = "call-active"
)

type VolunteerState interface {
	State() string
}

func (v volunteerState) State() string {
	return string(v)
}

type volunteerEvent string

const (
	ACCEPT_CHAT_REQUEST volunteerEvent = "1"
	DISCONNECT          volunteerEvent = "2"
	REFRESH             volunteerEvent = "3"
	RECONNECT           volunteerEvent = "4"
	END_CONVERSATION    volunteerEvent = "5"
	ACCEPT_CALL_REQUEST volunteerEvent = "6"
	HANG_UP             volunteerEvent = "7"
)

type VolunteerEvent interface {
	Event() string
}

func (v volunteerEvent) Event() string {
	return string(v)
}

type Volunteer struct {
	Email              string
	Conn               net.Conn
	FSM                *fsm.FSM
	UnsentChatMessages []response.ChatMessage
}

func NewVolunteer(conn net.Conn, email string) *Volunteer {
	s := &Volunteer{
		Email:              email,
		Conn:               conn,
		FSM:                volunteerFSMFactory(),
		UnsentChatMessages: make([]response.ChatMessage, 0),
	}
	return s
}

func volunteerFSMFactory() *fsm.FSM {
	fsmEvents := []fsm.EventDesc{
		{
			Name: ACCEPT_CHAT_REQUEST.Event(),
			Src:  []string{FREE.State()},
			Dst:  CHAT_ACTIVE.State(),
		},
		{
			Name: DISCONNECT.Event(),
			Src:  []string{CHAT_ACTIVE.State()},
			Dst:  CHAT_DISCONNECT.State(),
		},
		{
			Name: RECONNECT.Event(),
			Src:  []string{CHAT_DISCONNECT.State()},
			Dst:  CHAT_ACTIVE.State(),
		},
		{
			Name: END_CONVERSATION.Event(),
			Src:  []string{CHAT_ACTIVE.State(), CHAT_DISCONNECT.State(), FREE.State()},
			Dst:  FREE.State(),
		},
		{
			Name: ACCEPT_CALL_REQUEST.Event(),
			Src:  []string{FREE.State()},
			Dst:  CALL_ACTIVE.State(),
		},
		{
			Name: HANG_UP.Event(),
			Src:  []string{CALL_ACTIVE.State(), CHAT_ACTIVE.State(), CHAT_DISCONNECT.State()},
			Dst:  FREE.State(),
		},
	}

	fsmCallbacks := map[string]fsm.Callback{}

	return fsm.NewFSM(FREE.State(), fsmEvents, fsmCallbacks)
}

type VolunteerRepo struct {
	volunteers []*Volunteer
}

func NewVolunteerRepo() *VolunteerRepo {
	return &VolunteerRepo{
		volunteers: make([]*Volunteer, 0),
	}
}

func (vs *VolunteerRepo) SetConnByEmail(email string, conn net.Conn) error {
	for _, v := range vs.volunteers {
		if v.Email == email {
			v.Conn = conn
			return nil
		}
	}

	return fmt.Errorf(string(CONN_SET_FAILURE))
}

func (vs *VolunteerRepo) GetVolunteerByConnection(conn net.Conn) *Volunteer {
	for _, v := range vs.volunteers {
		if v.Conn == conn {
			return v
		}
	}
	return nil
}

func (vs *VolunteerRepo) GetVolunteerByEmail(email string) *Volunteer {
	for _, v := range vs.volunteers {
		if v.Email == email {
			return v
		}
	}
	return nil
}

func (vs *VolunteerRepo) NotifyAll(message string) {
	for _, v := range vs.volunteers {
		err := wsutil.WriteServerMessage(v.Conn, ws.OpText, []byte(message))
		if err != nil {
			fmt.Printf("ERROR writing this connection: %v (u.email: %s)", v.Conn, v.Email)
		}
	}
}

func (s *VolunteerRepo) SendUnsentMessagesByEmail(email string) {
	for _, v := range s.volunteers {
		if v.Email == email {
			for _, msg := range v.UnsentChatMessages {
				wsutil.WriteServerMessage(v.Conn, ws.OpText, response.SerializeChatMessage(msg))
			}
		}
	}
}

func (vs *VolunteerRepo) Add(v *Volunteer) {
	vs.volunteers = append(vs.volunteers, v)
}

func (vs *VolunteerRepo) PrepareStatusUpdate() []response.VolunteerStateUpdate {
	states := make([]response.VolunteerStateUpdate, 0)
	for _, v := range vs.volunteers {
		states = append(states, response.VolunteerStateUpdate{
			Email: v.Email,
			State: v.FSM.Current(),
		})
	}
	return states
}

func (v *VolunteerRepo) EventByEmail(email string, e VolunteerEvent) error {
	volunteer := v.GetVolunteerByEmail(email)
	err := volunteer.FSM.Event(e.Event())
	return err
}

func (v *VolunteerRepo) EventByConn(conn net.Conn, e VolunteerEvent) error {
	volunteer := v.GetVolunteerByConnection(conn)
	err := volunteer.FSM.Event(e.Event())
	return err
}

type VolunteerLog struct {
	Email string   `header:"Email"`
	State string   `header:"State"`
	Conn  net.Conn `header:"Socket Ptr Add"`
}

func (s *VolunteerRepo) ReadState() []VolunteerLog {
	logs := make([]VolunteerLog, 0)
	for _, v := range s.volunteers {
		logs = append(logs, VolunteerLog{
			Email: v.Email,
			State: v.FSM.Current(),
			Conn:  v.Conn,
		})
	}
	return logs
}

func (s *VolunteerRepo) RemoveByConn(conn net.Conn) {
	for i, _ := range s.volunteers {
		if s.volunteers[i].Conn == conn {
			s.volunteers[i] = s.volunteers[len(s.volunteers)-1]
			s.volunteers[len(s.volunteers)-1] = nil
			s.volunteers = s.volunteers[:len(s.volunteers)-1]
			break
		}
	}
}

func (s *VolunteerRepo) ExistVolunteerWithEmail(email string) bool {
	for _, v := range s.volunteers {
		if v.Email == email {
			return true
		}
	}
	return false
}

func (s *VolunteerRepo) SendMessageByEmail(email string, payload []byte) {
	for _, v := range s.volunteers {
		if v.Email == email {
			wsutil.WriteServerMessage(v.Conn, ws.OpText, payload)
		}
	}
}
