package logger

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"
	"wstest/lib/connection"
	"wstest/lib/events"
	"wstest/lib/student"
	"wstest/lib/volunteer"

	"github.com/lensesio/tableprinter"
	"go.uber.org/zap"
)

var eventCodeToString = map[events.Event]string{
	events.VOLUNTEER_ACCEPT_CHAT_REQUEST: "VOLUNTEER_ACCEPT_CHAT_REQUEST",
	events.VOLUNTEER_LOGIN:               "VOLUNTEER_LOGIN",
	events.VOLUNTEER_RECONNECT:           "VOLUNTEER_RECONNECT",
	events.STUDENT_REQUEST_FOR_CHAT:      "STUDENT_REQUEST_FOR_CHAT",
	events.STUDENT_RECONNECT:             "STUDENT_RECONNECT",
	events.SEND_MESSAGE:                  "SEND_MESSAGE",
	events.END_CONVERSATION:              "END_CONVERSATION",
	events.STUDENT_REQUEST_FOR_CALL:      "STUDENT_REQUEST_FOR_CALL",
	events.VOLUNTEER_ACCEPT_CALL:         "VOLUNTEER_ACCEPT_CALL",
	events.END_CALL:                      "END_CALL",
}

type AppLogger struct {
	Logger      *zap.SugaredLogger
	Printer     *tableprinter.Printer
	volunteers  *volunteer.VolunteerRepo
	students    *student.StudentRepo
	connections *connection.Connections
}

func NewLogger(file *os.File, volunteers *volunteer.VolunteerRepo, students *student.StudentRepo, connections *connection.Connections) *AppLogger {
	logger := new(AppLogger)
	logger.Logger = setupZapLogger()

	log.SetOutput(file)
	log.SetFlags(0)
	logger.Printer = tableprinter.New(file)

	logger.volunteers = volunteers
	logger.students = students
	logger.connections = connections

	return logger
}

func setupZapLogger() *zap.SugaredLogger {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{
		"appState.log",
	}
	logger, err := cfg.Build()
	if err != nil {
		log.Fatalf("Failed to setup logger: %v\n", err)
	}
	sugar := logger.Sugar()
	return sugar
}

func (l *AppLogger) Log(ev events.Event, conn net.Conn) {
	studentStates := l.students.ReadState()
	volunteerStates := l.volunteers.ReadState()
	connectionStates := l.connections.ReadState()

	log.Printf("Time: %d:%d:%d\n", time.Now().Hour(), time.Now().Minute(), time.Now().Second())
	log.Printf("EVENT: %s\n", eventCodeToString[ev])
	log.Printf("Event from: %s\n\n", l.GetEventSource(conn))
	log.Println("STUDENTS:")
	l.Printer.Print(studentStates)
	log.Printf("\n\n")
	log.Println("VOLUNTEERS:")
	l.Printer.Print(volunteerStates)
	log.Printf("\n\n")
	log.Println("CONNECTIONS:")
	l.Printer.Print(connectionStates)
	log.Printf("\n========================================================\n\n")
}

func (l *AppLogger) LogStudentCall(userId int) {
	studentStates := l.students.ReadState()
	volunteerStates := l.volunteers.ReadState()
	connectionStates := l.connections.ReadState()

	log.Printf("Time: %d:%d:%d\n", time.Now().Hour(), time.Now().Minute(), time.Now().Second())
	log.Printf("EVENT: Student Request for Call\n")
	log.Printf("Event from: Student (%d)\n\n", userId)
	log.Println("STUDENTS:")
	l.Printer.Print(studentStates)
	log.Printf("\n\n")
	log.Println("VOLUNTEERS:")
	l.Printer.Print(volunteerStates)
	log.Printf("\n\n")
	log.Println("CONNECTIONS:")
	l.Printer.Print(connectionStates)
	log.Printf("\n========================================================\n\n")
}

