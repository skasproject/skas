package loadsave

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"skas/sk-clientgo/internal/global"
)

func LoadStuff(path string, decode func(decoder *yaml.Decoder) error) bool {
	if file, err := os.Open(path); err == nil {
		dec := yaml.NewDecoder(file)
		dec.SetStrict(true)
		err = decode(dec)
		if err != nil {
			global.Log.Info("Invalid cached information. Will reset")
			return false
		}
		_ = file.Close()
		return true
	} else {
		return false
	}
}

func SaveStuff(path string, encode func(encoder *yaml.Encoder) error) error {
	err := ensureDir(filepath.Dir(path))
	if err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	encoder := yaml.NewEncoder(file)
	err = encode(encoder)
	if err != nil {
		return err
	}
	_ = encoder.Close()
	_ = file.Close()
	return nil
}

func ensureDir(dirName string) error {
	st, err := os.Stat(dirName)
	if err != nil {
		// We consider it is a file not found
		err = os.MkdirAll(dirName, 0700)
		if err != nil {
			return err
		}
		return nil
	}
	if !st.IsDir() {
		return fmt.Errorf("path '%s' is a file. We need this to be a folder", dirName)
	}
	return nil
}
