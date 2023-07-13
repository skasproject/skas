
# Usage


## Local client configuration

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
$ kubectl config current-context
skas@mycluster.internal
```

### Got a certificate issue ?

If your system is not configured with the CA which has been used to certify SKAS (cf the `clusterIssuer` parameter on initial installation), you will encounter an error like:

```shell
ERRO[0000] error on GET kubeconfig from remote server  
 error="error on http connection: Get \"https://skas.ingress.mycluster.internal/v1/kubeconfig\": 
 tls: failed to verify certificate: x509: certificate signed by unknown authority"
```

You may get rid of this error by providing the root CA certificate as a file:

```shell
$ kubectl sk init https://skas.ingress.mycluster.internal --authRootCaPath=./CA.crt
```

> _A CA certificate file is a text file which begin by `-----BEGIN CERTIFICATE-----` and ends with `-----END CERTIFICATE-----`. 
Such CA file must have been provided to you by some system administrator._

If you are unable to get such CA certificate, you can skip the test by setting a flag:

```shell
$ kubectl sk init --authInsecureSkipVerify=true https://skas.ingress.mycluster.internal
```

But, be aware this is a security breach, as the target site can be a fake one. Use this flag should be limited to initial evaluation context.

## First run with default admin account 

SKAS manage a local users database, where users are stored as Kubernetes resources.

For convenience, a first `admin` user has been created during the installation.  With password `admin`

By default, SKAS users are stored in the namespace `skas-system`.

You could list them, using standard kubectl commands. If you have configured your client as described above, you now 
have to be logged to perform any kubectl action. So the login/password interaction

```shell
$ kubectl -n skas-system get users.userdb.skasproject.io
Login:admin
Password:
NAME    COMMON NAMES             EMAILS   UID   COMMENT   DISABLED
admin   ["SKAS administrator"]
```

Several remarks:

- Default password is `admin`. **DON'T FORGET TO CHANGE IT**. See below.
- The `admin` user has been granted to access SKAS resources in `skas-system` namespace using kubernetes RBAC
- `kubectl -n skas-system get users` may no works, as `users` refers also to a standard kubernetes resources.

To ease SKAS user management, an alias `skuser` has been defined.

```shell
$ kubectl -n skas-system get skusers
NAME    COMMON NAMES             EMAILS   UID   COMMENT   DISABLED
admin   ["SKAS administrator"]
```

Note there is now no login/password interaction. A token has been granted during the first login. 
This token will expire after a delay of inactivity. (Like a Web session). This delay is 30mn by default.

### Password change

As stated above, you must change the password of this account:

```shell
$ kubectl sk password
Will change password for user 'admin'
Old password:
New password:
Confirm new password:
Password has been changed sucessfully.
```

Note the `sk`, as such command is performed by the SKAS kubectl extension.

There is a check about password strength. So, you may have such response:

```shell
$ kubectl sk password
Will change password for user 'admin'
Old password:
New password:
Confirm new password:
Unsatisfactory password strength!
```

There is no well defined password criteria (such as length, special character, etc...). 
An algorithm provide a score for the password, and this score must match a minimum (configurable) value.
There is also a check against a list of commonly used passwords.

The easiest way to overcome this restriction is to increase your password length.

### SKAS group binding

In fact, what has been granted to access SKAS resources is not the admin account (It could be), but a group named `skas-system`.

And the user `admin` has been included in the group by another SKAS resources named `groupbindings.userdb.skasproject.io`, with `groupbindings`as an alias/

```shell
$ kubectl -n skas-system get groupBindings
NAME               USER    GROUP
admin-skas-admin   admin   skas-admin
```

> _In kubernetes, a group does not exist as a concrete resources. It only exists as it is referenced by RBAC `roleBinding` or `clusterRoleBindigs`. Or by SKAS `groupBinding`_

### Be a cluster admin

Let's try the following:

```shell
$ kubectl get namespaces
Error from server (Forbidden): namespaces is forbidden: User "admin" cannot list resource "namespaces" in API group "" at the cluster scope
```

It is clear than we are successfully authenticated as `admin`, but this account has no permissions to perform cluster-wide operation.

Such permissions can be granted by binding this user to a group having such rights:

```
$ kubectl sk user bind admin system:masters
GroupBinding 'admin.system.masters' created in namespace 'skas-system'.
```

For this to be effective, logout and login back:

```
$ kubectl sk logout
Bye!

