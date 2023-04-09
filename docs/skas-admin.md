
```
kubectl sk user create admin --commonName "Administrator" --inputPassword
Password:
Confirm password:
User 'admin' created in namespace 'skas-system'.
```


```
kubectl sk user bind admin skas-admin
GroupBinding 'admin.skas-admin' created in namespace 'skas-system'.
```

```
kubectl sk user bind admin 'system:masters'
GroupBinding 'admin.system.masters' created in namespace 'skas-system'.```
```
