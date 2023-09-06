
# User guide

## Local client configuration

For this guide, we assume that kubectl is already installed. If it's not, you can refer to the [official Kubernetes documentation](https://kubernetes.io/docs/tasks/tools/) for installation instructions.

We also assume that you've installed the `kubectl-sk` CLI extension as outlined in the [installation guide](installation.md#installation-of-skas-cli).

To access a Kubernetes cluster using kubectl, you need a configuration file. By default, this file is located in `<homedir>/.kube/config`.

SKAS provides a mechanism to create or update this user's configuration file, simplifying the setup process.

```{.shell .copy}
kubectl sk init https://skas.ingress.mycluster.internal
```
```
Setup new context 'skas@mycluster.internal' in kubeconfig file '/Users/john/.kube/config'
```

You can verify that this new context is now set as the current one:

```{.shell .copy}
kubectl config current-context
```
```
skas@mycluster.internal
```


### Encountering Certificate Issues?

If your system doesn't have the CA certificate that was used to certify SKAS (refer to the `clusterIssuer` parameter 
during the initial installation), you may encounter an error similar to the following:

```shell
ERRO[0000] error on GET kubeconfig from remote server  
 error="error on http connection: Get \"https://skas.ingress.mycluster.internal/v1/kubeconfig\": 
 tls: failed to verify certificate: x509: certificate signed by unknown authority"
```

You can resolve this error by providing the root CA certificate as a file:

```{.shell .copy}
kubectl sk init https://skas.ingress.mycluster.internal --authRootCaPath=./CA.crt
```

Assuming you are a Kubernetes system administrator, here is how you can obtain the `CA.crt` file:

```{.shell .copy}
kubectl -n skas-system get secret skas-auth-cert -o=jsonpath='{.data.ca\.crt}' | base64 -d >./CA.crt
```

If you are unable to get such CA certificate, you can skip the test by setting a flag:

```{.shell .copy}
kubectl sk init --authInsecureSkipVerify=true https://skas.ingress.mycluster.internal
```

> _Using this flag should be limited to the initial evaluation context due to potential security risks, as the target site could be a fraudulent one._


## Basic usage

### Logging in with the Default Admin Account

SKAS manages a local user database where users are stored as Kubernetes resources. 

During installation, a default `admin` user with the password `admin` is created for convenience.

By default, SKAS users are stored in the namespace `skas-system`. You could list them, using standard kubectl commands:

```{.shell}
$ kubectl -n skas-system get users.userdb.skasproject.io
> Login:admin
> Password:
> NAME    COMMON NAMES             EMAILS   UID   COMMENT   DISABLED
> admin   ["SKAS administrator"]
```

If you have configured your client as described above, you must now be logged in to execute any kubectl action.
This involves the login and password interaction.

A few important points to note:

- The default password is `admin`. **It's crucial to change it for obvious security reasons**. See the instructions below.
- The `admin` user has been granted access to SKAS resources in the `skas-system` namespace using Kubernetes RBAC.
- The command `kubectl -n skas-system get users` might not work as expected, as users is also a standard Kubernetes resource.

To simplify SKAS user management, an alias `skuser` has been defined.

```{.shell}
$ kubectl -n skas-system get skusers
> NAME    COMMON NAMES             EMAILS   UID   COMMENT   DISABLED
> admin   ["SKAS administrator"]
```

Please note that there is no longer a login/password interaction. Instead, a token was provided during the first login. 
This token will expire after a period of inactivity, similar to a web session. The default inactivity timeout is 30 minutes.

### Logging Out and Logging In

Once you are logged in, you can use `kubectl` as you normally would. The token will be transparently used until it 
expires due to inactivity.

If the token expires, you will be prompted to enter your login and password again.

You can also log out at any time by using the following command:

```{.shell .copy}
$ kubectl sk logout
```

Then, you will be prompted again for your login and password when you run the next `kubectl`command.

Please note the `sk` subcommand, which instructs `kubectl` to forward the command to the `kubectl-sk` extension

Alternatively, you can also use explicit login:

