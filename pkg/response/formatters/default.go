package formatters

//ResponseFormatter format a response for http request
type ResponseFormatter interface {
	WriteError(status int, err error)
	Write(status int, o interface{})
}
