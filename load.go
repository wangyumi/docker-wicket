package main

import (
	_ "github.com/tg123/docker-wicket/acl/derelict"
	_ "github.com/tg123/docker-wicket/acl/htpasswd"
	_ "github.com/tg123/docker-wicket/acl/interdict"
	_ "github.com/tg123/docker-wicket/acl/mysqlauth"
	_ "github.com/tg123/docker-wicket/index/file"
	_ "github.com/tg123/docker-wicket/index/mem"
	_ "github.com/tg123/docker-wicket/mysqlpwd"
)
