

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
