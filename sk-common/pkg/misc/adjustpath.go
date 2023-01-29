package misc

import "path/filepath"

func AdjustPath(baseFolder string, path string) string {
	if path != "" {
		if !filepath.IsAbs(path) {
			path = filepath.Join(baseFolder, path)
		}
		path = filepath.Clean(path)
	}
	return path
}
