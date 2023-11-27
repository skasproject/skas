package misc

import (
	"fmt"
	"os"
)

func EnsureReadable(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("file '%s' does not exists", path)
	}
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("file '%s' exists but is not readable", path)
	}
	_ = f.Close()
	return nil
}
