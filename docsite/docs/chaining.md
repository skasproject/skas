
# Identity Providers chaining

## Overview

In the previous chapter, a configuration has been set up with two Identity Providers:

```{.yaml}
skMerge:
  providers:
    - name: crd
    - name: ldap
```

The `crd` provider refers to the user database stored in the `skas-namespace` while the `ldap` refers to a connected LDAP server.

The function of the `skMerge` module is to unify this chain of providers, allowing them to function as a single entity.. 

By default, user information is consolidated in the following manner:

- If a given user exists in only one provider, that provider is considered the authoritative source for that user.

- If a given user exists in several providers:
    - The resulting group set is the union of all groups from all providers hosting this user.
    - The resulting email set is the union of all emails from all providers hosting this user.
    - The resulting commonName set is the union of all commonNames from all providers hosting this user.
    - The first provider in the chain hosting this user will be the authoritative one for password validation.
        - This means that there can't be two valid passwords for a single user.
        - This also implies that the order of providers in the list is important.
        - There is one exception to this rule: If a user has no password defined (this is a valid case for our `crd` provider), then the authoritative one is the next provider in the list.
    - The UID will be defined by the authoritative provider. (The one who validate the password)

## CLI user management

Of course, all `kubectl sk user ...` operation such as `create`, `patch`, `bind/unbind` can only modify resources in 
the `crd` provider. They have no impact on `ldap` or other external provider.

> _From the SKAS perspective, LDAP is 'Read-Only'._

A specific `kubectl sk user describe` subcommand will display consolidated information for any user. For example:

```
$ kubectl sk user describe jsmith
> USER     STATUS              UID      GROUPS             EMAILS                                          COMMON NAMES   AUTH
> jsmith   passwordUnchecked   100001   devs,itdep,staff   john.smith@mycompany.com,jsmith@mycompany.com   John SMITH     crd
```

Note the last column, which indicates the authoritative provider for each user.

> _Access to this subcommand is restricted to members of the `skas-admin` group._

The flag `--explain` will help you understand where user's information is sourced from:

```
$ kubectl sk user describe jsmith --explain
> USER     STATUS              UID      GROUPS             EMAILS                                          COMMON NAMES   AUTH
> jsmith   passwordUnchecked   100001   devs,itdep,staff   john.smith@mycompany.com,jsmith@mycompany.com   John SMITH     crd

> Detail:
> PROVIDER   STATUS              UID          GROUPS        EMAILS                     COMMON NAMES
> crd        passwordUnchecked   100001       devs          jsmith@mycompany.com       John SMITH
> ldap       passwordUnchecked   1148400004   staff,itdep   john.smith@mycompany.com   John SMITH
```

There are also two flags (`--password` or `ìnputPassword`) for the administrator to validate a password, if they know it:

```
$ kubectl sk user describe jsmith --explain --inputPassword
> Password for user 'jsmith':
> USER     STATUS            UID      GROUPS             EMAILS                                          COMMON NAMES   AUTH
> jsmith   passwordChecked   100001   devs,itdep,staff   john.smith@mycompany.com,jsmith@mycompany.com   John SMITH     crd

> Detail:
> PROVIDER   STATUS            UID          GROUPS        EMAILS                     COMMON NAMES
> crd        passwordChecked   100001       devs          jsmith@mycompany.com       John SMITH
> ldap       passwordFail      1148400004   staff,itdep   john.smith@mycompany.com   John SMITH
```

## Group bindings

