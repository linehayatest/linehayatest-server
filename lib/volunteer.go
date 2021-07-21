package lib

import (
	"github.com/looplab/fsm"
)

type Volunteer struct {
	email string
	fsm   *fsm.FSM
}

func (v Volunteer) Transition() {

}

func NewVolunteer() *Volunteer {
	v := &Volunteer{}

	v.fsm = fsm.NewFSM(
		"free",
		fsm.Events{
			{Name: "accept chat request", Src: []string{"free"}, Dst: "chat-active"},
			{Name: "disconnect", Src: []string{"chat-active"}, Dst: "chat-disconnect"},
			{Name: "end conversation", Src: []string{"chat-active"}, Dst: "free"},
			{Name: "reconnect", Src: []string{"chat-disconnect"}, Dst: "chat-active"},
		},
		fsm.Callbacks{
			"enter_chat-active": func(e *fsm.Event) {
				if e.Src == "free" {
					// tell student chat request is accepted
					// add new connectedTo entry (if not already added)
					// notify all volunteers students' request has been accepted
				}
				// reconnects
				if e.Src == "chat-disconnect" {
					// 
				}
			},
			// disconnect
			"enter_chat-disconnect": func(e *fsm.Event) {
			},
			"enter_free": func(e *fsm.Event) {
				// end conversation
				if e.Src == "chat-active" {
					
				}
			},
		},
	)

	return v
}
