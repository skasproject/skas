package misc

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func LoadConfig(configFile string, conf interface{}) (absConfigFile string, err error) {
	absConfigFile, err = filepath.Abs(configFile)
	if err != nil {
		return "", err
	}
	content, err := os.ReadFile(absConfigFile)
	if err != nil {
		return absConfigFile, err
	}
	content2, err := ExpandEnv(string(content))
	if err != nil {
		return absConfigFile, err
	}
	decoder := yaml.NewDecoder(strings.NewReader(content2))
	decoder.SetStrict(true)
	if err = decoder.Decode(conf); err != nil {
		if err == io.EOF {
			// Empty file is not an error
			return absConfigFile, nil
		}
		return absConfigFile, fmt.Errorf("file '%s': %w", configFile, err)
	}
	return absConfigFile, nil
}
