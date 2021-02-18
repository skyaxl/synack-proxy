package registry

import (
	"context"
	"net/http"
	"time"

	"github.com/kataras/golog"
	cache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/skyaxl/synack-proxy/pkg/registryclient"
)

//Registry registrador
type Registry interface {
	Authenticate(ctx context.Context, user, password string) (ok bool, err error)
	Reg(ctx context.Context, dumpReq, dumpRes []byte) error
}

//RegistryProvider struct to provide
type RegistryProvider struct {
	cache  *cache.Cache
	server string
}

//Get get registry
func (reg *RegistryProvider) Get(username string) Registry {
	if obj, ok := reg.cache.Get(username); ok {
		if res, okCast := obj.(*SideCarRegistry); okCast {
			return res
		}
	}

	scr := &SideCarRegistry{server: reg.server, username: username}
	reg.cache.Set(username, scr, time.Hour)
	return scr
}

//NewProvider  new default provider
var NewProvider = func(server string) *RegistryProvider {
	return &RegistryProvider{
		server: server,
		cache:  cache.New(time.Hour, time.Hour),
	}
}

//SideCarRegistry registry
type SideCarRegistry struct {
	server      string
	client      registryclient.ClientWithResponsesInterface
	username    string
	autenticate bool
}

//Authenticate authenticate
func (r *SideCarRegistry) Authenticate(ctx context.Context, user string, password string) (ok bool, err error) {

	if r.autenticate {
		return true, nil
	}

	r.client, err = registryclient.NewClientWithResponses(r.server, registryclient.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		golog.Infof("[Proxy API] Setting basic credentions user %s password %s", user, password)
		req.SetBasicAuth(user, password)
		return nil
	}))

	if err != nil {
		golog.Errorf("Error to create cliente err: %v", err)
		return false, err
	}
	var userRegistry *registryclient.GetuserResponse
	if userRegistry, err = r.client.GetuserWithResponse(ctx, user); err != nil || userRegistry.JSON200 == nil {
		golog.Errorf("User %s was not authorized to access proxy err: %v", user, err)
		return false, errors.Wrapf(err, "User %s was not authorized to access proxy", user)
	}
	r.autenticate = true
	golog.Infof("User logger: %v", userRegistry.JSON200.Name)

	return true, nil
}

//Reg register
func (r *SideCarRegistry) Reg(ctx context.Context, dumpReq []byte, dumpRes []byte) error {
	//golog.Debugf("Sending dump %v", string(dumpReq))
	//golog.Debugf("Sending dump res %v", string(dumpRes))
	_, err := r.client.RegWithResponse(ctx, registryclient.RegJSONRequestBody{
		Username: r.username,
		DumpReq:  dumpReq,
		DumpRes:  dumpRes,
	})
	return err
}