$ kubectl get namespaces
Login:admin
Password:
NAME              STATUS   AGE
cert-manager      Active   4d21h
default           Active   4d21h
ingress-nginx     Active   4d21h
.....
```

You can check the new `groupBindings` list:

```shell
$ kubectl -n skas-system get groupBindings
NAME                   USER    GROUP
admin-skas-admin       admin   skas-admin
admin.system.masters   admin   system:masters
```

**WARNING: This means any member of the group `skas-admin` can promote itself as a full cluster administrator. 
In fact, anybody able to create or modify resources in the `skas-admin` namespace can take control of the cluster. So, access to this namespace should be strictly controlled**  

Refer to [Advanced configuration/Delegated user management]() to delegate user management without compromise cluster security. 

## CLI users management

The SKAS kubectl extension plugin provide a `user` command with several subcommands

You can have a complete list of such subcommands:

```
$ kubectl sk user --help
.......
```

NB: _You must be logged as a member of the group `skas-admin` to be able to use this command._


### Create user

Here is an example of user's creation:

```shell
$ kubectl sk user create luser1 --commonName "Local user1" --email "luser1@internal" --password "RtVksSuMgP5f"
User 'luser1' created in namespace 'skas-system'.
```

The only mandatory parameters is the user's name:

```shell
$ kubectl sk user create luser2
User 'luser2' created in namespace 'skas-system'.
```

> _As there is no password provided, login to the user will be impossible_

A complete list of user's creation options can be displayed: 

```shell
$ kubectl sk user create --help
Create a new user

Usage:
  kubectl-sk user create <user> [flags]

Flags:
      --comment string        User's comment
      --commonName string     User's common name
      --email string          User's email
      --generatePassword      Generate and display a password
  -h, --help                  help for create
      --inputPassword         Interactive password request
  -n, --namespace string      User's DB namespace (default "skas-system")
      --password string       User's password
      --passwordHash string   User's password hash (Result of 'kubectl skas hash')
      --state string          User's state (enabled|disabled) (default "enabled")
      --uid int               User's UID

```

Most of the options match a user's properties

- `comment`, `commonName`, `email`, `uid` are just descriptive parameters, inspired from Unix user attributes.
- `state` will allow to temporary disable a user.

The `--namespace` allow to store the user resources in another namespace. See [Advanced configuration/Delegated user management]()

There is several options related to the password:

- `--password`: The password is provided as a parameter on the command line.
- `--inputPassword`: There will be a `Password: / Confirm password:` user interaction.
- `--generatePassword`: A random password is generated and displayed.
- `--passwordHash`: Provide the hash of the password, as it will be stored in the resource. 
  Use `kubectl sk hash` to generate the value. NB: Doing this way skip the check about password strength. 

### List users

Users can be listed using standard `kubectl` commands:

```
$ kubectl -n skas-system get skusers
NAME     COMMON NAMES             EMAILS                UID   COMMENT   DISABLED
admin    ["SKAS administrator"]
luser1   ["Local user1"]          ["luser1@internal"]                   false
luser2                                                                  false
```

### Modify user

A subcommand `patch` is provided to modify a user. As an example:

```shell
$ kubectl sk user patch luser2 --state=disabled
User 'luser2' updated in namespace 'skas-system'.

