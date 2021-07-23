package connection

import (
	"fmt"
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

func NewConnections(students *student.StudentRepo, volunteers *volunteer.VolunteerRepo) Connections {
	return Connections{
		connections: make(map[*student.Student]*volunteer.Volunteer),
		students:    students,
		volunteers:  volunteers,
	}
}

func (c Connections) AddConnection(volunteerConn net.Conn, userID int) error {
	v := c.volunteers.GetVolunteerByConnection(volunteerConn)
	if v == nil {
		return fmt.Errorf("Unable to add new connection, volunteer not found")
	}

	s := c.students.GetStudentByUserID(userID)
	if s == nil {
		return fmt.Errorf("Unable to add new connection, student not found")
	}

	c.connections[s] = v

	return nil
}

func (c Connections) SendToConnected(conn net.Conn, message string) {
	var socket net.Conn
	for k, v := range c.connections {
		if k.Conn == conn {
			socket = v.Conn
			break
		}

		if v.Conn == conn {
			socket = v.Conn
			break
		}
	}

	if socket != nil {
		wsutil.WriteServerMessage(socket, ws.OpText, []byte(message))
	}
}

func (c Connections) RemoveConnection(conn net.Conn) {
	for k, v := range c.connections {
		if k.Conn == conn || v.Conn == conn {
			k.Conn.Close()
			delete(c.connections, k)
		}
	}
}

func (c Connections) EventConnectedVolunteer(conn net.Conn, event volunteer.VolunteerEvent) (err error) {
	err = c.volunteers.EventByConn(conn, event)
	return err
}
