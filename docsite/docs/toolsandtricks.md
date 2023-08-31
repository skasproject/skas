# Tools and Tricks

## reloader

Forgetting to restart a POD after a configuration change is a common source of errors. Fortunately, some tools can 
help for this. Such as [Reloader](https://github.com/stakater/Reloader)

The SKAS Helm chart add appropriate annotations on the `deployment`:  

```
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    configmap.reloader.stakater.com/reload: skas-merge-config,skas-auth-config,skas-crd-config,
```

> _The list of `configMap` is built dynamically by the Helm chart._

Of course, if Reloader is not installed in your cluster, this annotation will have no effect.

## Secret generator

As stated in [Two LDAP servers configuration](twoldapservers.md) o [Delegated users management](delegated.md), there is the need to generate 
a random secret in the deployment. For this, one can use [kubernetes-secret-generator](https://github.com/mittwald/kubernetes-secret-generator),
a custom kubernetes controller.

Here is a manifest which, once applied, will create the secret `skas2-client-secret` used the authenticate the communication between the two PODs of the two LDAP configuration referenced above.  

``` { .yaml .copy }
---
apiVersion: "secretgenerator.mittwald.de/v1alpha1"
kind: "StringSecret"
metadata:
  name: skas2-client-secret
  namespace: skas-system
spec:
  fields:
    - fieldName: "clientSecret"
      encoding: "base64"
      length: "15"
```

## k9s

We would like to say two words about this great tool which is [k9s](https://github.com/derailed/k9s)

As it is able to handle Custom Resources Definition out of the box, K9s is a perfect tool to dynamically display, modify or delete SKAS resources.

Note than, as User and Group are ambiguous names, which are used also by others API, alias are provided to ensure ambiguous access.

For example, you can access this screen under `skusers` resource name:

![](images/k9s-1.png)

This one using `groupbindings`:

![](images/k9s-2.png)

This one using `tokens`:

![](images/k9s-3.png)

Of course, k9s can't do more than what the launching user is allowed to do. This user can be authenticated using SKAS, but it must have a minimum set of rights to behave correctly.

For example, you can launch k9s under the `admin` user account we have set up in the installation process (Provided it is member of the `system:masters` group).

```shell
$ kubectl sk login admin
Password:
logged successfully..

$ k9s
....
```

## Kubernetes dashboard

Login to the Kubernetes dashboard with SKAS is quite easy.

First, you must be logged using the CLI. Then using the `--all` option of the `kubectl sk whoami` command, you can get your current allocated token:

```shell
$ kubectl sk login admin
Password:
logged successfully..

$ kubectl sk whoami --all
USER    ID   GROUPS                      AUTH.   TOKEN
admin   0    skas-admin,system:masters   crd     znitotnewjbqbuolqacckvgxyhptoxsuykznrzdacuvdhimy
```

Now, you just have to cut and paste the token value in the dashboard login screen:

![](images/dashboard1.png)

Of course, the set of operation you will be able to perform through the dashboard will be limited by the logged user's permissions.

## Tricks: Handle two different sessions

When working on user permissions, it could be useful to have separate session, at least one as admin, and one as a user to test its capability.

But the default Kubernetes configuration is not bound to a terminal session, but to a user. 
So, any modification (`kubectl config ....`) of the local configuration will have effect on all session.

The solution is to change the location of the kubernetes configuration for a given session, by modifying the `KUBECONFIG` environment variable: 

```shell
$ export KUBECONFIG=/tmp/kconfig
```

> `/tmp/kconfig` may be an empty or un-existing file

Then you can initialize a new Kubernetes/SKAS context

```shell
$ kubectl sk init https://skas.ingress.mycluster.internal
Setup new context 'skas@mycluster.internal' in kubeconfig file '/tmp/kconfig'
```



