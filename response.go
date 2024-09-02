package prouter

import (
	"net/http"
	"reflect"
)

var (
	responseTmpl reflect.Type = reflect.TypeOf((*Ret)(nil)).Elem()
)

type Response interface {
	SetCode(int) Response
	SetData(any) Response
	SetMessage(string) Response
	GetCode() int
	GetMessage() string
	GetData() any
}

type ResponseTmpl interface {
	Response
}

func SetResponseTmpl(tmpl ResponseTmpl) {
	responseTmpl = reflect.TypeOf(tmpl).Elem()
}

func NewResponseTmpl() ResponseTmpl {
	return reflect.New(responseTmpl).Interface().(ResponseTmpl)
}

var _ Response = (*Ret)(nil)

type Ret struct {
	Code    int    `json:"code"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

func (r *Ret) SetCode(i int) Response {
	r.Code = i
	return r
}

func (r *Ret) SetData(a any) Response {
	r.Data = a
	return r
}

func (r *Ret) SetMessage(s string) Response {
	r.Message = s
	return r
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

func SuccessResponse(data any) Response {
	ret := NewResponseTmpl()
	ret.SetCode(http.StatusOK).SetData(data)
	return ret
}

func ErrorResponse(code int, message string) Response {
	ret := NewResponseTmpl()
	ret.SetCode(code).SetMessage(message)
	return ret
}
