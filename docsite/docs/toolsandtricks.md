# Tools and Tricks

## k9s

## Kubernetes dashboard

## reloader

## kubectx



## Tricks: Another session

```
$ kubectl sk whoami
USER     ID   GROUPS
ladmin   0    skas-admin,system:masters
```

In another terminal:

```
export KUBECONFIG=/tmp/kconfig

$ kubectl sk init https://skas.ingress.ksprayX
Setup new context 'skas@ksprayX.vb' in kubeconfig file '/tmp/kconfig'

$ kubectx
skas@ksprayX.vb

$ kubectl get ns
Login:luser1
Password:
Error from server (Forbidden): namespaces is forbidden: User "luser1" cannot list resource "namespaces" in API group "" at the cluster scope

$ kubectl sk whoami
USER     ID   GROUPS
luser1   0
```

Back to initial terminal:

```
$ kubectl sk whoami
USER     ID   GROUPS
ladmin   0    skas-admin,system:masters

$ kubectl get ns
NAME              STATUS   AGE
cert-manager      Active   69m
default           Active   76m
ingress-nginx     Active   68m
kube-node-lease   Active   76m
kube-public       Active   76m
kube-system       Active   76m
kube-tools        Active   66m
kyverno           Active   69m
metallb-system    Active   69m
skas-system       Active   66m
topolvm-system    Active   68m
```
