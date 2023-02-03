package misc

import "fmt"

// To display token or base64 encoded certificate in error message

func ShortenString(str string) string {
	if len(str) <= 30 {
		return str
	} else {
		return fmt.Sprintf("%s.......%s", str[:10], str[len(str)-10:])
	}
}
