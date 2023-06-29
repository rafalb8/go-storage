package jsondb

import "fmt"

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
