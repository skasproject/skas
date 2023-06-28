
# Usage

## Initial configuration

### Local client configuration

It is assumed here than `kubectl` is installed. (If not, [see here](https://kubernetes.io/docs/tasks/tools/))

It is also assumed the `kubectl-sk` CLI extension has been installed (If not, [see here](./installation.md#skas-cli-installation))

For accessing a kubernetes cluster with kubectl, you need a configuration file (By default in `<homedir>/.kube/config`).

SKAS provide a mechanism to create or update this user's configuration file.

```
$ kubectl sk init https://skas.ingress.mycluster.internal
Setup new context 'skas@mycluster.internal' in kubeconfig file '/Users/john/.kube/config'
```

You can validate this new context is now the current one:

```shell
kubectl config current-context
skas@mycluster.internal
```

#### Got a certificate issue ?

If your system is not configured with the CA which has been used to certify SKAS (cf the `clusterIssuer` parameter on initial installation), you will encounter an error like:

```shell
ERRO[0000] error on GET kubeconfig from remote server  
 error="error on http connection: Get \"https://skas.ingress.mycluster.internal/v1/kubeconfig\": 
 tls: failed to verify certificate: x509: certificate signed by unknown authority"
```

You may get ride of this error by providing the root CA certificate as a file:

```shell
kubectl sk init https://skas.ingress.mycluster.internal --authRootCaPath=./CA.crt
```

> _A CA certificate file is a text file which begin by `-----BEGIN CERTIFICATE-----` and ends with `-----END CERTIFICATE-----`. 
Such CA file must have been provided to you by some system administrator._

If you are unable to get such CA certificate, you can skip the test by setting a flag:

```shell
kubectl sk init --authInsecureSkipVerify=true https://skas.ingress.mycluster.internal
```

But, be aware this is a security breach, as the target site can be a fake one. Use this flag should be limited to initial evaluation context.

### Use default admin account 

SKAS manage a local users database, where users are stored a Kubernetes resources. 


```
$ kubectl -n skas-system get skusers
Login:admin
Password:
NAME    COMMON NAMES             EMAILS   UID   COMMENT   DISABLED
admin   ["SKAS administrator"]
```

```
$ kubectl -n skas-system get groupbindings
NAME               USER    GROUP
admin-skas-admin   admin   skas-admin
```





```shell
kubectl explain pods
Login:admin
Password:
KIND:       Pod
VERSION:    v1

DESCRIPTION:
Pod is a collection of containers that can run on a host. This resource is
created by clients and scheduled onto hosts.
...........
```




```
kubectl get ns
Error from server (Forbidden): namespaces is forbidden: User "admin" cannot list resource "namespaces" in API group "" at the cluster scope
```



```
$ kubectl sk user bind admin system:masters
GroupBinding 'admin.system.masters' created in namespace 'skas-system'.
```

```
$ kubectl sk logout
Bye!

$ kubectl get ns
Login:admin
Password:
NAME              STATUS   AGE
cert-manager      Active   4d21h
default           Active   4d21h
ingress-nginx     Active   4d21h
.....
```

```
$ kubectl sk password
Will change password for user 'admin'
Old password:
New password:
Confirm new password:
Password has been changed sucessfully.
```


```
$ kubectl sk password
Will change password for user 'admin'
Old password:
New password:
Confirm new password:
Unsatisfactory password strength!
```





-------------------------------------------------------------------------------------------------------------

# Getting started


## Initial local admin creation

```
# Set As a kube administator
KUBECONFIG=....../ksprayX/build/config

$ kubectl sk user create ladmin --commonName "Local admin" --email "ladmin@ksprayX.local" --inputPassword
User 'ladmin' created in namespace 'skas-system'.

$ kubectl sk user bind ladmin system:masters
GroupBinding 'ladmin.system.masters' created in namespace 'skas-system'.

unset KUBECONFIG
```

## User context initialisation

```
$ kubectl sk init https://skas.ingress.ksprayX
Setup new context 'skas@ksprayX.vb' in kubeconfig file '/Users/sa/.kube/config'

$ kubectx
skas@ksprayX.vb
skas@kspray3.vb

$ kubectl get ns
Login:ladmin
Password:
NAME              STATUS   AGE
cert-manager      Active   23h
.......

```

Set also as skas administrator

```
$ kubectl sk user describe admin
Unauthorized!

$ kubectl sk user bind ladmin skas-admin
GroupBinding 'ladmin.skas-admin' created in namespace 'skas-system'.

$ kubectl sk logout
Bye!

$ kubectl sk user describe admin
Login:ladmin
Password:
USER    STATUS         UID   GROUPS   EMAILS   COMMON NAMES   AUTH
admin   userNotFound   0

$ kubectl sk user describe ladmin
USER     STATUS              UID   GROUPS                      EMAILS                 COMMON NAMES   AUTH
ladmin   passwordUnchecked   0     skas-admin,system:masters   ladmin@ksprayX.local   Local admin    crd

```

Create another user

```
$ kubectl sk user create luser1 --commonName "Local user1" --email "luser1@ksprayX.local" --password luser1
User 'luser1' created in namespace 'skas-system'.
```

List local users

```
$ kubectl -n skas-system get users
NAME     COMMON NAMES      EMAILS                     UID   COMMENT   DISABLED
ladmin   ["Local admin"]   ["ladmin@ksprayX.local"]                   false
luser1   ["Local user1"]   ["luser1@ksprayX.local"]                   false
```

# Tricks: Another session

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

# Argo cd

Login on the front using ladmin account.

View user

Try to create a project. Should fail

```
$ kubectl sk user bind ladmin argocd-admin
```

Now, on the front, logout and login.

View user

Try to create a project. Should fail


Using the command line:

```
$ argocd login argocd.ingress.ksprayX --username ladmin --sso
Opening browser for authentication
Performing authorization_code flow login: https://argocd.ingress.ksprayX/api/dex/auth?access_type=offline&client_id=argo-cd-cli&code_challenge=R40jEBd-oQZ48N4tI2amiEg_0UULb_V4ARUk_U3r1Hc&code_challenge_method=S256&redirect_uri=http%3A%2F%2Flocalhost%3A8085%2Fauth%2Fcallback&response_type=code&scope=openid+profile+email+groups+offline_access&state=GLEuRGPrdUsKGNCDZCmsxhxS
Authentication successful
'ladmin@ksprayX.local' logged in successfully
Context 'argocd.ingress.ksprayX' updated
```

# Services setup

## Prerequisite

Setup ook8s and launch scripts for secret creation.

Add repo:

```
$ argocd repo add https://github.com/KubeDP/kdp-savb.git --username "SergeALEXANDRE" --password <pull kdc01 token>
```

Login on argocd (as admin)

And apply apps of apps ....

```

cd .../kdp-savb/kargo
kubectl apply -f metaX.yaml
```

Now, skas config has been modified to integrate an ldap server, with 'sa' account

```
kubectl sk user describe sa --explain
USER   STATUS              UID    GROUPS                              EMAILS                 COMMON NAMES      AUTH
sa     passwordUnchecked   2002   all,dawf01-admin,devs,inst1-admin   sa@broadsoftware.com   Serge ALEXANDRE   ldap

Detail:
PROVIDER   STATUS              UID    GROUPS                              EMAILS                 COMMON NAMES
crd        userNotFound        0
ldap       passwordUnchecked   2002   all,devs,inst1-admin,dawf01-admin   sa@broadsoftware.com   Serge ALEXANDRE
```

On can log as 'sa' on argocd.

But to be able to work on it:

```
$ kubectl sk user bind sa argocd-admin
GroupBinding 'sa.argocd-admin' created in namespace 'skas-system'.
```


One can also log on spark-histo (ladmin and sa) and argocd of inst1