```{.shell}
$ kubectl sk login
> Login:admin
> Password:
> logged successfully..
```

or

```{.shell}
$ kubectl sk login admin
> Password:
> logged successfully..
```

or

```{.shell}
$ kubectl sk login admin ${ADMIN_PASSWORD}
> logged successfully..
```

> _Running `sk login` will first perform an `sk logout` if you are currently logged in._

### Password change

As previously mentioned, it's essential to change the password of this account for security reasons. 
Here's how you can do it:

```{.shell}
$ kubectl sk password
> Will change password for user 'admin'
> Old password:
> New password:
> Confirm new password:
> Password has been changed sucessfully.
```

Please note the use of `sk` as this command is executed by the SKAS kubectl extension.

There is a password strength check in place, so you may receive a response like this:

```{.shell .copy}
$ kubectl sk password
> Will change password for user 'admin'
> Old password:
> New password:
> Confirm new password:
> Unsatisfactory password strength!
```

The password criteria do not follow specific rules such as length or special character requirements. 
Instead, an algorithm assigns a score to the password, and this score must meet a minimum (configurable) value. 
Additionally, there is a check against a list of commonly used passwords.

The simplest way to meet these criteria is to increase the length of your password.


### SKAS group binding

In reality, access to SKAS resources is granted not to the `admin` account (although it could be), but to a group named `skas-system`.

The user `admin` has been included in the group `skas-system` through another SKAS resource named `groupbindings.userdb.skasproject.io`, with `groupbindings` serving as an alias.

```shell
$ kubectl -n skas-system get groupBindings
> NAME               USER    GROUP
> admin-skas-admin   admin   skas-admin
```

> _In Kubernetes, a group doesn't exist as a concrete resource; it only exists as a reference used in RBAC `roleBinding` or `clusterRoleBindings`, or in SKAS `groupBindings`._

### Becoming a Cluster Administrator

Let's attempt the following:

```shell
$ kubectl get namespaces
> Error from server (Forbidden): namespaces is forbidden: User "admin" cannot list resource "namespaces" in API group "" at the cluster scope
```

It is clear that we have successfully authenticated as `admin`. However, this account does not possess the necessary permissions to execute cluster-wide operations.

To gain these permissions, we must associate this user with a group that has the required rights:

```shell
$ kubectl sk user bind admin system:masters
> GroupBinding 'admin.system.masters' created in namespace 'skas-system'.
```

To make this effective, please log out and then log back in:

```shell
$ kubectl sk logout
> Bye!

$ kubectl get namespaces
> Login:admin
> Password:
> NAME              STATUS   AGE
> cert-manager      Active   4d21h
> default           Active   4d21h
> ingress-nginx     Active   4d21h
> .....
```

You can verify the updated list of `groupBindings`:

```shell
$ kubectl -n skas-system get groupBindings
> NAME                   USER    GROUP
> admin-skas-admin       admin   skas-admin
> admin.system.masters   admin   system:masters
```

**WARNING: This implies that any member of the `skas-admin` group can elevate their privileges to become a full cluster
administrator. In reality, anyone with the capability to create or modify resources in the `skas-admin` namespace 
can potentially take control of the entire cluster. Therefore, access to this namespace must be rigorously managed 
and restricted.**  

You can refer to Advanced [Delegated User Management](delegated.md) to learn how to delegate certain aspects of user management without compromising cluster security.

### An issue with `stdin`

If you issue a `kubectl` command that use `stdin` as input, you may encounter the following error message:

```shell
$ cat mymanifest.yaml | kubectl apply -f -
> Login:
> Unable to access stdin to input login. Try login with `kubectl sk login' or 'kubectl-sk login'.` and issue this command again

> Unable to connect to the server: getting credentials: exec: executable kubectl-sk failed with exit code 18
```

This issue arises when your token has expired, and there is a conflict in using `stdin` for entering your login/password.

The solution is to make sure you are logged in before executing such a command:

```shell
$ kubectl sk login
> Login:oriley
> Password:
> logged successfully..

