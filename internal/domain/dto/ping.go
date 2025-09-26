package dto

type PingOutput struct {
	Body struct {
		Status string `json:"status" example:"Pong!"`
	}
}
