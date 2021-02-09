package proxy

import (
	"context"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/kataras/golog"
	"github.com/skyaxl/synack-proxy/pkg/response/formatters"
)

//Registry registrador
type Registry interface {
	Authenticate(ctx context.Context, user, password string) (ok bool, err error)
	Reg(ctx context.Context, dumpReq, dumpRes []byte) error
}

//ResponseFormatterFactory create a response format
type ResponseFormatterFactory interface {
	Create(res http.ResponseWriter, req *http.Request) formatters.ResponseFormatter
}

type Requester interface {
	Do(req *http.Request) (*http.Response, error)
}

//Handler structure
type Handler struct {
	registry  Registry
	resFac    ResponseFormatterFactory
	requester Requester
}

//NewHandler create new handler
func NewHandler(registry Registry, resFac ResponseFormatterFactory, requester Requester) *Handler {
	return &Handler{registry, resFac, requester}
}

// parseBasicAuth parses an HTTP Basic Authentication string.
// "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==" returns ("Aladdin", "open sesame", true).
//golang default example
func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	// Case insensitive prefix match. See Issue 22736.
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}

//ServeHTTP to intercept
func (p *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	start := time.Now()
	golog.Infof("Start %v \n", start)
	proxyAuth := req.Header.Get("Proxy-Authorization")
	user, pass, _ := parseBasicAuth(proxyAuth)
	resFmt := p.resFac.Create(rw, req)
	if ok, err := p.registry.Authenticate(req.Context(), user, pass); !ok {
		resFmt.WriteError(http.StatusForbidden, err)
		return
	}
	var dumped, resDumped []byte
	var err error
	if dumped, err = httputil.DumpRequest(req, true); err != nil {
		golog.Warnf("Error to get dump %v \n", err)
	}
	req.Header.Del("Accept-Encoding")
	// curl can add that, see
	// https://jdebp.eu./FGA/web-proxy-connection-header.html
	req.Header.Del("Proxy-Connection")
	req.Header.Del("Proxy-Authenticate")
	req.Header.Del("Proxy-Authorization")
	req.RequestURI = ""
	originalRequestStart := time.Now()
	res, err := p.requester.Do(req)

	if err != nil {
		resFmt.WriteError(http.StatusInternalServerError, err)
		return
	}

	originalRequestEnd := time.Now()
	if resDumped, err = httputil.DumpResponse(res, true); err != nil {
		golog.Warnf("Error to get res dump %v \n", err)
	}

	for k := range res.Header {
		rw.Header().Add(k, res.Header.Get(k))
	}

	if res.Body != nil {
		bts, _ := ioutil.ReadAll(res.Body)
		rw.Write(bts)
	}
	rw.WriteHeader(res.StatusCode)

	go func(ctx context.Context, reg Registry, dumped, resDumped []byte) {

		err := reg.Reg(ctx, dumped, resDumped)
		if err != nil {
			golog.Errorf("Error to save log: %v \n", err)
		}

	}(req.Context(), p.registry, dumped, resDumped)
	end := time.Now()
	total := end.Sub(start)
	onlyRequest := originalRequestEnd.Sub(originalRequestStart)
	golog.Infof("End %v, duration: %v, only request: %v, this app: %v \n", end, total, onlyRequest, total-onlyRequest)
}
