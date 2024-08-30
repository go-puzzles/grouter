package prouter

import (
	"net/http"
	"reflect"
)

var (
	responseTmpl reflect.Type = reflect.TypeOf((*Ret)(nil)).Elem()
)

type Response interface {
	GetCode() int
	GetMessage() string
	GetData() any
}

type ResponseTmpl interface {
	SetCode(int)
	SetData(any)
	SetMessage(string)
}

func SetResponseTmpl(tmpl ResponseTmpl) {
	responseTmpl = reflect.TypeOf(tmpl).Elem()
}

func NewResponseTmpl() ResponseTmpl {
	return reflect.New(responseTmpl).Interface().(ResponseTmpl)
}

type Ret struct {
	Code    int    `json:"code"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

func (r *Ret) SetCode(code int) {
	r.Code = code
}

func (r *Ret) SetMessage(msg string) {
	r.Message = msg
}

func (r *Ret) SetData(data any) {
	r.Data = data
}

func (r *Ret) GetCode() int {
	return r.Code
}

func (r *Ret) GetData() any {
	return r.Data
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
	}
}
