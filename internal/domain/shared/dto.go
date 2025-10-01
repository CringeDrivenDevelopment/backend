package shared

type ApiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   error  `json:"-"`
}
