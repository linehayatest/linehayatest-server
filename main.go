package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"
	"wstest/lib/assert"
	"wstest/lib/connection"
	"wstest/lib/events"
	"wstest/lib/handlers"
	"wstest/lib/logger"
	"wstest/lib/response"
	"wstest/lib/student"
	"wstest/lib/volunteer"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/lainio/err2"
)

var volunteers = volunteer.NewVolunteerRepo()
var students = student.NewStudentRepo()
var connections = connection.NewConnections(students, volunteers)

var appLogger *logger.AppLogger

var count = 0

func main() {
	r := gin.Default()

	// setup logging
	var file, err = os.Create("appState.log")
	if err != nil {
		log.Fatalln("Unable to open file for logging")
	}
	defer file.Close()

	appLogger = logger.NewLogger(file, volunteers, students, connections)
	defer appLogger.Logger.Sync()

	h := handlers.NewHandler(students, volunteers, connections, appLogger)

	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "PATCH", "PUT", "POST", "OPTIONS", "HEAD", "DELETE"},
	}))

	r.GET("can_volunteer_login/:email", h.CanLogin)

	r.GET("can_volunteer_reconnect/:email", h.CanVolunteerReconnect)

	r.GET("can_student_reconnect/:userId", h.CanStudentReconnect)

	r.GET("is_student_active_on_another_tab/:userId", h.IsStudentActiveOnAnotherTab)

	r.PUT("student_end_conversation/:userId", h.StudentEndConversation)

	r.GET("/ws", func(c *gin.Context) {
		defer err2.Catch(func(err error) {
			return
		})

		conn, _, _, err := ws.UpgradeHTTP(c.Request, c.Writer)
		err2.Check(err)

		log.Printf("Time: %d:%d:%d\n", time.Now().Hour(), time.Now().Minute(), time.Now().Second())
		log.Printf("\nNEW CONNECTION\n\n\n")

		go func() {
			defer conn.Close()

			for {
				data, _, err := wsutil.ReadClientData(conn)
				if err != nil {
					conn.Close()
					h.HandleSocketReadError(err, conn)
					return
				}

				msg, err := parseMessage(data)
				if err != nil {
					continue
				}

				handleSocketMessage(msg, data, conn)
			}
		}()
	})

	r.Run(":8080") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func handleSocketMessage(msg events.Message, data []byte, conn net.Conn) {
	defer err2.Catch(func(err error) {
		fmt.Printf("Error: %v", err)
	})

	var err error

	switch msg.Type {
	case events.VOLUNTEER_ACCEPT_CHAT_REQUEST:
		payload := new(events.AcceptChatPayload)
		err = json.Unmarshal(data, payload)
		assert.NoError(err, "Failed to handle ACCEPT_CHAT event")
		handleAcceptChatRequest(conn, msg.Metadata.Identity, payload.Payload.UserID)
	case events.END_CONVERSATION:
		switch msg.Metadata.UserType {
		case events.VOLUNTEER_TYPE:
			handleVolunteerEndConversation(conn, msg.Metadata.Identity)
		case events.STUDENT_TYPE:
			id, err := strconv.Atoi(msg.Metadata.Identity)
			assert.NoErrorf(err, "Failed to handle chat message event. Failed to parse user id: %v", err)
			handleStudentEndConversation(conn, id)
		default:
			log.Println("Unrecognized user type")
		}
	case events.SEND_MESSAGE:
		payload := new(events.SendMessagePayload)
		err = json.Unmarshal(data, payload)
		assert.NoError(err, "Failed to handle SEND_MESSAGE event")
		switch msg.Metadata.UserType {
		case events.VOLUNTEER_TYPE:
			handleVolunteerSendMessage(conn, msg.Metadata.Identity, payload.Payload.Message)
		case events.STUDENT_TYPE:
			id, err := strconv.Atoi(msg.Metadata.Identity)
			assert.NoErrorf(err, "Failed to handle chat message event. Failed to parse user id: %v", err)
			handleStudentSendMessage(conn, id, payload.Payload.Message)
		default:
			log.Println("Unrecognized user type")
		}
	case events.VOLUNTEER_LOGIN:
		assert.NoError(err, "Failed to handle VOLUNTEER_LOGIN event")
		handleVolunteerLogin(msg.Metadata.Identity, conn)
	case events.STUDENT_REQUEST_FOR_CHAT:
		handleRequestForChat(conn)
	case events.VOLUNTEER_RECONNECT:
		handleVolunteerReconnect(msg.Metadata.Identity, conn)
	case events.STUDENT_RECONNECT:
		id, err := strconv.Atoi(msg.Metadata.Identity)
		assert.NoError(err, "Failed to handle STUDENT_RECONNECT event. Failed to parse user id")
		handleStudentReconnect(id, conn)
	}
}

func parseMessage(data []byte) (events.Message, error) {
	var msg events.Message
	err := json.Unmarshal(data, &msg)
	return msg, err
}

