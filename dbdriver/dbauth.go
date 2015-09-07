package dbdriver

import (
	"github.com/tg123/docker-wicket/acl"
	"log"
)

type AclDriver struct {
	dbc *DbDriver
}

func init() {
	d := &AclDriver{}

	dbc, err := NewDbDriver()

	if err != nil {
		panic(err.Error())
	}

	d.dbc = dbc

	acl.Register("mysqlauth", d, func() error {
		log.Printf("INFO: register acl driver mysqlauth")
		return nil
	})
}

func (d *AclDriver) CanLogin(username acl.Username, password acl.Password) (bool, error) {
	return d.dbc.CanLogin(string(username), string(password)), nil
}

func (d *AclDriver) CanAccess(username acl.Username, namespace, repo string, perm acl.Permission) (bool, error) {
	return d.dbc.CanAccess(string(username), namespace, repo, int(perm)), nil
}
