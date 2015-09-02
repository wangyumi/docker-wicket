package mysqlauth

import (
	"github.com/docker/docker/pkg/mflag"
	_ "github.com/go-sql-driver/mysql"
	"github.com/tg123/docker-wicket/acl"
	"github.com/tg123/docker-wicket/mysqlpwd"
)

type Driver struct {
	dbc *mysqlpwd.DbConnection
}

func init() {
	d := &Driver{}

	var username string
	var password string
	var ip string
	var port uint
	var database string

	mflag.StringVar(&username, []string{"-mysql_username"}, "docker", "username for mysql")
	mflag.StringVar(&password, []string{"-mysql_password"}, "dockerdocker", "username for mysql")
	mflag.StringVar(&ip, []string{"-mysql_ip"}, "10.97.232.22", "ip for mysql")
	mflag.UintVar(&port, []string{"-mysql_port"}, 3306, "port for mysql")
	mflag.StringVar(&database, []string{"-mysql_database"}, "docker", "database name for mysql")

	acl.Register("mysqlauth", d, func() error {

		dbc, err := mysqlpwd.New(username, password, ip, port, database)

		if err != nil {
			panic(err.Error())
		}

		d.dbc = dbc

		return nil
	})
}

func (d *Driver) CanLogin(username acl.Username, password acl.Password) (bool, error) {
	return d.dbc.CanLogin(string(username), string(password)), nil
}

func (d *Driver) CanAccess(username acl.Username, namespace, repo string, perm acl.Permission) (bool, error) {
	return d.dbc.CanAccess(string(username), namespace, repo, int(perm)), nil
}
