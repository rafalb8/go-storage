package jsondb

import (
	"fmt"

	"github.com/rafalb8/go-storage"
)

type JsonDBOpts func(*JsonDB) error

func File(path string) JsonDBOpts {
	return func(j *JsonDB) error {
		j.path = path
		j.singleFile = true
		return nil
	}
}

func Dir(dir string) JsonDBOpts {
	return func(j *JsonDB) error {
		j.path = dir
		return fmt.Errorf("JsonDB.Dir: not supported")
	}
}

func Logger(lg storage.Logger) JsonDBOpts {
	return func(j *JsonDB) error {
		j.lg = lg
		return nil
	}
}
