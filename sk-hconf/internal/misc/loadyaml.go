package misc

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"strings"
)

func LoadYaml(fileName string, target interface{}) error {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	cnt, err := ExpandEnv(string(content))
	if err != nil {
		return fmt.Errorf("error in '%s': %w", fileName, err)
	}
	dec := yaml.NewDecoder(strings.NewReader(cnt))
	dec.KnownFields(true)
	err = dec.Decode(target)
	//err = yaml.Unmarshal([]byte(cnt), target)

	if err != nil {
		if err != io.EOF { // EOF is not an error. Just an empty file (with or without comment)
			return fmt.Errorf("error while unmarshalling '%s': %w", fileName, err)
		}
	}
	return nil
}
