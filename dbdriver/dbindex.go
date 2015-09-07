package dbdriver

import (
	"github.com/tg123/docker-wicket/index"
	"log"
)

type IndexDriver struct {
	dbc *DbDriver
}

func init() {

	d := &IndexDriver{}

	dbc, err := NewDbDriver()

	if err != nil {
		panic(err.Error())
	}

	d.dbc = dbc

	index.Register("v1hub", d, func() error {
		log.Printf("INFO: register index driver v1hub")
		return nil
	})
}

func (d *IndexDriver) GetIndexImages(namespace, repo string) ([]index.Image, error) {
	log.Printf("INFO: namespace=%s, repo=%s", namespace, repo)
	return []index.Image{}, nil
}

func (d *IndexDriver) UpdateIndexImages(namespace, repo string, images []index.Image) error {
	log.Printf("INFO: namespace=%s, repo=%s", namespace, repo)
	return nil
}

func (d *IndexDriver) CreateRepo(namespace, repo string) error {
	log.Printf("INFO: namespace=%s, repo=%s", namespace, repo)
	return d.dbc.CreateRepo(namespace, repo)
}

func (d *IndexDriver) DeleteRepo(namespace, repo string) error {
	log.Printf("INFO: namespace=%s, repo=%s", namespace, repo)
	return nil
}
