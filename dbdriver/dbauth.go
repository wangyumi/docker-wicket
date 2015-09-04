package dbdriver

import (
	"github.com/tg123/docker-wicket/acl"
)

type Driver struct {
	dbc *DbDriver
}

func init() {
	d := &Driver{}

	dbc, err := NewDbDriver()

	if err != nil {
		panic(err.Error())
	}

	d.dbc = dbc

	acl.Register("mysqlauth", d, func() error {
		return nil
	})
}

func (d *Driver) CanLogin(username acl.Username, password acl.Password) (bool, error) {
	return d.dbc.CanLogin(string(username), string(password)), nil
}

func (d *Driver) CanAccess(username acl.Username, namespace, repo string, perm acl.Permission) (bool, error) {
	return d.dbc.CanAccess(string(username), namespace, repo, int(perm)), nil
}
