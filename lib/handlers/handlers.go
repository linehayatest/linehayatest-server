package handlers

import (
	"net"
	"strconv"
	"time"
	"wstest/lib/connection"
	"wstest/lib/logger"
	"wstest/lib/response"
	"wstest/lib/student"
	"wstest/lib/volunteer"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	volunteers  *volunteer.VolunteerRepo
	students    *student.StudentRepo
	connections *connection.Connections
	logger      *logger.AppLogger
}

func NewHandler(s *student.StudentRepo, v *volunteer.VolunteerRepo, c *connection.Connections, l *logger.AppLogger) *Handler {
	return &Handler{v, s, c, l}
}

func (h *Handler) CanLogin(c *gin.Context) {
	email := c.Params.ByName("email")

	v := h.volunteers.ExistVolunteerWithEmail(email)
	c.JSON(200, CanLoginResponse{!v})
}

func (h *Handler) IsStudentActiveOnAnotherTab(c *gin.Context) {
	userId, err := strconv.Atoi(c.Params.ByName("userId"))
	if err != nil {
		c.JSON(200, CanReconnectResponse{false})
	}
	s := h.students.GetStudentByUserID(userId)
	isActive := s != nil &&
		(s.FSM.Current() == student.WAIT.State() ||
			s.FSM.Current() == student.CHAT_ACTIVE.State())
	c.JSON(200, IsStudentActiveOnAnotherTabResponse{isActive})
}

func (h *Handler) StudentEndConversation(c *gin.Context) {
	userId, err := strconv.Atoi(c.Params.ByName("userId"))
	if err != nil {
		c.JSON(200, gin.H{})
	}
	h.connections.SendToConnectedVolunteer(userId, string(response.PartyHasEndConversationFactory()))
	v := h.connections.GetConnectedVolunteer(userId)
	if v != nil {
		if v.FSM.Current() == volunteer.CHAT_DISCONNECT.State() {
			h.volunteers.RemoveByConn(v.Conn)
		} else {
			h.volunteers.EventByConn(v.Conn, volunteer.END_CONVERSATION)
		}
	}
	h.connections.RemoveConnectionByStudentID(userId)
	h.students.RemoveByUserID(userId)
}

func (h *Handler) CanVolunteerReconnect(c *gin.Context) {
	email := c.Params.ByName("email")
	v := h.volunteers.GetVolunteerByEmail(email)
	isChatting := v != nil && (v.FSM.Current() == volunteer.CHAT_DISCONNECT.State() ||
		v.FSM.Current() == volunteer.CHAT_ACTIVE.State())
	c.JSON(200, CanReconnectResponse{isChatting})
}

func (h *Handler) CanStudentReconnect(c *gin.Context) {
	userId, err := strconv.Atoi(c.Params.ByName("userId"))
	if err != nil {
		c.JSON(200, CanReconnectResponse{false})
	}
	s := h.students.GetStudentByUserID(userId)
	isChatting := s != nil &&
		(s.FSM.Current() == student.CHAT_DISCONNECT.State() ||
			s.FSM.Current() == student.CHAT_ACTIVE.State())
	c.JSON(200, CanReconnectResponse{isChatting})
}

func (h *Handler) HandleSocketReadError(err error, conn net.Conn) {
	// refresh timeout
	// disconnect timeout

	conn.Close()
	s := h.students.GetStudentByConn(conn)

	if s != nil {
		if s.FSM.Current() == student.CHAT_ACTIVE.State() {
			// TODO: timeout
			h.connections.SendToConnectedVolunteer(s.UserID, string(response.PartyHasDisconnectFactory()))
			h.students.EventByConn(conn, student.DISCONNECT)
			h.logger.LogStudentDisconnect(s.UserID, conn)

			// timeout if student still in disconnect state
			// Question: how do you handle the case where student
			// disconnects, reconnects on 20th seconds, then disconnect on 39nth second?
			// need to record the disconnect timestamp on student
			// which updates on every disconnect event
			time.AfterFunc(40, func() {

			})
		}
		if s.FSM.Current() == student.WAIT.State() {
			h.students.RemoveByConn(conn)
			h.updateStatus()
			h.logger.LogStudentDisconnect(s.UserID, conn)
		}
		return
	}

	v := h.volunteers.GetVolunteerByConnection(conn)
	if v != nil {
		if v.FSM.Current() == volunteer.CHAT_ACTIVE.State() {
			// TODO: timeout
			h.connections.SendToConnectedStudent(v.Email, string(response.PartyHasDisconnectFactory()))
			h.volunteers.EventByConn(conn, volunteer.DISCONNECT)
			h.logger.LogVolunteerDisconnect(v.Email, conn)

			// timeout if volunteer still in disconnect state
			time.AfterFunc(40, func() {

			})
		}
		if v.FSM.Current() == volunteer.FREE.State() {
			h.volunteers.RemoveByConn(conn)
			h.updateStatus()
			h.logger.LogVolunteerDisconnect(v.Email, conn)
		}

	}

	// if student is chat-active
	// inform the connected user
	// set time out
	// connections.InformByConn

	// if student is wait
	// remove from students

	// if volunteer is chat-active
	// inform connected student
	// set timeout

	// if volunteer is free
	// remove from volunteers

	// check
	// log
}

func (h *Handler) updateStatus() {
	v := h.volunteers.PrepareStatusUpdate()
	s := h.students.PrepareStatusUpdate()
	h.volunteers.NotifyAll(string(response.DashboardStatusUpdate(v, s)))
}