$ cat mymanifest.yaml | kubectl apply -f -
> pod/mypod created
```


## CLI users management

The SKAS kubectl extension plugin offers a `user` command with several subcommands

You can obtain a complete list of these subcommands by running:

```{ .shell}
$ kubectl sk user --help
> Skas user management
>
> Usage:
>   kubectl-sk user [command]
>
> Available Commands:
> .....
```

> _To use this subcommand, you must be logged in as a member of the `skas-admin` group._

### Create a new user

Here  is an example of the user creation process:

```shell
$ kubectl sk user create luser1 --commonName "Local user1" --email "luser1@internal" --password "RtVksSuMgP5f"
> User 'luser1' created in namespace 'skas-system'.
```

The only mandatory parameters is the user's name:

```shell
$ kubectl sk user create luser2
> User 'luser2' created in namespace 'skas-system'.
```

> _Since no password is provided, it will be impossible for this user to log in._

You can display a complete list of user creation options by running:

```
$ kubectl sk user create --help
> Create a new user
> 
> Usage:
>   kubectl-sk user create <user> [flags]
> 
> Flags:
>       --comment string        User's comment
>       --commonName string     User's common name
>       --email string          User's email
>       --generatePassword      Generate and display a password
>   -h, --help                  help for create
>       --inputPassword         Interactive password request
>   -n, --namespace string      User's DB namespace (default "skas-system")
>       --password string       User's password
>       --passwordHash string   User's password hash (Result of 'kubectl skas hash')
>       --state string          User's state (enabled|disabled) (default "enabled")
>       --uid int               User's UID
```

Many of the options correspond to a user's properties.

- The `comment`, `commonName`, `email`, `uid` parameters are purely descriptive and draw inspiration from Unix user attributes.
- The `state` parameter will enable the temporary disabling of a user account.

The `--namespace` option permits the storage of user resources in a different namespace. 
Refer to [Delegated User Management](delegated.md) for more details.

There are several options related to the password:

- `--password`: The password is supplied as a parameter on the command line.
- `--inputPassword`: : This prompts the user for input with `Password:` / `Confirm password`: interaction.
- `--generatePassword`: A random password is generated and displayed.
- `--passwordHash`: This option allows you to provide the hash of the password, as it will be stored in the resource. 
  You can use `kubectl sk hash` command to generate this value. Please note that using this method bypasses the 
  password strength check.

### List users

You can list users using standard `kubectl` commands:

```shell
$ kubectl -n skas-system get skusers
> NAME     COMMON NAMES             EMAILS                UID   COMMENT   DISABLED
> admin    ["SKAS administrator"]
> luser1   ["Local user1"]          ["luser1@internal"]                   false
> luser2                                                                  false
```

### Modify user

There's a `patch` subcommand available for modifying a user. Here's an example:

```shell
$ kubectl sk user patch luser2 --state=disabled
> User 'luser2' updated in namespace 'skas-system'.

$ kubectl -n skas-system get skuser luser2
> NAME     COMMON NAMES   EMAILS   UID   COMMENT   DISABLED
> luser2                                           true
```

Most of the options are the same as the `user create` subcommand. 

Additionally, there is a `--create` option that allows user creation if it does not already exist.

### Delete user

You can delete users using standard `kubectl` commands:

```shell
$ kubectl -n skas-system delete skuser luser2
> user.userdb.skasproject.io "luser2" deleted
```

### Manage user groups and permissions.

To illustrate how SKAS interact with Kubernetes RBAC, we will setup a simple example. We will:

- Create a namespace named `ldemo`.
- Create a role named `configurator` in this namespace to manage resources of type `configMaps`.
- Create a `roleBinding` between this role and a group named `ldemo-devs`.
- Add the user `luser1` to this group.

We assume we are logged as `admin` to perform theses tasks:

```shell
$ kubectl create namespace ldemo
> namespace/ldemo created

$ kubectl -n ldemo create role configurator --verb='*' --resource=configMaps
> role.rbac.authorization.k8s.io/configurator created

$ kubectl -n ldemo create rolebinding configurator-ldemo-devs --role=configurator --group=ldemo-devs
> rolebinding.rbac.authorization.k8s.io/configurator-ldemo-devs created