$ kubectl -n skas-system get skuser luser2
NAME     COMMON NAMES   EMAILS   UID   COMMENT   DISABLED
luser2                                           true
```

Most of the options are the same as the `user create` subcommand. 

There is also a `--create` option which will allow user creation if it does not exists.

### Delete user

Users can be deleted using standard `kubectl` commands:

```shell
$ kubectl -n skas-system delete skuser luser2
user.userdb.skasproject.io "luser2" deleted
```

### Manage user's groups and permissions.

To illustrate how SKAS interact with Kubernetes RBAC, we will setup a simple example. We will:

- Create a namespace named `ldemo`.
- Create a role named `configurator` in this namespace to manage resources of type `configMaps`.
- Create a roleBinding between this role and a group named `ldemo-devs`.
- Add the user `luser1` to this group.

We assume we are logged as 'admin' to perform theses tasks:

```shell
$ kubectl create namespace ldemo
namespace/ldemo created

$ kubectl -n ldemo create role configurator --verb='*' --resource=configMaps
role.rbac.authorization.k8s.io/configurator created

$ kubectl -n ldemo create rolebinding configurator-ldemo-devs --role=configurator --group=ldemo-devs
rolebinding.rbac.authorization.k8s.io/configurator-ldemo-devs created

$ kubectl sk user bind luser1 ldemo-devs
GroupBinding 'luser1.ldemo-devs' created in namespace 'skas-system'.

```
Now, we can test. First logout and login under `luser1`:

```shell
$ kubectl sk logout
Bye!

$ kubectl sk login
Login:luser1
Password:
logged successfully..

$ kubectl sk whoami
USER     ID   GROUPS
luser1   0    ldemo-devs
```

Now ensure we can create a `configMap` and view it.:

```shell
$ kubectl -n ldemo create configmap my-config --from-literal=key1=config1
configmap/my-config created

$ kubectl -n ldemo get configmaps my-config -o yaml
apiVersion: v1
data:
  key1: config1
kind: ConfigMap
metadata:
  creationTimestamp: "2023-07-11T14:56:27Z"
  name: my-config
  namespace: ldemo
  resourceVersion: "257983"
  uid: ad55b282-9803-4688-b2df-a1c35f708313
```

Also, ensure we can delete it

```shell
$ kubectl -n ldemo delete configmap my-config
configmap "my-config" deleted
```

> _Please, note than `roles` and `roleBindings` are namespaced resources while `users` and `groups` are cluster-wide resources._


### Kubernetes RBAC referential integrity

Kubernetes does not check referential integrity when creating a resource referencing another one. For example, the following will works:

```shell
kubectl -n ldemo create rolebinding missing-integrity --role=unexisting-role --group=unexisting-group
rolebinding.rbac.authorization.k8s.io/missing-integrity created

$ kubectl sk user bind unexisting-user unexisting-group
GroupBinding 'unexisting-user.unexisting-group' created in namespace 'skas-system'.
```

May be the referenced resource will be created later. Or the link will be useless.

This is clearly a design choice of Kubernetes. SKAS follow the same logic.

## Using Manifests instead of CLI

As users ans groups are defined as Kubernetes custom resources, they can be created and managed as any other resources, through manifests. 

By default, all SKAS users and groups resources are stored in the namespace `skas-system`.

Kubernetes RBAC has been configured during installation to allow management of such resources by all members of the `skas-admin` group.

### User resources

Here is the manifest corresponding to the users we created previously:

```yaml
---
apiVersion: userdb.skasproject.io/v1alpha1
kind: User
metadata:
  name: luser1
  namespace: skas-system
spec:
  commonNames:
  - Local user1
  emails:
  - luser1@internal
  passwordHash: $2a$10$q6nEVmP.MHo6VLAprTdTBuy6AHPel1uh3NocZdNjt.yh8HDE7Ja.m
```

- The resources name is the user login.
- The password is stored in a non reversible hash form. The command `kubectl sk hash` is provided to compute such hash.

Below is a sample of a user with all properties defined:

```yaml
---
apiVersion: userdb.skasproject.io/v1alpha1
kind: User
metadata:
  name: jsmith
  namespace: skas-system
spec:
  commonNames:  
    - John SMITH
  passwordHash: $2a$10$qumINdiGJIM1si2wi8ceDOczChq2twfDEDa6DR7jiYL8rJNzeYtmu
  emails: 
    - jsmith@mycompany.com
  uid: 100001
  comment: A sample user
  disabled: false 
