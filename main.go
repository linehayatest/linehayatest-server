package main

import (
	"encoding/json"
	"fmt"
	"net"
	"wstest/lib/connection"
	"wstest/lib/events"
	"wstest/lib/response"
	"wstest/lib/student"
	"wstest/lib/volunteer"

	"github.com/gin-gonic/gin"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/lainio/err2"
	"github.com/lainio/err2/assert"
)

var volunteers = volunteer.NewVolunteerRepo()
var students = student.NewStudentRepo()
var connections = connection.NewConnections(students, volunteers)

func main() {
	r := gin.Default()
	r.GET("/login_permission", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/ws", func(c *gin.Context) {
		defer err2.Catch(func(err error) {
			fmt.Printf("%v", err)
			return
		})

		conn, _, _, err := ws.UpgradeHTTP(c.Request, c.Writer)
		err2.Check(err)

		go func() {
			defer conn.Close()

			for {
				data, _, err := wsutil.ReadClientData(conn)
				if err != nil {
					conn.Close()
					handleSocketReadError(err, conn)
					return
				}

				msg, err := parseMessage(data)
				if err != nil {
					continue
				}

				handleSocketMessage(msg, conn)
			}
		}()
	})

	r.Run(":8050") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func handleSocketReadError(err error, conn net.Conn) {
	// refresh timeout
	// disconnect timeout
}

func handleSocketMessage(msg events.Message, conn net.Conn) {
	defer err2.Catch(func(err error) {
		fmt.Printf("Error: %v", err)
	})

	switch eventType := msg.Event; eventType {
	case events.ACCEPT_CHAT_REQUEST:
		payload, ok := msg.Payload.(events.AcceptChatPayload)
		assert.P.True(ok)
		handleAcceptChatRequest(conn, payload.UserID)
	case events.END_CONVERSATION:
		handleEndConversation(conn)
	case events.SEND_MESSAGE:
		payload, ok := msg.Payload.(events.SendMessagePayload)
		assert.P.True(ok)
		handleSendMessage(conn, payload.Message)
	case events.VOLUNTEER_RECONNECT:
		payload, ok := msg.Payload.(events.VolunteerReconnectPayload)
		assert.P.True(ok)
		handleVolunteerReconnect(payload.Email, conn)
	case events.STUDENT_RECONNECT:
		payload, ok := msg.Payload.(events.StudentReconnectPayload)
		assert.P.True(ok)
		handleStudentReconnect(payload.UserID, conn)
	case events.VOLUNTEER_LOGIN:
		payload, ok := msg.Payload.(events.VolunteerLoginPayload)
		assert.P.True(ok)
		handleVolunteerLogin(payload.Email, conn)
	case events.REQUEST_FOR_CHAT:
		handleRequestForChat(conn)
	}
}

func parseMessage(data []byte) (events.Message, error) {
	var msg events.Message
	err := json.Unmarshal(data, &msg)
	return msg, err
}

func handleEndConversation(conn net.Conn) {
	connections.SendToConnected(conn, string(response.PartyHasDisconnectFactory()))
	connections.RemoveConnection(conn)
	connections.EventConnectedVolunteer(conn, volunteer.END_CONVERSATION)
}

func handleSendMessage(conn net.Conn, message string) {
	connections.SendToConnected(conn, string(response.ChatMessageFactory(message)))
}

func handleAcceptChatRequest(volunteerConn net.Conn, userID int) {
	err := connections.AddConnection(volunteerConn, userID)
	err2.Check(err)

	volunteers.EventByConn(volunteerConn, volunteer.ACCEPT_CHAT_REQUEST)
	students.EventByUserID(userID, student.CHAT_REQUEST_ACCEPTED)

	connections.SendToConnected(volunteerConn, string(response.ChatRequestAcceptedFactory()))
	updateStatus()
}

func handleVolunteerReconnect(email string, conn net.Conn) {
	err := volunteers.SetConnByEmail(email, conn)
	err2.Check(err)

	volunteers.EventByConn(conn, volunteer.RECONNECT)

	connections.SendToConnected(conn, string(response.PartyHasReconnectFactory()))
	updateStatus()
}

func handleStudentReconnect(userID int, conn net.Conn) {
	err := students.SetConnByUserID(userID, conn)
	err2.Check(err)

	students.EventByConn(conn, student.RECONNECT)

	connections.SendToConnected(conn, string(response.PartyHasReconnectFactory()))
	updateStatus()
}

func handleVolunteerLogin(email string, conn net.Conn) {
	volunteer := volunteer.NewVolunteer(conn, email)
	volunteers.Add(volunteer)

	updateStatus()
}

func handleRequestForChat(conn net.Conn) {
	s := student.NewStudent(conn)
	students.Add(s)

	updateStatus()
}

func updateStatus() {
	v := volunteers.PrepareStatusUpdate()
	s := students.PrepareStatusUpdate()
	volunteers.NotifyAll(string(response.DashboardStatusUpdate(v, s)))
}