$ kubectl sk user bind luser1 ldemo-devs
> GroupBinding 'luser1.ldemo-devs' created in namespace 'skas-system'.
```

Now, we can proceed with testing. First, log out and then log in as `luser1`:"

```shell
$ kubectl sk logout
> Bye!

$ kubectl sk login
> Login:luser1
> Password:
> logged successfully..

$ kubectl sk whoami
> USER     ID   GROUPS
> luser1   0    ldemo-devs
```

Now, let's ensure that we can create a `configMap` and view it:

```shell
$ kubectl -n ldemo create configmap my-config --from-literal=key1=config1
> configmap/my-config created

$ kubectl -n ldemo get configmaps my-config -o yaml
> apiVersion: v1
> data:
>   key1: config1
> kind: ConfigMap
> metadata:
>   creationTimestamp: "2023-07-11T14:56:27Z"
>   name: my-config
>   namespace: ldemo
>   resourceVersion: "257983"
>   uid: ad55b282-9803-4688-b2df-a1c35f708313
```

Also, make sure that we can delete it:

```shell
$ kubectl -n ldemo delete configmap my-config
> configmap "my-config" deleted
```

> _Please, note than `roles` and `roleBindings` are namespaced resources while `users` and `groups` are cluster-wide resources._

### Kubernetes RBAC referential integrity

Kubernetes does not check referential integrity when creating a resource that references another one. 
For example, the following will work:

```shell
kubectl -n ldemo create rolebinding missing-integrity --role=unexisting-role --group=unexisting-group
> rolebinding.rbac.authorization.k8s.io/missing-integrity created

$ kubectl sk user bind unexisting-user unexisting-group
> GroupBinding 'unexisting-user.unexisting-group' created in namespace 'skas-system'.
```

Maybe the referenced resource will be created later, or the link will be useless.

This is evidently a design choice in Kubernetes, and SKAS follows the same logic.

## Using Manifests instead of the CLI

As users and groups are defined as Kubernetes custom resources, they can be created and managed using manifests.

By default, all SKAS user and group resources are stored in the skas-system namespace.

Kubernetes RBAC has been configured during installation to allow members of the `skas-admin` group to manage these resources.

### User resources

Here is the manifest for the users we previously created:

```{.yaml .copy}
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
- The password is stored in a non-reversible hashed form. You can compute such a hash using the `kubectl sk hash` command.

Below is a sample user with all properties defined:

```{.yaml .copy}
---
apiVersion: userdb.skasproject.io/v1alpha1
kind: User
metadata:
  name: jsmith
  namespace: skas-system
spec:
  commonNames:  
    - John SMITH
  passwordHash: $2a$10$lnweus6Oe3/XMoRaIImnVOwmxZ.xMp7iRB3X1TOcszzHE8nxfiwJK  # Password: "Xderghy12"
  emails: 
    - jsmith@mycompany.com
  uid: 100001
  comment: A sample user
  disabled: false 
```

To define a user, save the YAML definition in a file and execute the `kubectl apply -f <fileName>` command.


### GroupBinding resources

The SKAS `GroupBinding` resources can also be defined using manifest:

