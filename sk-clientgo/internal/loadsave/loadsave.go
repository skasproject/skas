package loadsave

import (
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

func LoadStuff(path string, decode func(decoder *yaml.Decoder) error) bool {
	if file, err := os.Open(path); err == nil {
		dec := yaml.NewDecoder(file)
		dec.SetStrict(true)
		err = decode(dec)
		if err != nil {
			panic(err)
		}
		_ = file.Close()
		return true
	} else {
		return false
	}
}

func SaveStuff(path string, encode func(encoder *yaml.Encoder) error) {
	ensureDir(filepath.Dir(path))
	var err error
	var file *os.File
	if file, err = os.Create(path); err == nil {
		if err = encode(yaml.NewEncoder(file)); err == nil {
			err = file.Close()
		}
	}
	if err != nil {
		panic(err)
	}
}

func ensureDir(dirName string) {
	if _, serr := os.Stat(dirName); serr != nil {
		merr := os.MkdirAll(dirName, 0700)
		if merr != nil {
			panic(merr)
		}
	}
}
