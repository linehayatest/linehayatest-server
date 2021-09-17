package student

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
	CONN_SET_FAILURE Err = "Unable to set connection, student with specified user ID not found"
)

type studentState string

const (
	WAIT            studentState = "wait"
	WAIT_CALL       studentState = "wait-call"
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
	CALL                  studentEvent = "5"
)

func (e studentEvent) Event() string {
	return string(e)
}

type StudentEvent interface {
	Event() string
}

type Student struct {
	UserID             int
	PeerID             string
	Conn               net.Conn
	FSM                *fsm.FSM
	UnsentChatMessages []response.ChatMessage
}

func NewStudent(conn net.Conn) *Student {
	s := &Student{
		UserID:             idGenerator.getNewID(),
		Conn:               conn,
		FSM:                studentFSMFactory(),
		UnsentChatMessages: make([]response.ChatMessage, 0),
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
		{
			Name: CALL.Event(),
			Src:  []string{WAIT.State()},
			Dst:  WAIT_CALL.State(),
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

func (s *StudentRepo) GetStudentByConn(conn net.Conn) *Student {
	for _, v := range s.students {
		if v.Conn == conn {
			return v
		}
	}
	return nil
}

func (s *StudentRepo) GetStudentByUserID(userID int) *Student {
	for _, v := range s.students {
		if v != nil && v.UserID == userID {
			return v
		}
	}

	return nil
}

func (s *StudentRepo) SetConnByUserID(userID int, conn net.Conn) error {
	for _, v := range s.students {
		if v.UserID == userID {
			v.Conn = conn
			return nil
		}
	}

	return fmt.Errorf(string(CONN_SET_FAILURE))
}

func (s *StudentRepo) SendUnsentMessagesByUserID(userID int) {
	for _, v := range s.students {
		if v.UserID == userID {
			for _, msg := range v.UnsentChatMessages {
				wsutil.WriteServerMessage(v.Conn, ws.OpText, response.SerializeChatMessage(msg))
			}
		}
	}
}

func (s *StudentRepo) Add(student *Student) {
	s.students = append(s.students, student)
}

func (s *StudentRepo) PrepareStatusUpdate() []response.StudentStateUpdate {
	states := make([]response.StudentStateUpdate, 0)
	for _, stud := range s.students {
		if stud.FSM.Current() == "wait" || stud.FSM.Current() == "wait-call" {
			states = append(states, response.StudentStateUpdate{
				UserID: stud.UserID,
				State:  stud.FSM.Current(),
			})
		}

	}
	return states
}

func (s *StudentRepo) EventByConn(conn net.Conn, e StudentEvent) error {
	student := s.GetStudentByConn(conn)
	err := student.FSM.Event(e.Event())
	return err
}

func (s *StudentRepo) EventByUserID(id int, e StudentEvent) error {
	student := s.GetStudentByUserID(id)
	err := student.FSM.Event(e.Event())
	return err
}

type StudentLog struct {
	UserID int      `header:"UserID"`
	State  string   `header:"State"`
	Conn   net.Conn `header:"Socket Ptr Add"`
}

func (s *StudentRepo) ReadState() []StudentLog {
	logs := make([]StudentLog, 0)
	for _, v := range s.students {
		logs = append(logs, StudentLog{
			UserID: v.UserID,
			State:  v.FSM.Current(),
			Conn:   v.Conn,
		})
	}
	return logs
}

func (s *StudentRepo) RemoveByConn(conn net.Conn) {
	for i, _ := range s.students {
		if s.students[i].Conn == conn {
			s.students[i] = s.students[len(s.students)-1]
			s.students[len(s.students)-1] = nil
			s.students = s.students[:len(s.students)-1]
		}
	}
}

func (s *StudentRepo) RemoveByUserID(userId int) {
	for i, _ := range s.students {
		if s.students[i].UserID == userId {
			fmt.Println("DEBUG 1")
			// copy last element to this element, and remove the last element
			s.students[i] = s.students[len(s.students)-1]
			fmt.Println("DEBUG 2")
			s.students[len(s.students)-1] = nil
			fmt.Println("DEBUG 3")
			s.students = s.students[:len(s.students)-1]
			fmt.Println("DEBUG 4")
		}
	}
}

func (s *StudentRepo) SendMessageByUserID(userID int, payload []byte) {
	for _, v := range s.students {
		if v.UserID == userID {
			wsutil.WriteServerMessage(v.Conn, ws.OpText, payload)
		}
	}
}
