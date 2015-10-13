package dbdriver

import (
	"database/sql"
	"github.com/docker/docker/pkg/mflag"
	_ "github.com/go-sql-driver/mysql"
	"github.com/tg123/docker-wicket/acl"
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

	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	dbc = &DbDriver{}

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

	log.Printf("INFO: data source : %s", dataSourceName)

	db, err := sql.Open("mysql", dataSourceName)

	if err != nil {
		log.Printf("ERROR: %v", err)
	}

	dbc.db = db

	return dbc, nil
}

func (dbc *DbDriver) CanLogin(username string, password string) bool {
	log.Printf("INFO: Authenticate user %s", username)

	if acl.Username(username) == acl.Anonymous {
		return true
	}
	var passwd string
	err := dbc.db.QueryRow("select token from t_users where name = ?", username).Scan(&passwd)

	switch {
	case err == sql.ErrNoRows:
		log.Printf("INFO: unknown user %s", username)
		return false
	case err != nil:
		log.Printf("ERROR: %v", err)
		return false
	default:
		if passwd == password {
			log.Printf("INFO: %s has correct passwod", username)
			return true
		} else {
			log.Printf("INFO: %s has wrong passwod", username)
			return false
		}
	}
}

func (dbc *DbDriver) CanAccess(username string, namespace string, repo string, permission int) bool {
	var ret bool
	if namespace == "library" {
		ret = dbc.canAccessPublic(username, namespace, repo, permission)
	} else {
		ret = dbc.canAccessPrivate(username, namespace, repo, permission)
	}
	log.Printf("Verify if user '%s' can access image %s/%s, permission=%d, canAccess=%t", username, namespace, repo, permission, ret)
	return ret
}

func (dbc *DbDriver) canAccessPublic(username string, namespace string, repo string, permission int) (ret bool) {

	if permission == 0 {
		//read, pull
		ret = true
		return
	}

	if username == "" {
		//anonymous cannot push image
		ret = false
		return
	}

	var project_id int
	err := dbc.db.QueryRow("select project_id from t_repos where project_id=1 and name=? limit 0,1", repo).Scan(&project_id)
	switch {
	case err == sql.ErrNoRows:
	case err != nil:
		log.Printf("ERROR: %v", err)
	default:
		log.Printf("INFO: cannot push docker.io images from docker registry directly, repo=%s", repo)
		ret = false
		return
	}

	//verify write permission
	var user, projectname, imagename string
	var uid int
	err = dbc.db.QueryRow("select t_users.uid, t_users.name, t_projects.name, t_repos.name from t_users, t_repos, t_projects where t_users.uid=t_repos.uid and t_projects.id=t_repos.project_id and t_projects.name= 'public' and t_repos.name= ? limit 0,1", repo).Scan(&uid, &user, &projectname, &imagename)

	switch {
	case err == sql.ErrNoRows:
		//first write
		log.Printf("INFO: do not find repos %s/%s", namespace, repo)
		if err = dbc.db.QueryRow("select uid from t_users where name= ?", username).Scan(&uid); err != nil {
			log.Printf("ERROR: fail tp create project %s -- %v", namespace, err)
		} else {
			go dbc.CreateRepoByUser(uid, "public", repo)
		}
		ret = true
		return
	case err != nil:
		log.Printf("ERROR: %v", err)
		ret = false
		return
	}

	log.Printf("Image %s owner: project=%s, user=%s", imagename, projectname, user)

	if user == username {
		ret = true
	} else {
		ret = false
	}

	return
}

func (dbc *DbDriver) canAccessPrivate(username string, namespace string, repo string, permission int) (ret bool) {

	if username == "" {
		//anonymous cannot push/pull private images
		ret = false
		return
	}

	repo = namespace + "/" + repo

	if permission == 1 {
		//write, push
		var project_id int
		err := dbc.db.QueryRow("select project_id from t_repos where project_id=1 and name=? limit 0,1", repo).Scan(&project_id)
		switch {
		case err == sql.ErrNoRows:
		case err != nil:
			log.Printf("ERROR: %v", err)
		default:
			log.Printf("INFO: cannot push docker.io images from docker registry directly, repo=%s", repo)
			ret = false
			return
		}
	}

	var user, project string
	var uid int
	err := dbc.db.QueryRow("select t_users.uid, t_users.name, t_projects.name from t_users, t_projects where t_users.uid=t_projects.uid and t_projects.name= ? limit 0,1", namespace).Scan(&uid, &user, &project)
	switch {
	case err == sql.ErrNoRows:
		if permission == 1 {
			//create project proactively if first write a namespace
			if err = dbc.db.QueryRow("select uid from t_users where name= ?", username).Scan(&uid); err != nil {
				log.Printf("ERROR: fail tp create project %s -- %v", namespace, err)
			} else {
				if _, err = dbc.db.Exec("insert into t_projects(uid,name,scm_type,scm_address) values(?,?,'none','')", uid, namespace); err != nil {
					log.Printf("ERROR: fail to create project %s -- %v", namespace, err)
				} else {
					go dbc.CreateRepoByUser(uid, namespace, repo)
				}
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

func (dbc *DbDriver) CreateRepoByUser(uid int, namespace, repo string) error {

	log.Printf("INFO: will create repo %s/%s by uid=%d", namespace, repo, uid)

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
		log.Printf("INFO: repo %s/%s is existing in DB", namespace, repo)
		return nil
	}

	var project_id int
	err = tx.QueryRow("select id from t_projects where name =?", namespace).Scan(&project_id)
	if err != nil {
		log.Printf("ERROR: create repo (%s/%s) fails -- %v", namespace, repo, err)
		return nil
	}

	if _, err = tx.Exec("insert into t_repos(uid, project_id, name, updated) values(?, ?, ?, 0)", uid, project_id, repo); err != nil {
		log.Printf("ERROR: create repo (%s/%s) fails -- %v", namespace, repo, err)
	}
	return nil

}

func (dbc *DbDriver) CreateRepo(namespace, repo string) error {

	log.Printf("INFO: will create repo %s/%s", namespace, repo)

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
		log.Printf("INFO: repo %s/%s is existing in DB", namespace, repo)
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
