package models

type ResultCode int

const (
	ResultCodeOK         ResultCode = 0
	ResultCodeBadRequest ResultCode = 400
)

type Result struct {
	Code ResultCode `json:"code"`
	Msg  string     `json:"msg"`
	Data any        `json:"data"`
}

func OK(data any) *Result {
	return &Result{
		Code: ResultCodeOK,
		Msg:  "success",
		Data: data,
	}
}

func BadRequest(message string) *Result {
	return &Result{
		Code: ResultCodeBadRequest,
		Msg:  message,
		Data: nil,
	}
}