In the [User guide](userguide.md#skas-group-binding), it has been explained how to bind a group to a user from the `crd` provider. 
This capability is also possible for any user, regardless of their provider. 

For example, let's say we have a user `oriley` in the LDAP server (although not defined in our `crd` provider):"

```
$ kubectl sk user describe oriley --explain
> USER     STATUS              UID          GROUPS        EMAILS                 COMMON NAMES   AUTH
> oriley   passwordUnchecked   1148400003   itdep,staff   oriley@mycompany.com   Oliver RILEY   ldap

> Detail:
> PROVIDER   STATUS              UID          GROUPS        EMAILS                 COMMON NAMES
> crd        userNotFound        0
> ldap       passwordUnchecked   1148400003   staff,itdep   oriley@mycompany.com   Oliver RILEY
```

Let's say we want this user to be able to be an admin for SKAS and also for the Kubernetes cluster. For this, we need to set up two GroupBindings:

```
$ kubectl sk user bind oriley system:masters
> GroupBinding 'oriley.system.masters' created in namespace 'skas-system'.

$ kubectl sk user bind oriley skas-admin
> GroupBinding 'oriley.skas-admin' created in namespace 'skas-system'.

$ $ kubectl sk user describe oriley --explain
> USER     STATUS              UID          GROUPS                                  EMAILS                 COMMON NAMES   AUTH
> oriley   passwordUnchecked   1148400003   itdep,skas-admin,staff,system:masters   oriley@mycompany.com   Oliver RILEY   ldap

> Detail:
> PROVIDER   STATUS              UID          GROUPS                      EMAILS                 COMMON NAMES
> crd        userNotFound        0            system:masters,skas-admin
> ldap       passwordUnchecked   1148400003   staff,itdep                 oriley@mycompany.com   Oliver RILEY
```

Of course, this group binding could have been performed on the LDAP server. However, this would require having some 
write access on the LDAP server. It is often considered a best practice to manage cluster authorization at the cluster
level. (We will explore a way to centralize authorization in a multi-cluster context later on).

## Role binding

As it is possible to bind a group to a user defined in whatever provider, it is also possible to bind a Kubernetes 
`role` (or `clusterRole`) to a group defined in the LDAP provider:

```
$ kubectl -n ldemo create rolebinding configurator-itdep --role=configurator --group=itdep
> rolebinding.rbac.authorization.k8s.io/configurator-itdep created
```

> _See the [User guide](userguide.md#manage-user-groups-and-permissions) for a sample, including the `configurator` role definition._

## Provider configuration.

Up to this point, the configuration has defined the provider chain as follows:

```{.yaml}
skMerge:
  providers:
    - name: crd
    - name: ldap
```

Each provider can support optional attributes. Here is a snippet with all the attributes and their default values:

```{.yaml .copy}
skMerge:
  providers:
    - name: crd
      credentialAuthority: true
      groupAuthority: true
      critical: true
      groupPattern: "%s"
      uidOffset: 0
    - name: ldap
      credentialAuthority: true
      groupAuthority: true
      critical: true
      groupPattern: "%s"
      uidOffset: 0
```

- `credentialAuthority`:  Setting this attribute to 'false' will prevent this provider from authenticating any user.
- `groupAuthority`: Setting this attribute to `false` will prevent the groups of this provider from being added to each user.
- `critical`: Defines the behavior of the chain if this provider is down or out of order (e.g., LDAP server is down). 
If set, then all authentication will fail in such a case.
- `groupPattern`: Allows you to 'decorate' all groups provided by this provider. See the example below.
- `uidOffset`: This will be added to the UID value if this provider is the authority for this user.

For example:

```{.yaml .copy}
skMerge:
  providers:
    - name: crd
      credentialAuthority: false
      groupAuthority: true
      critical: true
      groupPattern: "%s"
      uidOffset: 0
    - name: ldap
      credentialAuthority: true
      groupAuthority: true
      critical: true
      groupPattern: "dep1_%s"
      uidOffset: 0
```

The `crd` provider will not be able to authenticate any user (`credentialAuthority` is set to `false`). This means we have 'lost' our initial `admin` user.

Fortunately, we previously granted `oriley` with full admin rights. 

```
$ kubectl sk login oriley
> Password:
> logged successfully..

$ kubectl sk whoami
> USER     ID           GROUPS
> oriley   1148400003   dep1_itdep,dep1_staff,skas-admin,system:masters
```

We can check here that this user still belong to the kubernetes admin groups (`skas-admin`, `system:masters`) but 
the groups of the `ldap` provider has been renamed with the `dep1_` prefix.

We can see here that this user still belongs to the Kubernetes admin groups (`skas-admin` and `system:masters`), 
but the groups from the `ldap` provider have been prefixed with  `dep1_`.

Let's take a closer look at this user:

```
$ kubectl sk user describe oriley --explain
> USER     STATUS              UID          GROUPS                                            EMAILS                 COMMON NAMES   AUTH
> oriley   passwordUnchecked   1148400003   dep1_itdep,dep1_staff,skas-admin,system:masters   oriley@mycompany.com   Oliver RILEY   ldap

> Detail:
> PROVIDER   STATUS              UID          GROUPS                      EMAILS                 COMMON NAMES
> crd        userNotFound        0            system:masters,skas-admin
> ldap       passwordUnchecked   1148400003   staff,itdep                 oriley@mycompany.com   Oliver RILEY
```

Now, let's check the `admin` user:

```
$ kubectl sk user describe admin --explain
> USER    STATUS            UID   GROUPS                      EMAILS   COMMON NAMES         AUTH
> admin   passwordMissing   0     skas-admin,system:masters            SKAS administrator

> Detail:
> PROVIDER   STATUS            UID   GROUPS                      EMAILS   COMMON NAMES
> crd        passwordMissing   0     skas-admin,system:masters            SKAS administrator
> ldap       userNotFound      0
```

The fact that we denied `credentialAuthority` will translate to `passwordMissing` (While, in fact, the password is still physically present in the storage.) 

Such a configuration aims to comply with certain overarching management policies:

- A corporate policy that requires all users to be referenced in a central LDAP server. This constraint is fulfilled, 
as even though a user can still be created in the`crd` provider, their corresponding credentials will not be activated.
- As Kubernetes cluster administrators, we want to have exclusive control over who can manage the cluster. By adding 
a group decorator (`groupPattern: "dep1_%s"`), we prevent a malicious LDAP administrator from granting access to
critical groups (`skas-admin`, `system:master`, ...) to any LDAP users."

Two complementary remarks:

- There may still be an interest in creating a user in the `crd` provider. This is done to add more information, such as an email or commonName, to a user existing in the `ldap` provider.
- The `crd` provider must be the first in the list. Otherwise, an LDAP administrator may create a user with the same name as an existing administrator and gain authority over its password, to gain full Kubernetes access.