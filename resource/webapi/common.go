package webapi

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

type WebRespose struct {
	Status int			`json:"status"`
	ErrMsg string		`json:"err_msg"`
	ErrCode int			`json:"err_code"`
	Data interface{}	`json:"data"`
}

func (wr *WebRespose)EmptySuccessString() string  {
	j,_:=json.Marshal(*wr)

	return string(j)
}

func EmptySuccessString() string {
	wr:=&WebRespose{}

	return wr.EmptySuccessString()
}

func (wr *WebRespose)Response(status int, errMsg string, errCode int, data interface{}) string  {
	wr.Status = status
	wr.ErrMsg = errMsg
	wr.ErrCode = errCode
	wr.Data = data

	j,_:=json.Marshal(*wr)

	return string(j)
}

func Respponse(status int, errMsg string, errCode int, data interface{}) string {
	wr:=&WebRespose{}

	return wr.Response(status,errMsg,errCode,data)
}

func SimpleResponse(status int,errMsg string, errCode int) string {
	wr:=&WebRespose{}

	return wr.Response(status,errMsg,errCode,nil)
}

func ReadReq(r *http.Request) ([]byte,error)  {
	if r.Method != "POST"{
		return nil,errors.New("method is not correct")
	}

	if body,err := ioutil.ReadAll(r.Body);err!=nil{
		return nil, err
	}else{
		return body,nil
	}
}