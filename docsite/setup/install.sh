#!/bin/bash


MYDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"


BASE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"


if [[  -f /etc/redhat-release ]]
then
	echo "On a RHEL or Centos system."
	echo "Assuming python3,virtualenv and python3-pip already installed"
	echo "else: sudo yum install python3 python3-pip python-virtualenv "
	VENV=venv_rhel
	virtualenv -p python3  "${MYDIR}/../${VENV}"
	echo "*" >"${MYDIR}/../${VENV}/.gitignore"
  # shellcheck disable=SC1090
  source "${MYDIR}"/../${VENV}/bin/activate
  pip install --upgrade pip
  pip install -r "${MYDIR}"/requirements.txt
elif [[ "$OSTYPE" == "darwin"* ]]
then
	echo "On a MacOS system."
	VENV=venv
	echo "Assuming python3 virtualenv and python3 pip already installed"
	PYTHON=$(type python3 | awk '{ print $3 }')
	if [  ! -x "$PYTHON" ]
	then
		echo "Missing python3 interpreter"
		exit 1
	fi
	virtualenv --python="$PYTHON" "${MYDIR}"/../${VENV}
  # shellcheck disable=SC1090
	if [ -f "${BASE_DIR}"/venv/bin/activate ]; then
	  source "${BASE_DIR}"/venv/bin/activate
	elif [ -f "${BASE_DIR}"/venv/usr/local/bin/activate ]; then
	  source "${BASE_DIR}"/venv/usr/local/bin/activate
	else
	  echo "Unable to find an 'activate' script!"
	  exit 1
	fi
  pip install --upgrade pip
  pip install -r "${MYDIR}"/requirements.txt
elif [[  -f /etc/os-release ]]
then
  . /etc/os-release
  if [[ "x$NAME" == "xUbuntu" ]]
  then
    echo "On Ubuntu system"
  	echo "Assuming python3, virtualenv and python3-pip already installed"
	  echo "else: sudo apt update && sudo apt install python3 python3-pip virtualenv "
    VENV=venv
  	virtualenv -p python3  "${MYDIR}/../${VENV}"
  	echo "*" >"${MYDIR}/../${VENV}/.gitignore"
    source "${MYDIR}"/../${VENV}/bin/activate
    pip install --upgrade pip
    pip install -r "${MYDIR}"/requirements.txt
  else
    echo
    echo "Not on a RHEL,Centos, Ubuntu or MacOs system. Exiting!"
    # shellcheck disable=SC2034
    exit 1
  fi
else
	echo
	echo "Not on a RHEL,Centos, Ubuntu or MacOs system. Exiting!"
	# shellcheck disable=SC2034
	exit 1
fi
