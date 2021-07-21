package lib

import (
	"github.com/looplab/fsm"
)

type Student struct {
	userID int
	fsm    *fsm.FSM
}

func (s Student) TransitionWaitToActive() error {
	err := s.fsm.Event("chat-active")
	if err != nil {
		return err
	}
	return nil
}

func (s Student) TransitionDisconnectToActive( /* skt socket */ ) error {
	err := s.fsm.Event("chat-active" /* skt */)
	if err != nil {
		return err
	}
	return nil
}

func (s Student) TransitionActiveToDisconnect() error {
	err := s.fsm.Event("chat-disconnect")
	if err != nil {
		return err
	}
	return nil
}

func NewStudent() *Student {
	s := &Student{}

	// TODO: assign user ID

	s.fsm = fsm.NewFSM(
		"wait",
		fsm.Events{
			{Name: "chat request accepted", Src: []string{"wait"}, Dst: "chat-active"},
			{Name: "disconnect", Src: []string{"chat-active"}, Dst: "chat-disconnect"},
			{Name: "reconnect", Src: []string{"chat-disconnect"}, Dst: "chat-active"},
		},
		fsm.Callbacks{
			"enter_chat-active": func(e *fsm.Event) {
				// volunteer accepts call
				if e.Src == "wait" {
					// 1. tell student chat request is accepted
					// 2. add new ConnectedTo to connectedList
					// 3. Notify all volunteers student's request has been accepted
				}

				// student reconnects
				if e.Src == "chat-disconnect" {
					// 1. tell connectedTo volunteer that student has suddenly reconnect
					// * If volunteer is disconnected, skip
					// 2. Find the same student in Students array and reassign its socket connection to new socket

					// newSkt := e.Args[0].(socket)
					// add(newSkt)
				}
			},
			"enter_chat-disconnect": func(e *fsm.Event) {
				// 1. close socket connection
				// 2. set a refresh timeout
				// after timeout, notify connected volunteer of disconnect
				// 3. set a offline timeout
				// after timeout, notify volunteer student is unreachable
			},
		},
	)

	return s
}
