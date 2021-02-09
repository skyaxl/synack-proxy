package formatters

import (
	"encoding/json"
	"net/http"
)

//Json json formmater
type Json struct {
	res http.ResponseWriter
	req *http.Request
}

//NewJson
func NewJson(res http.ResponseWriter, req *http.Request) *Json {
	return &Json{res, req}
}

//WriteError write error
func (j Json) WriteError(status int, err error) {
	bts, _ := json.Marshal(&struct{ message string }{err.Error()})
	j.res.Write(bts)
	j.res.WriteHeader(status)
}

//Write test
func (j Json) Write(status int, o interface{}) {
	bts, _ := json.Marshal(o)
	j.res.Write(bts)
	j.res.WriteHeader(status)
}
