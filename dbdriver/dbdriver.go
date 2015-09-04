package dbdriver

import (
	"database/sql"
	"github.com/docker/docker/pkg/mflag"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strconv"
	"sync"
)

type mysqlConfig struct {
	Username string
	Password string
	Address  string
	Port     int
	Database string
}

type DbDriver struct {
	db *sql.DB
}

var (
	dbc  *DbDriver
	lock = &sync.Mutex{}
)

func NewDbDriver() (*DbDriver, error) {

	lock.Lock()
	defer lock.Unlock()

	if dbc != nil {
		return dbc, nil
	}

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

	dataSourceName := username + ":" + password + "@tcp(" + ip + ":" + strconv.Itoa(int(port)) + ")/" + database + "?charset=utf8"

	log.Printf("dat source : %s", dataSourceName)

	db, err := sql.Open("mysql", dataSourceName)

	if err != nil {
		log.Printf("ERROR: %v", err)
	}

	dbc.db = db

	return dbc, nil
}

func (dbc *DbDriver) CanLogin(username string, password string) bool {
	rows, err := dbc.db.Query("select token from t_users where name = ?", username)

	if err != nil {
		log.Printf("ERROR: %v", err)
		return false
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		log.Printf("ERROR: %v", err)
		return false
	}

	for rows.Next() {
		var passwd string
		if err := rows.Scan(&passwd); err != nil {
			log.Printf("ERROR: %v", err)
			return false
		}

		if passwd == password {
			log.Printf("INFO: %s has correct passwod", username)
			return true
		} else {
			log.Printf("INFO: %s has wrong passwod", username)
			return false
		}
	}

	return false
}

func (dbc *DbDriver) CanAccess(username string, namespace string, repo string, permission int) bool {
	if namespace == "library" {
		return dbc.canAccessPublic(username, namespace, repo, permission)
	} else {
		return dbc.canAccessPrivate(username, namespace, repo, permission)
	}
}

func (dbc *DbDriver) canAccessPublic(username string, namespace string, repo string, permission int) (ret bool) {

	defer log.Printf("Verify if user %s can access public image %s/%s, permission=%d, canAccess=%t", username, namespace, repo, permission, ret)

	if permission == 0 {
		//read
		ret = true
		return
	}

	//verify write permission
	var t_users, t_projects, t_images string
	err := dbc.db.QueryRow("select t_users.name, t_projects.name , t_images.name from t_users, t_images, t_projects where t_users.uid=t_images.uid and  t_projects.id=t_images.project_id and t_projects.name= 'public' and t_images.name= ? limit 0,1", repo).Scan(&t_users, &t_projects, &t_images)

	switch {
	case err == sql.ErrNoRows:
		//first write
		ret = true
		return
	case err != nil:
		log.Printf("ERROR: %v", err)
		ret = false
		return
	}

	log.Printf("Image %s owner: project=%s, user=%s", t_images, t_projects, t_users)

	if t_users == username {
		ret = true
	} else {
		ret = false
	}

	return
}

func (dbc *DbDriver) canAccessPrivate(username string, namespace string, repo string, permission int) (ret bool) {
	defer log.Printf("Verify if user %s can access private image %s/%s, permission=%d, canAccess=%t", username, namespace, repo, permission, ret)

	repo = namespace + "/" + repo
	var user, project string
	var uid int
	err := dbc.db.QueryRow("select t_users.uid, t_users.name, t_projects.name from t_users, t_projects where t_users.uid=t_projects.uid and t_projects.name= ? limit 0,1", namespace).Scan(&uid, &user, &project)

	switch {
	case err == sql.ErrNoRows:
		//create project proactively if first write a namespace
		if permission == 1 {
			if _, err = dbc.db.Exec("insert into t_projects(uid,name,scm_type,scm_address) values(?,?,'none',''", uid, namespace); err != nil {
				log.Printf("ERROR: fail tp create project %s -- %v", namespace, err)
			}
		}
		ret = true
		return
	case err != nil:
		log.Printf("ERROR: %v", err)
		ret = false
		return
	}

	log.Printf("Image %s owner: project=%s, user=%s", repo, project, user)

	if user == username {
		ret = true
	} else {
		ret = false
	}
	return
}

func (dbc *DbDriver) CreateRepo(namespace, repo string) error {
	tx, err := dbc.db.Begin()
	if err != nil {
		log.Printf("ERROR: create repo (%s/%s) fails -- %v", namespace, repo, err)
		return nil
	}

	defer func() {
		if err == nil {
			err = tx.Commit()
		} else {
			err = tx.Rollback()
		}
		if err != nil {
			log.Printf("ERROR: create repo (%s/%s) fails -- %v", namespace, repo, err)
		}
	}()

	var repo_id int
	err = tx.QueryRow("select t_repos.id from t_repos, t_projects where t_repos.project_id = t_projects.id and t_repos.name=? and  t_projects.name", repo, namespace).Scan(&repo_id)
	if err == nil {
		//repo is existing
		return nil
	}

	var project_id, uid int
	err = tx.QueryRow("select id,uid from t_projects where name =?", namespace).Scan(&project_id, &uid)
	if err != nil {
		log.Printf("ERROR: create repo (%s/%s) fails -- %v", namespace, repo, err)
		return nil
	}

	var id int
	err = tx.QueryRow("select id from t_repos where project_id =? and name=?", project_id, repo).Scan(&id)

	switch {
	case err == sql.ErrNoRows:
		if _, err = tx.Exec("insert into t_repos(uid, project_id, name) values(?, ?, ?)", uid, project_id, repo); err != nil {
			log.Printf("ERROR: create repo (%s/%s) fails -- %v", namespace, repo, err)
			return nil
		}
	case err != nil:
		log.Printf("ERROR: create repo (%s/%s) fails -- %v", namespace, repo, err)
		return nil
	default:
		log.Printf("ERROR: repo (%s/%s) is existing ", namespace, repo)
		return nil
	}
	return nil
}
