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

	scr := &SideCarRegistry{server: reg.server}
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
	server string
	client registryclient.ClientWithResponsesInterface
}

//Authenticate authenticate
func (r *SideCarRegistry) Authenticate(ctx context.Context, user string, password string) (ok bool, err error) {
	r.client, err = registryclient.NewClientWithResponses(r.server, registryclient.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.SetBasicAuth(user, password)
		return nil
	}))

	if err != nil {
		return false, err
	}
	var userRegistry *registryclient.GetuserResponse
	if userRegistry, err = r.client.GetuserWithResponse(ctx, user); err != nil || userRegistry.JSON200 == nil {
		golog.Errorf("User %s was not authorized to access proxy", user)
		return false, errors.Wrapf(err, "User %s was not authorized to access proxy", user)
	}

	golog.Infof("User logger: %v", userRegistry.JSON200.Name)

	return true, nil
}

//Reg register
func (r *SideCarRegistry) Reg(ctx context.Context, dumpReq []byte, dumpRes []byte) error {
	golog.Infof("Dump %v", string(dumpReq))
	golog.Infof("Dump res %v", string(dumpRes))
	return nil
}
