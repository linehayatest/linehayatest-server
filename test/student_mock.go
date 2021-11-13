package test

import "net"

// mocks student front-end
type StudentMock struct {
	localStorage map[string]string
	conn         net.Conn
}

func (s StudentMock) requestChat() {
	
}

func (s StudentMock) sendMessage() {

}

func (s StudentMock) disconnect() {

}

func (s StudentMock) reactToMessage(message string) {

}
