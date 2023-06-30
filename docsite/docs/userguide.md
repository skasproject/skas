
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

You may get ride of this error by providing the root CA certificate as a file:

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

You could list them, using standard kubectl commands:

```shell
$ kubectl -n skas-system get users.userdb.skasproject.io
Login:admin
Password:
NAME    COMMON NAMES             EMAILS   UID   COMMENT   DISABLED
admin   ["SKAS administrator"]
```

Several remarks:

- If you have configured your client as described above, you now have to be logged to perform any kubectl action. 
  So the login/password interaction
- Default password is `admin`. **DON'T FORGET TO CHANGE IT**. See below.
- The `admin` user has been granted to access SKAS resources in `skas-system` namespace using kubernetes RBAC
- `kubectl -n skas-system get users` will no works, as `users` refers to a standard kubernetes resources.

To ease SKAS user management, an alias has been defined: `skuser`

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

In fact, what has been granted to access SAKS resources is not the admin account (It could be), but a group named `skas-system`.

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

It is clear than we are authenticated as `admin`, but this account has no permissions to perform cluster-wide operation.

Such permissions can be granted by binding this user to a group having such rights:

```
$ kubectl sk user bind admin system:masters
GroupBinding 'admin.system.masters' created in namespace 'skas-system'.
```

For this to be effective, logout and login back:

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
luser3                                                                  false
```

### Modify user

A subcommand `patch` is provided to modify a user. As an example:

```shell
$ kubectl sk user patch luser1 --state=disabled
User 'luser1' updated in namespace 'skas-system'.

$ kubectl -n skas-system get skuser luser1
NAME     COMMON NAMES      EMAILS                UID   COMMENT   DISABLED
luser1   ["Local user1"]   ["luser1@internal"]                   true
```

Most of the options are the same as the `user create` subcommand. 

There is also a `--create` option which will allow user creation if it does not exists.

### Delete user

Users can be deleted using standard `kubectl` commands:

```shell
$ kubectl -n skas-system delete skuser luser2
user.userdb.skasproject.io "luser2" deleted
```

### Manage user's groups

## Manifests users management

### User resources

### GroupBinding resources

## Session management


## Others `kubectl sk` commands

### hash

### init

### login

### logout

### password

### version

### whoami


``` 
---------------------------------------------------------------------------------------------------------------
```

# Argo cd
