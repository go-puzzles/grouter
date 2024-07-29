package prouter

import (
	"errors"
	"net/http"
)

type Ret struct {
	Code    int    `json:"code"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Err     error  `json:"-"`
}

func (r *Ret) GetCode() int {
	return r.Code
}

func (r *Ret) GetError() error {
	return r.Err
}

func (r *Ret) GetData() any {
	return r
}

func (r *Ret) GetMessage() string {
	return r.Message
}

func SuccessResponse(data any) *Ret {
	return &Ret{
		Code: http.StatusOK,
		Data: data,
	}
}

func ErrorResponse(code int, message string) *Ret {
	return &Ret{
		Code:    code,
		Message: message,
		Err:     errors.New(message),
	}
}