```

To define such user, save the yaml definition if a file and perform a `kubectl apply -f <filaName>`

> _Unfortunately, when logged using SKAS, it is impossible to use stdin on kubectl. <br>So, `cat <filename> | kubectl apply -f -` 
will not work. This is inherent to the way the kubernetes client-go credential plugin works_

### GroupBinding resources

The SKAS `GroupBinding` resources can also be defined as manifest:

```yaml
---
apiVersion: userdb.skasproject.io/v1alpha1
kind: GroupBinding
metadata:
  name: luser1.ldemo-devs
  namespace: skas-system
spec:
  group: ldemo-devs
  user: luser1
```

## Session management

### View active sessions

En each user login, a token is generated. This token will expire after a delay of inactivity. (Like a Web session). This delay is 30mn by default.

On server side, the SKAS tokens are also stored as Kubernetes custom resources, in the namespace `skas-system`. 
And RBAC has been configured to allow access by any member of the `skas-admin` group.  

The SKAS tokens can be listed as any other kubernetes resources:

```shell
$ kubectl -n skas-system get tokens
NAME                                               CLIENT   USER LOGIN   AUTH.   USER ID   CREATION               LAST HIT
khrvvqwvpotcufiltvymuumsvrodsbiuwypbzrjiqudzjthg            admin        crd     0         2023-07-12T08:23:36Z   2023-07-12T08:32:13Z
ltdrlwnzzzhpxipgqgsvsaftmmucxxmfzhhwrdtuijabhvfd            luser1       crd     0         2023-07-12T08:27:19Z   2023-07-12T08:27:19Z
```

Each token represents an active user session. SKAS will remove it automatically after 30 minutes of inactivity by default.

Also, there is a maximum token duration, which is set to 12 hours by default.

A detailed view of each token can be displayed:

```shell
$ kubectl -n skas-system get tokens ltdrlwnzzzhpxipgqgsvsaftmmucxxmfzhhwrdtuijabhvfd -o yaml
apiVersion: session.skasproject.io/v1alpha1
kind: Token
metadata:
  creationTimestamp: "2023-07-12T08:27:19Z"
  generation: 1
  name: ltdrlwnzzzhpxipgqgsvsaftmmucxxmfzhhwrdtuijabhvfd
  namespace: skas-system
  resourceVersion: "513150"
  uid: 220471a2-2ec1-4b7f-af85-8647c4406343
spec:
  authority: crd
  client: ""
  creation: "2023-07-12T08:27:19Z"
  user:
    commonNames:
    - Local user1
    emails:
    - luser1@internal
    groups:
    - ldemo-devs
    login: luser1
    uid: 0
status:
  lastHit: "2023-07-12T08:27:19Z"
```

### Terminate session

To end a session, the corresponding token is to be deleted:

```shell
$ kubectl -n skas-system delete tokens ltdrlwnzzzhpxipgqgsvsaftmmucxxmfzhhwrdtuijabhvfd
token.session.skasproject.io "ltdrlwnzzzhpxipgqgsvsaftmmucxxmfzhhwrdtuijabhvfd" deleted
```

Note there is a local cache of 30 seconds on the client side. So the session will remains active on this short (and configurable) delay.

## Others `kubectl sk` commands

### hash

This command compute the Hash value of a password. It is intended to be used when creating a user through a manifest.

Note there is no password strength check doing this way.

### init

This command has been used at the beginning of this chapter. If you enter `kubectl sk init --help`, you can see there is some more options:

- Some are related to certificate management and was already mentioned.
- Some allow overriding of values provided by the server.
- `clientId/Secret` is an optional method to restrict access to this command to users provided with these information. To be configured on the server.

### login

Perform the login/pasword interaction. 

Will also allow providing login and password on the command line. 

### logout

Logout the user, by deleting locally cached token.

### password

To change current user password. 

To change the password of another user, use the `kubectl sk user patch` command. Of course, you need to be member of the group `skas-admin` to do so.

### whoami

Display the currently logged user and the groups its belong to.

### version

Display the current version of this SKAS plugin




