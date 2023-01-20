

# Deployment

```
. ./tmp/.kspray6
```


```
cd helm/sk-crd
kubectl create ns skas-system
helm -n skas-system upgrade -i sk-crd1 .
cd ../..
```

Test
```
helm -n skas-system list
kubectl get crd | grep skas
```

```
cd helm/sk-userdb
kubectl create ns skas-userdb
helm -n skas-userdb upgrade -i sk-userdb1 .
cd ../..
```

curl -i -X GET http://localhost:7012/userstatus -d '{ "login": "raoul", "password": "raoul" }'


Delete

```
helm -n skas-userdb uninstall sk-userdb1
helm -n skas-system uninstall sk-crd1

kubectl delete crd users.userdb.skasproject.io  
kubectl delete crd groupbindings.userdb.skasproject.io  

```