func (l *AppLogger) LogVolunteerAcceptCall(email string, userId int) {
	studentStates := l.students.ReadState()
	volunteerStates := l.volunteers.ReadState()
	connectionStates := l.connections.ReadState()

	log.Printf("Time: %d:%d:%d\n", time.Now().Hour(), time.Now().Minute(), time.Now().Second())
	log.Printf("EVENT: Volunteer Accept Call\n")
	log.Printf("Event from: Volunteer (%s) - Accepting Student (%d) \n\n", email, userId)
	log.Println("STUDENTS:")
	l.Printer.Print(studentStates)
	log.Printf("\n\n")
	log.Println("VOLUNTEERS:")
	l.Printer.Print(volunteerStates)
	log.Printf("\n\n")
	log.Println("CONNECTIONS:")
	l.Printer.Print(connectionStates)
	log.Printf("\n========================================================\n\n")
}

func (l *AppLogger) LogEndCall(email string) {
	studentStates := l.students.ReadState()
	volunteerStates := l.volunteers.ReadState()
	connectionStates := l.connections.ReadState()

	log.Printf("Time: %d:%d:%d\n", time.Now().Hour(), time.Now().Minute(), time.Now().Second())
	log.Printf("EVENT: Volunteer End Call\n")
	log.Printf("Event from: Volunteer (%s)\n\n", email)
	log.Println("STUDENTS:")
	l.Printer.Print(studentStates)
	log.Printf("\n\n")
	log.Println("VOLUNTEERS:")
	l.Printer.Print(volunteerStates)
	log.Printf("\n\n")
	log.Println("CONNECTIONS:")
	l.Printer.Print(connectionStates)
	log.Printf("\n========================================================\n\n")
}

func (l *AppLogger) LogVolunteerDisconnect(email string, conn net.Conn) {
	studentStates := l.students.ReadState()
	volunteerStates := l.volunteers.ReadState()
	connectionStates := l.connections.ReadState()

	log.Printf("Time: %d:%d:%d\n", time.Now().Hour(), time.Now().Minute(), time.Now().Second())
	log.Printf("EVENT: Volunteer Disconnect\n")
	log.Printf("Event from: Volunteer (%s)\n\n", email)
	log.Println("STUDENTS:")
	l.Printer.Print(studentStates)
	log.Printf("\n\n")
	log.Println("VOLUNTEERS:")
	l.Printer.Print(volunteerStates)
	log.Printf("\n\n")
	log.Println("CONNECTIONS:")
	l.Printer.Print(connectionStates)
	log.Printf("\n========================================================\n\n")
}

func (l *AppLogger) LogStudentDisconnect(userId int, conn net.Conn) {
	studentStates := l.students.ReadState()
	volunteerStates := l.volunteers.ReadState()
	connectionStates := l.connections.ReadState()

	log.Printf("Time: %d:%d:%d\n", time.Now().Hour(), time.Now().Minute(), time.Now().Second())
	log.Printf("EVENT: Student Disconnect\n")
	log.Printf("Event from: Student (%d)\n\n", userId)
	log.Println("STUDENTS:")
	l.Printer.Print(studentStates)
	log.Printf("\n\n")
	log.Println("VOLUNTEERS:")
	l.Printer.Print(volunteerStates)
	log.Printf("\n\n")
	log.Println("CONNECTIONS:")
	l.Printer.Print(connectionStates)
	log.Printf("\n========================================================\n\n")
}

func (l *AppLogger) GetEventSource(conn net.Conn) (output string) {
	if s := l.students.GetStudentByConn(conn); s != nil {
		return fmt.Sprintf("STUDENT: %d [Connection: %v]", s.UserID, s.Conn)
	} else if v := l.volunteers.GetVolunteerByConnection(conn); v != nil {
		return fmt.Sprintf("VOLUNTEER: %s [Connection: %v]", v.Email, v.Conn)
	}
	return fmt.Sprintf("Unknown source: [Connection: %v]", conn)
}
