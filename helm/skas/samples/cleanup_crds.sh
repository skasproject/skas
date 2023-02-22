#!/bin/bash


kubectl delete crd tokens.session.skasproject.io
kubectl delete crd groupbindings.userdb.skasproject.io
kubectl delete crd users.userdb.skasproject.io


