# Tools and Tricks

## Secret generator

As stated in [Advanced configuration](advancedconfiguration.md#use-a-kubernetes-secrets), there is the need to generate 
a random secret in the deployment. For this, one can use [kubernetes-secret-generator](https://github.com/mittwald/kubernetes-secret-generator),
a custom kubernetes controller.

Here is a manifest which, once applied, will create the secret `ldap2-client-secret` used the authenticate the communication between the two PODs of the two LDAP configuration referenced above.  

```yaml
---
apiVersion: "secretgenerator.mittwald.de/v1alpha1"
kind: "StringSecret"
metadata:
  name: ldap2-client-secret
  namespace: skas-system
spec:
  fields:
    - fieldName: "clientSecret"
      encoding: "base64"
      length: "15"
```

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
