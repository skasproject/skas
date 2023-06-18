package passwordvalidator

import (
	"bufio"
	"github.com/nbutton23/zxcvbn-go"
	"strings"
)

var cpMap map[string]bool

func init() {
	cpMap = make(map[string]bool)

	scanner := bufio.NewScanner(strings.NewReader(commonPasswordUsTxt))
	for scanner.Scan() {
		//fmt.Printf("--%s--\n", scanner.Text())
		cpMap[scanner.Text()] = true
	}
	scannerFr := bufio.NewScanner(strings.NewReader(commonPasswordUsTxt))
	for scannerFr.Scan() {
		//fmt.Printf("--%s--\n", scanner.Text())
		cpMap[scannerFr.Text()] = true
	}

}

func Validate(password string, userInputs []string) ( /*score*/ int /*isCommon*/, bool) {
	passwordStrength := zxcvbn.PasswordStrength(password, userInputs)
	_, isCommon := cpMap[strings.ToLower(password)]
	return passwordStrength.Score, isCommon
}
