package volunteer

import (
	"fmt"
	"net"

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
	FREE            volunteerState = "wait"
	CHAT_ACTIVE     volunteerState = "chat-active"
	CHAT_DISCONNECT volunteerState = "chat-disconnect"
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
)

type VolunteerEvent interface {
	Event() string
}

func (v volunteerEvent) Event() string {
	return string(v)
}

type Volunteer struct {
	email string
	Conn  net.Conn
	fsm   *fsm.FSM
}

func NewVolunteer(conn net.Conn, email string) *Volunteer {
	s := &Volunteer{
		email: email,
		Conn:  conn,
		fsm:   volunteerFSMFactory(),
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

func (vs VolunteerRepo) SetConnByEmail(email string, conn net.Conn) error {
	for _, v := range vs.volunteers {
		if v.email == email {
			v.Conn = conn
			return nil
		}
	}

	return fmt.Errorf(string(CONN_SET_FAILURE))
}

func (vs VolunteerRepo) GetVolunteerByConnection(conn net.Conn) *Volunteer {
	for _, v := range vs.volunteers {
		if v.Conn == conn {
			return v
		}
	}
	return nil
}

func (vs VolunteerRepo) NotifyAll(message string) {
	for _, v := range vs.volunteers {
		wsutil.WriteServerMessage(v.Conn, ws.OpText, []byte(message))
	}
}

func (vs VolunteerRepo) Add(v *Volunteer) {
	vs.volunteers = append(vs.volunteers, v)
}

type VolunteerStateUpdate struct {
	Email string `json:"email"`
	State string `json:"state"`
}

func (vs VolunteerRepo) PrepareStatusUpdate() []VolunteerStateUpdate {
	states := make([]VolunteerStateUpdate, 5)
	for _, v := range vs.volunteers {
		states = append(states, VolunteerStateUpdate{
			Email: v.email,
			State: v.fsm.Current(),
		})
	}
	return states
}

func (v VolunteerRepo) EventByConn(conn net.Conn, e VolunteerEvent) error {
	volunteer := v.GetVolunteerByConnection(conn)
	err := volunteer.fsm.Event(e.Event())
	return err
}
