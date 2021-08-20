package handlers

import (
	"fmt"
	"net"
	"strconv"
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

	s := h.students.GetStudentByConn(conn)

	if s != nil {
		if s.FSM.Current() == student.CHAT_ACTIVE.State() {
			fmt.Println("OMG GG")

			h.connections.SendToConnectedVolunteer(s.UserID, string(response.PartyHasDisconnectFactory()))
			h.students.EventByConn(conn, student.DISCONNECT)
			h.logger.LogStudentDisconnect(s.UserID, conn)
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
			h.connections.SendToConnectedStudent(v.Email, string(response.PartyHasDisconnectFactory()))
			h.volunteers.EventByConn(conn, volunteer.DISCONNECT)
			h.logger.LogVolunteerDisconnect(v.Email, conn)
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
