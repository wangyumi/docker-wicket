package dbdriver

import (
	"github.com/tg123/docker-wicket/index"
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
		return nil
	})
}

func (d *IndexDriver) GetIndexImages(namespace, repo string) ([]index.Image, error) {
	return []index.Image{}, nil
}

func (d *IndexDriver) UpdateIndexImages(namespace, repo string, images []index.Image) error {
	return nil
}

func (d *IndexDriver) CreateRepo(namespace, repo string) error {
	return d.dbc.CreateRepo(namespace, repo)
}

func (d *IndexDriver) DeleteRepo(namespace, repo string) error {
	return nil
}
