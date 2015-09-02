package mysqlpwd

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strconv"
)

type mysqlConfig struct {
	Username string
	Password string
	Address  string
	Port     int
	Database string
}

type DbConnection struct {
	db *sql.DB
}

func New(username string, password string, ip string, port uint, database string) (*DbConnection, error) {
	dbc := &DbConnection{}

	dataSourceName := username + ":" + password + "@tcp(" + ip + ":" + strconv.Itoa(int(port)) + ")/" + database + "?charset=utf8"

	log.Printf("dat source : %s", dataSourceName)

	db, err := sql.Open("mysql", dataSourceName)

	if err != nil {
		log.Printf("ERROR: %v", err)
	}

	dbc.db = db

	return dbc, nil
}

func (dbc *DbConnection) CanLogin(username string, password string) bool {
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

func (dbc *DbConnection) CanAccess(username string, namespace string, repo string, permission int) bool {
	if namespace == "library" {
		return dbc.canAccessPublic(username, namespace, repo, permission)
	} else {
		return dbc.canAccessPrivate(username, namespace, repo, permission)
	}
}

func (dbc *DbConnection) canAccessPublic(username string, namespace string, repo string, permission int) (ret bool) {

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

func (dbc *DbConnection) canAccessPrivate(username string, namespace string, repo string, permission int) (ret bool) {
	defer log.Printf("Verify if user %s can access private image %s/%s, permission=%d, canAccess=%t", username, namespace, repo, permission, ret)

	repo = namespace + "/" + repo
	var t_users, t_projects, t_images string
	err := dbc.db.QueryRow("select t_users.name, t_projects.name , t_images.name from t_users, t_images, t_projects where t_users.uid=t_images.uid and  t_projects.id=t_images.project_id and t_projects.name= ? and t_images.name= ? limit 0,1", namespace, repo).Scan(&t_users, &t_projects, &t_images)

	switch {
	case err == sql.ErrNoRows:
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
