package handlers

type CanLoginResponse struct {
	CanLogin bool `json:"canLogin"`
}

type CanReconnectResponse struct {
	CanReconnect bool `json:"canReconnect"`
}

type IsStudentActiveOnAnotherTabResponse struct {
	IsActive bool `json:"isActive"`
}
