package test

/*
    I've commented out this entire test file, since it looks like we're missing whatever was in wstest/models
    and wstest/types, which makes running these tests impossible
*/

//
//import (
//	"context"
//	"encoding/json"
//	"fmt"
//	"net"
//	"testing"
//	"wstest/models"
//	"wstest/types"
//
//	"github.com/gobwas/ws"
//	"github.com/gobwas/ws/wsutil"
//	"github.com/stretchr/testify/assert"
//)
//
///*
//MILESTONE 1: Student able to send request for chat with volunteers
//MILESTONE 2: Student able to request for chat and call with volunteers
//MILESTONE 3: Volunteer can start a session / close a session
//
//1. Setup 5 clients (2 volunteers, 3 students)
//
//Test following user stories:
//
//* When: Volunteers connect to the server, server will respond with list of online volunteers (and their status) & students requesting for call
//* So that: Volunteers can respond to any student's awaiting call
//
//* When: Student can call into the server, student will be assigned a name and volunteers will be notified
//* So that: Volunteer can see who is calling
//
//* When: Volunteer respond to a call, the student is notified that their call is received
//* So that: Student can know who respond to their call and that they can start sending messages
//
//* When: Volunteer respond to a call, all other volunteers will be notified
//* So that: Volunteers no longer have to respond to the call
//
//* When: Volunteer respond to a student call that has already been responded to, they will receive error message
//* So that: A student is not assigned to 2 volunteers at a time
//
//* When: Volunteer sends a chat message to a student, the student will receive the chat message
//* So that: A student can see the volunteer's chat message
//
//* When: Student sends a chat message to a volunteer, the volunteer will receive the chat message
//* So that: A volunteer can see the student's chat message
//
//* When: Volunteer or Student end the session, the other end will be notified (Volunteer's status will change to Free)
//* So that: Volunteer can proceed to handle the next student
//*/
//
//type User struct {
//	id   int
//	conn net.Conn
//}
//
//type Volunteer struct {
//	User
//	name string
//}
//
//func (v Volunteer) SignIn() error {
//	req := map[string]string{
//		"name": v.name,
//		"type": types.FlagFromVolunteerSignIn,
//	}
//	body, err := json.Marshal(req)
//	if err != nil {
//		return err
//	}
//	err = wsutil.WriteClientMessage(v.conn, ws.OpText, body)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func (v Volunteer) Disconnect() error {
//	return v.conn.Close()
//}
//
//func TestSum(t *testing.T) {
//	// Arrange
//	go func() {
//		runServer()
//	}()
//
//	/*
//		var student1 = User{
//			conn: setupConnection(),
//		}
//
//		var student2 = User{
//			conn: setupConnection(),
//		}
//	*/
//	var volunteer1 = Volunteer{
//		User: User{conn: setupConnection()},
//		name: "VOLUNTEER 1",
//	}
//	var volunteer2 = Volunteer{
//		User: User{conn: setupConnection()},
//		name: "VOLUNTEER 2",
//	}
//	/*
//		var volunteer3 = Volunteer{
//			User: User{conn: setupConnection()},
//			name: "VOLUNTEER 3",
//		}
//	*/
//
//	// Act
//	// 1. volunteer sign in
//	//	* all volunteers should receive notifications on updated active volunteers
//	// 2. student call
//	//	* all volunteers should receive notifications on updated list of awaiting students
//	// 3. volunteer respond to call
//	// 	* the responded student should receive message of volunteer
//	//	* all volunteers should receive notifications on updated list of awaiting students and busy volunteers
//	// 4. volunteer sends chat message
//	//
//	// 5. student sends chat message
//
//	// * Volunteer 1 sign in
//	// * Volunteer 2 sign in
//	// 	* Volunteer 1, 2 should receive message of Volunteer 1 and 2
//
//	// * Volunteer 1 sign in, and should receive response of blank awaiting students/volunteers
//
//	err := volunteer1.SignIn()
//	assert.NoError(t, err)
//
//	// Assert
//	msg, _, err := wsutil.ReadServerData(volunteer1.conn)
//	var resp types.ToAllVolunteersStudentAndVolunteerStatusUpdate
//	err = json.Unmarshal(msg, &resp)
//	assert.NoError(t, err)
//	assert.Equal(t, types.FlagToAllVolunteersStudentAndVolunteerStatusUpdate, resp.MessageType)
//	assert.Len(t, resp.AwaitingStudents, 0)
//	assert.Len(t, resp.OnlineVolunteers, 1)
//	assert.Equal(t, models.VolunteerStatusFree, resp.OnlineVolunteers[0].Status)
//	assert.Equal(t, volunteer1.name, resp.OnlineVolunteers[0].Name)
//	assert.Equal(t, 1, resp.OnlineVolunteers[0].UserID)
//
//	err = volunteer2.SignIn()
//	assert.NoError(t, err)
//
//	// * Volunteer 2 sign in
//	// 	* Volunteer 1, 2 should receive message of Volunteer 1 and 2
//	// Assert
//	msg, _, err = wsutil.ReadServerData(volunteer1.conn)
//	var resp2 types.ToAllVolunteersVolunteerUpdate
//	err = json.Unmarshal(msg, &resp2)
//	assert.NoError(t, err)
//	assert.Equal(t, types.FlagToAllVolunteersVolunteerUpdate, resp2.MessageType)
//	assert.Len(t, resp2.OnlineVolunteers, 2)
//
//	msg, _, err = wsutil.ReadServerData(volunteer2.conn)
//	var resp3 types.ToAllVolunteersStudentAndVolunteerStatusUpdate
//	err = json.Unmarshal(msg, &resp3)
//	assert.NoError(t, err)
//	assert.Equal(t, types.FlagToAllVolunteersStudentAndVolunteerStatusUpdate, resp3.MessageType)
//	assert.Len(t, resp3.AwaitingStudents, 0)
//	assert.Len(t, resp3.OnlineVolunteers, 2)
//	assert.Equal(t, models.VolunteerStatusFree, resp3.OnlineVolunteers[0].Status)
//	assert.Equal(t, models.VolunteerStatusFree, resp3.OnlineVolunteers[1].Status)
//
//	// * Volunteer 2 goes offline
//	// 	* Volunteer 1 should receive  message
//	// Act
//	err = volunteer1.Disconnect()
//	assert.NoError(t, err)
//
//	// Assert
//	msg, _, err = wsutil.ReadServerData(volunteer2.conn)
//	var resp4 types.ToAllVolunteersStudentAndVolunteerStatusUpdate
//	err = json.Unmarshal(msg, &resp4)
//	assert.NoError(t, err)
//	assert.Equal(t, types.FlagToAllVolunteersStudentAndVolunteerStatusUpdate, resp4.MessageType)
//	assert.Len(t, resp4.AwaitingStudents, 0)
//	assert.Len(t, resp4.OnlineVolunteers, 1)
//	assert.Equal(t, models.VolunteerStatusFree, resp4.OnlineVolunteers[0].Status)
//
//}
//
///*
//Table testing model
//
//{
//	Action: func action {}
//	Student1: {
//		structToUnmarshalTo
//		expectedStruct
//	}
//	Student2: nil (if nil, avoid ReadServerData and assert)
//	Volunteer1:
//	Volunteer2:
//	Volunteer3
//
//	* nil means no need to read server data
//	* will read server data
//	compare struct
//	but how to know which struct to unmarhsal to? we pass the struct to unmarshal to:
//	then we assert the result
//}
//*/
//
//func setupConnection() net.Conn {
//	conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), "ws://localhost:8050")
//	if err != nil {
//		fmt.Printf("TEST: Unable to setup connection: %v", err)
//	}
//	return conn
//}
