#!/bin/bash



BASE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"


if [[  -f /etc/redhat-release ]]
then
	# shellcheck disable=SC1090
	source "${BASE_DIR}"/venv_rhel/bin/activate
	PS1='[mkdocs] \h:\W \u\$ '
elif [[  -f /etc/os-release ]]
then
  . /etc/os-release
  if [[ "x$NAME" == "xUbuntu" ]]
  then
    echo "On Ubuntu system"
  	source "${BASE_DIR}"/venv/bin/activate
	  PS1='[mkdocs] \h:\W \u\$ '
  else
    echo
    echo "Not on a RHEL,Centos, Ubuntu or MacOs system. Exiting!"
    # shellcheck disable=SC2034
    read a
    exit 1
  fi
elif [[ "$OSTYPE" == "darwin"* ]]
then
	# shellcheck disable=SC1090
	if [ -f "${BASE_DIR}"/venv/bin/activate ]; then
	  source "${BASE_DIR}"/venv/bin/activate
	elif [ -f "${BASE_DIR}"/venv/usr/local/bin/activate ]; then
	  source "${BASE_DIR}"/venv/usr/local/bin/activate
	else
	  echo "Unable to find an 'activate' script!"
	  read a
	  exit 1
	fi
	PS1='[mkdocs] \h:\W \u\$ '
else
	echo
	echo "Not on a RHEL or Centos or MacOs system. Exiting!"
	# shellcheck disable=SC2034
	read a
	exit 1
fi
