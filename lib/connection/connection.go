package connection

import (
	"fmt"
	"log"
	"net"
	"wstest/lib/student"
	"wstest/lib/volunteer"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// Map Student Connection -> Volunteer Connection
type Connections struct {
	connections map[*student.Student]*volunteer.Volunteer
	students    *student.StudentRepo
	volunteers  *volunteer.VolunteerRepo
}

func NewConnections(students *student.StudentRepo, volunteers *volunteer.VolunteerRepo) *Connections {
	return &Connections{
		connections: make(map[*student.Student]*volunteer.Volunteer),
		students:    students,
		volunteers:  volunteers,
	}
}

func (c *Connections) AddConnection(volunteerEmail string, studentUserID int) error {
	v := c.volunteers.GetVolunteerByEmail(volunteerEmail)
	if v == nil {
		return fmt.Errorf("Unable to add new connection, volunteer not found")
	}

	s := c.students.GetStudentByUserID(studentUserID)
	if s == nil {
		return fmt.Errorf("Unable to add new connection, student not found")
	}

	c.connections[s] = v

	return nil
}

func (c *Connections) SendToConnectedVolunteer(studentUserId int, message string) {
	for s, v := range c.connections {
		if s != nil && s.UserID == studentUserId && s.Conn != nil {
			err := wsutil.WriteServerMessage(v.Conn, ws.OpText, []byte(message))
			if err != nil {
				log.Println("ERROR Sending chat message")
			}
			break
		}
	}
}

func (c *Connections) SendToConnectedStudent(volunteerEmail string, message string) {
	for s, v := range c.connections {
		if v != nil && v.Email == volunteerEmail && v.Conn != nil {
			err := wsutil.WriteServerMessage(s.Conn, ws.OpText, []byte(message))
			fmt.Println("IM WRITING TO THIS STUDENT")
			fmt.Println(message)
			if err != nil {
				log.Println("ERROR Sending chat message")
			}
			break
		}
	}
}

func (c *Connections) SendToConnected(conn net.Conn, message string) {
	var socket net.Conn
	for k, v := range c.connections {
		if k.Conn == conn {
			socket = v.Conn
			break
		}

		if v.Conn == conn {
			socket = k.Conn
			break
		}
	}

	if socket != nil {
		err := wsutil.WriteServerMessage(socket, ws.OpText, []byte(message))
		if err != nil {
			log.Println("ERROR Sending Chat Messages")
		}
	}
}

func (c *Connections) RemoveConnection(conn net.Conn) {
	for k, v := range c.connections {
		if k.Conn == conn || v.Conn == conn {
			k.Conn.Close()
			delete(c.connections, k)
		}
	}
}

func (c *Connections) GetConnectedStudent(volunteerEmail string) *student.Student {
	for k, v := range c.connections {
		if v.Email == volunteerEmail {
			return k
		}
	}
	return nil
}

func (c *Connections) GetConnectedVolunteer(userId int) *volunteer.Volunteer {
	for k, v := range c.connections {
		if k.UserID == userId {
			return v
		}
	}
	return nil
}

type ConnectionState struct {
	StudentID      int    `header:"student id"`
	VolunteerEmail string `header:"volunteer email"`
}

func (c Connections) ReadState() []ConnectionState {
	states := make([]ConnectionState, 0)
	for k, v := range c.connections {
		states = append(states, ConnectionState{
			StudentID:      k.UserID,
			VolunteerEmail: v.Email,
		})
	}
	return states
}