func handleStudentEndConversation(conn net.Conn, userId int) {
	appLogger.Log(events.END_CONVERSATION, conn)
	connections.SendToConnectedVolunteer(userId, string(response.PartyHasEndConversationFactory()))
	v := connections.GetConnectedVolunteer(userId)
	if v != nil {
		if v.FSM.Current() == volunteer.CHAT_DISCONNECT.State() {
			volunteers.RemoveByConn(v.Conn)
		} else {
			volunteers.EventByConn(v.Conn, volunteer.END_CONVERSATION)
		}
	}
	connections.RemoveConnection(conn)
	students.RemoveByUserID(userId)
	updateStatus()
}

func handleVolunteerEndConversation(conn net.Conn, email string) {
	appLogger.Log(events.END_CONVERSATION, conn)
	connections.SendToConnectedStudent(email, string(response.PartyHasEndConversationFactory()))
	volunteers.EventByConn(conn, volunteer.END_CONVERSATION)
	s := connections.GetConnectedStudent(email)
	if s != nil {
		students.RemoveByUserID(s.UserID)
	}
	connections.RemoveConnection(conn)
	updateStatus()
}

func handleStudentSendMessage(conn net.Conn, userId int, message string) {
	appLogger.Log(events.SEND_MESSAGE, conn)

	v := connections.GetConnectedVolunteer(userId)
	if v == nil {
		return
	}

	if v.FSM.Current() == volunteer.CHAT_DISCONNECT.State() {
		v.UnsentChatMessages = append(v.UnsentChatMessages, response.NewChatMessage(message))
		return
	}

	connections.SendToConnectedVolunteer(userId, string(response.ChatMessageFactory(message)))
}

func handleVolunteerSendMessage(conn net.Conn, email, message string) {
	appLogger.Log(events.SEND_MESSAGE, conn)

	s := connections.GetConnectedStudent(email)
	if s == nil {
		return
	}

	if s.FSM.Current() == student.CHAT_DISCONNECT.State() {
		s.UnsentChatMessages = append(s.UnsentChatMessages, response.NewChatMessage(message))
		return
	}

	connections.SendToConnectedStudent(email, string(response.ChatMessageFactory(message)))
}

func handleAcceptChatRequest(volunteerConn net.Conn, volunteerEmail string, studentUserID int) {
	err := connections.AddConnection(volunteerEmail, studentUserID)
	err2.Check(err)

	volunteers.EventByEmail(volunteerEmail, volunteer.ACCEPT_CHAT_REQUEST)
	students.EventByUserID(studentUserID, student.CHAT_REQUEST_ACCEPTED)

	connections.SendToConnectedStudent(volunteerEmail, string(response.ChatRequestAcceptedFactory()))
	updateStatus()
	appLogger.Log(events.VOLUNTEER_ACCEPT_CHAT_REQUEST, volunteerConn)
}

func handleVolunteerReconnect(email string, conn net.Conn) {
	err := volunteers.SetConnByEmail(email, conn)
	err2.Check(err)

	volunteers.SendUnsentMessagesByEmail(email)

	volunteers.EventByEmail(email, volunteer.RECONNECT)

	s := connections.GetConnectedStudent(email)
	if s == nil || (s != nil && s.FSM.Current() == student.CHAT_DISCONNECT.State()) {
		volunteers.SendMessageByEmail(email, response.PartyHasDisconnectFactory())
	}

	connections.SendToConnected(conn, string(response.PartyHasReconnectFactory()))
	updateStatus()
	appLogger.Log(events.VOLUNTEER_RECONNECT, conn)
}

func handleStudentReconnect(userID int, conn net.Conn) {
	err := students.SetConnByUserID(userID, conn)
	err2.Check(err)

	students.SendUnsentMessagesByUserID(userID)

	students.EventByConn(conn, student.RECONNECT)

	v := connections.GetConnectedVolunteer(userID)
	if v == nil || (v != nil && v.FSM.Current() == volunteer.CHAT_DISCONNECT.State()) {
		students.SendMessageByUserID(userID, response.PartyHasDisconnectFactory())
	}

	connections.SendToConnected(conn, string(response.PartyHasReconnectFactory()))
	updateStatus()
	appLogger.Log(events.STUDENT_RECONNECT, conn)
}

func handleVolunteerLogin(email string, conn net.Conn) {
	volunteer := volunteer.NewVolunteer(conn, email)
	volunteers.Add(volunteer)

	updateStatus()
	appLogger.Log(events.VOLUNTEER_LOGIN, conn)
}

func handleRequestForChat(conn net.Conn) {
	defer err2.Catch(func(err error) {
		log.Printf("ERROR handling student's chat request: %v\n", err)
	})

	s := student.NewStudent(conn)
	students.Add(s)

	err := wsutil.WriteServerMessage(conn, ws.OpText, response.ChatRequestReply(s.UserID))
	assert.NoErrorf(err, "Fail to send student request reply. (UserID: %d)", s.UserID)

	updateStatus()
	appLogger.Log(events.STUDENT_REQUEST_FOR_CHAT, conn)
}

func updateStatus() {
	v := volunteers.PrepareStatusUpdate()
	s := students.PrepareStatusUpdate()
	volunteers.NotifyAll(string(response.DashboardStatusUpdate(v, s)))
}
