package factory

import (
	"net/http"

	"github.com/skyaxl/synack-proxy/pkg/response/formatters"
)

//Default factory
type Default struct {
}

//NewDefault new default factory
func NewDefault() *Default {
	return &Default{}
}

//ResponseFormatter format a response for http request
type ResponseFormatter interface {
	WriteError(status int, err error)
	Write(status int, o interface{})
}

//Create a new factory
func (fac *Default) Create(res http.ResponseWriter, req *http.Request) formatters.ResponseFormatter {
	return formatters.NewJson(res, req)
}
