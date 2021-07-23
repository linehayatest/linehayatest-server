package student

import (
	"fmt"
	"net"

	"github.com/looplab/fsm"
)

type Err string

const (
	CONN_SET_FAILURE Err = "Unable to set connection, student with specified user ID not found"
)

type studentState string

const (
	WAIT            studentState = "wait"
	CHAT_ACTIVE     studentState = "chat-active"
	CHAT_DISCONNECT studentState = "chat-disconnect"
)

type StudentState interface {
	State() string
}

func (s studentState) State() string {
	return string(s)
}

type studentEvent string

const (
	CHAT_REQUEST_ACCEPTED studentEvent = "1"
	DISCONNECT            studentEvent = "2"
	REFRESH               studentEvent = "3"
	RECONNECT             studentEvent = "4"
)

func (e studentEvent) Event() string {
	return string(e)
}

type StudentEvent interface {
	Event() string
}

type Student struct {
	userID int
	Conn   net.Conn
	fsm    *fsm.FSM
}

func NewStudent(conn net.Conn) *Student {
	s := &Student{
		userID: idGenerator.getNewID(),
		Conn:   conn,
		fsm:    studentFSMFactory(),
	}
	return s
}

func studentFSMFactory() *fsm.FSM {
	fsmEvents := []fsm.EventDesc{
		{
			Name: CHAT_REQUEST_ACCEPTED.Event(),
			Src:  []string{WAIT.State()},
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
	}

	fsmCallbacks := map[string]fsm.Callback{}

	return fsm.NewFSM(WAIT.State(), fsmEvents, fsmCallbacks)
}

type StudentRepo struct {
	students []*Student
}

func NewStudentRepo() *StudentRepo {
	return &StudentRepo{
		students: make([]*Student, 0),
	}
}

func (s StudentRepo) GetStudentByConn(conn net.Conn) *Student {
	for _, v := range s.students {
		if v.Conn == conn {
			return v
		}
	}
	return nil
}

func (s StudentRepo) GetStudentByUserID(userID int) *Student {
	for _, v := range s.students {
		if v.userID == userID {
			return v
		}
	}

	return nil
}

func (s StudentRepo) SetConnByUserID(userID int, conn net.Conn) error {
	for _, v := range s.students {
		if v.userID == userID {
			v.Conn = conn
			return nil
		}
	}

	return fmt.Errorf(string(CONN_SET_FAILURE))
}

func (s StudentRepo) Add(student *Student) {
	s.students = append(s.students, student)
}

type StudentStateUpdate struct {
	UserID int `json:"userId"`
}

func (s StudentRepo) PrepareStatusUpdate() []StudentStateUpdate {
	states := make([]StudentStateUpdate, 5)
	for _, stud := range s.students {
		states = append(states, StudentStateUpdate{
			UserID: stud.userID,
		})
	}

	return states
}

func (s StudentRepo) EventByConn(conn net.Conn, e StudentEvent) error {
	student := s.GetStudentByConn(conn)
	err := student.fsm.Event(e.Event())
	return err
}

func (s StudentRepo) EventByUserID(id int, e StudentEvent) error {
	student := s.GetStudentByUserID(id)
	err := student.fsm.Event(e.Event())
	return err
}
