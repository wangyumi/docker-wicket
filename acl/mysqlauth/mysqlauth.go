package mysqlauth

import (
	"github.com/tg123/docker-wicket/acl"
	"github.com/tg123/docker-wicket/mysqlpwd"
)

type Driver struct {
	dbc *mysqlpwd.DbConnection
}

func init() {
	d := &Driver{}

	dbc, err := mysqlpwd.New()

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
