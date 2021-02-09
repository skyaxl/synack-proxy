package registry

import (
	"context"

	"github.com/kataras/golog"
)

type LocalRegistry struct {
}

//Authenticate authenticate
func (r LocalRegistry) Authenticate(ctx context.Context, user string, password string) (ok bool, err error) {
	golog.Infof("User %s, pass %s", user, password)
	return true, nil
}

//Reg register
func (r LocalRegistry) Reg(ctx context.Context, dumpReq []byte, dumpRes []byte) error {
	golog.Infof("Dump %v", string(dumpReq))
	golog.Infof("Dump res %v", string(dumpRes))
	return nil
}