```{.yaml .copy}
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

A token is generated for each user login, and it will expire after a period of inactivity, similar to a web session. 
By default, this expiration period is set to 30 minutes.
Additionally, there is a maximum token duration, which is set to 12 hours by default.

On the server side, SKAS tokens are also stored as Kubernetes custom resources in the skas-system namespace. 
RBAC has been configured to grant access to these resources to any member of the skas-admin group.

SKAS tokens can be listed just like any other Kubernetes resources:

```shell
$ kubectl -n skas-system get tokens
> NAME                                               CLIENT   USER LOGIN   AUTH.   USER ID   CREATION               LAST HIT
> khrvvqwvpotcufiltvymuumsvrodsbiuwypbzrjiqudzjthg            admin        crd     0         2023-07-12T08:23:36Z   2023-07-12T08:32:13Z
> ltdrlwnzzzhpxipgqgsvsaftmmucxxmfzhhwrdtuijabhvfd            luser1       crd     0         2023-07-12T08:27:19Z   2023-07-12T08:27:19Z
```

Each token represents an active user session, and SKAS will automatically remove it after 30 minutes of inactivity by default.

You can view the details of each token in a detailed manner:

```shell
$ kubectl -n skas-system get tokens ltdrlwnzzzhpxipgqgsvsaftmmucxxmfzhhwrdtuijabhvfd -o yaml
> apiVersion: session.skasproject.io/v1alpha1
> kind: Token
> metadata:
>   creationTimestamp: "2023-07-12T08:27:19Z"
>   generation: 1
>   name: ltdrlwnzzzhpxipgqgsvsaftmmucxxmfzhhwrdtuijabhvfd
>   namespace: skas-system
>   resourceVersion: "513150"
>   uid: 220471a2-2ec1-4b7f-af85-8647c4406343
> spec:
>   authority: crd
>   client: ""
>   creation: "2023-07-12T08:27:19Z"
>   user:
>     commonNames:
>     - Local user1
>     emails:
>     - luser1@internal
>     groups:
>     - ldemo-devs
>     login: luser1
>     uid: 0
> status:
>   lastHit: "2023-07-12T08:27:19Z"
```

### Terminate session

To end a session, you need to delete the corresponding token:

```shell
$ kubectl -n skas-system delete tokens ltdrlwnzzzhpxipgqgsvsaftmmucxxmfzhhwrdtuijabhvfd
> token.session.skasproject.io "ltdrlwnzzzhpxipgqgsvsaftmmucxxmfzhhwrdtuijabhvfd" deleted
```

Please note that there is a local cache of 30 seconds on the client side. So, the session will remain active for this short (and configurable) period even after the token is deleted.

## Others `kubectl sk` commands

### hash

This command computes the hash value of a password. It is intended to be used when creating a user through a manifest.
Please note that there is no password strength check when using this method.

### init

This command has been used at the beginning of this chapter. If you enter `kubectl sk init --help`, you can see there is some more options:

- Some are related to certificate management and was already mentioned.
- Some allow overriding of values provided by the server.
- `clientId/Secret` is an optional method to restrict access to this command to users provided with these information. To be configured on the server.


This command has been used at the beginning of this chapter. If you enter `kubectl sk init --help`, you can see that 
there are some more options:

- Some are related to certificate management and have already been mentioned.
- Some allow for overriding values provided by the server.
- `clientId/Secret` is an optional method to restrict access to this command to users provided with this information.
  This needs to be configured on the server.

### login

Perform the login/password interaction. You can also provide the login and password on the command line.

### logout

Log out the user by deleting the locally cached token.

### password

Change the current user password. 

If you want to change the password of another user, you can use the `kubectl sk user patch` command. However, 
please note that you need to be a member of the `skas-admin` group to perform this action.

### whoami

Display the currently logged-in user and the groups to which it belongs.

### version

Display the current version of this SKAS plugin

## What to provide to other Kubernetes users

Here is a small checklist of what to provide to non-admin users to allow them to use kubectl on a SKAS enabled cluster.

- Obviously, instructions to install `kubectl`.
- Instructions to install `kubectl-sk`
- If needed, the `CA.crt` certificate file
- The `kubectl sk init https://skas.....` command line.
- The namespace(s) they are allowed to access

About this last point: You can instruct them to add the `--namespaceOverride` option on `kubectl sk init ...` command.
This will define the provided namespace as the default one in the `~/.kube/config` file.


Here is a small checklist of what to provide to non-admin users to allow them to use kubectl on a SKAS enabled cluster:

- Installation instructions for `kubectl`.
- Installation instructions for `kubectl-sk`.
- The `CA.crt` certificate file (if needed).
- The `kubectl sk init https://skas.....` command line.
- The namespace(s) they are allowed to access.
- 
Regarding the last point, you can instruct them to add the `--namespaceOverride` option to the `kubectl sk init ...` 
command. This will set the provided namespace as the default one in the `~/.kube/config` file.



