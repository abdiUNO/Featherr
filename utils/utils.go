package utils

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   error
}

type ErrorResp struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}
