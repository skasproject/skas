package misc

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var test1 = `Ceci est un test
Home: '${MYHOME}'
Et oui
Autre ${123456
}
User: ${MYUSER}
Virtual env: "${VIRTUAL_ENV}"
Fin`

var test1r = `Ceci est un test
Home: '/home/user'
Et oui
Autre ${123456
}
User: fred
Virtual env: "/home/user/venv"
Fin`

func TestTest1(t *testing.T) {
	err := os.Setenv("MYHOME", "/home/user")
	assert.Nil(t, err)
	err = os.Setenv("VIRTUAL_ENV", "/home/user/venv")
	assert.Nil(t, err)
	err = os.Setenv("MYUSER", "fred")
	assert.Nil(t, err)

	x, err := ExpandEnv(test1)
	assert.Nil(t, err)
	//fmt.Println(x)
	assert.Equal(t, x, test1r)
}

var test2 = `Test error
User: '${USER}'
Missing var: ${A_VARIABLE}
Fin`

func TestTest2(t *testing.T) {
	_, err := ExpandEnv(test2)
	assert.NotNil(t, err)
	assert.Equal(t, 3, err.(MissingVariableError).line)
	assert.Equal(t, "A_VARIABLE", err.(MissingVariableError).variable)
	assert.Equal(t, "variable 'A_VARIABLE' at line 3 is not defined", err.Error())
}

var test3 = `${VAR}
Test end

${MYHOME}`

var test3r = `value
Test end

/home/user`

func TestTest3(t *testing.T) {
	err := os.Setenv("MYHOME", "/home/user")
	assert.Nil(t, err)
	err = os.Setenv("VAR", "value")
	assert.Nil(t, err)

	x, err := ExpandEnv(test3)
	assert.Nil(t, err)
	//fmt.Println(x)
	assert.Equal(t, x, test3r)
}

var test4 = `${XXXX
Test end

${MYHOME`

var test4r = `${XXXX
Test end

${MYHOME`

func TestTest4(t *testing.T) {
	err := os.Setenv("MYHOME", "/home/user")
	assert.Nil(t, err)

	x, err := ExpandEnv(test4)
	assert.Nil(t, err)
	//fmt.Println(x)
	assert.Equal(t, x, test4r)
}
