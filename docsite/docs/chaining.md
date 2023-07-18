
# Identity Providers chaining

## Overview

In the previous chapter, a configuration has been setup with two Identity Providers:

```
skMerge:
  providers:
    - name: crd
    - name: ldap
```

The `crd` provider refers to the user database stored in the `skas-namespace` while the `ldap` refers to a connected LDAP server.

The function of the `merge' module is to make this chain of provider acting as a single one. 

By default, user information are consolidated the following way: 

- If a given user exits only in one provider, this one is the authoritative on.

- If a given user exists in several providers:
    - The resulting group set is the union of all groups of all providers hosting this user.
    - The resulting email set is the union of all emails of all providers hosting this user.
    - The resulting commonName set is the union of all commonNames of all providers hosting this user.
    - The first provider hosting this user in the chain will be the authoritative one for the password validation.
        - This imply there can't be two valid passwords for a single user.
        - This also imply providers order is important in the list.
        - There is an exception on this rule: If a user has no password defined (This is a valid case for our 
          `crd` provider), then the authoritative one is the next in the list. 
    - The UID will be defined by the authoritative provider.


## CLI user management

Of course, all `kubctl sk user ...` operation such as `create`, `patch`, `bind/unbind` can only modify resources in the `crd` provider. They have no impact on `ldap` or other external provider.

> _From the SKAS perspective, LDAP is 'Read Only'_

For user member of the `skas-admin` group, there is a `kubectl sk user describe...` subcommand. This will display 
consolidated information for any user. For example:

```
$ kubectl sk user describe jsmith
USER     STATUS              UID      GROUPS             EMAILS                                          COMMON NAMES   AUTH
jsmith   passwordUnchecked   100001   devs,itdep,staff   john.smith@mycompany.com,jsmith@mycompany.com   John SMITH     crd
```

Note the last column, which indicate which provider is authoritative for this user.

The flag `--explain` will allow to understand from where user's information are sourced:

```
$ kubectl sk user describe jsmith --explain
USER     STATUS              UID      GROUPS             EMAILS                                          COMMON NAMES   AUTH
jsmith   passwordUnchecked   100001   devs,itdep,staff   john.smith@mycompany.com,jsmith@mycompany.com   John SMITH     crd

Detail:
PROVIDER   STATUS              UID          GROUPS        EMAILS                     COMMON NAMES
crd        passwordUnchecked   100001       devs          jsmith@mycompany.com       John SMITH
ldap       passwordUnchecked   1148400004   staff,itdep   john.smith@mycompany.com   John SMITH
```

There is also two flags (`--password` or `ìnputPassword`) for the administrator to validate a password, if it know it:

```
$ kubectl sk user describe jsmith --explain --inputPassword
Password for user 'jsmith':
USER     STATUS            UID      GROUPS             EMAILS                                          COMMON NAMES   AUTH
jsmith   passwordChecked   100001   devs,itdep,staff   john.smith@mycompany.com,jsmith@mycompany.com   John SMITH     crd

Detail:
PROVIDER   STATUS            UID          GROUPS        EMAILS                     COMMON NAMES
crd        passwordChecked   100001       devs          jsmith@mycompany.com       John SMITH
ldap       passwordFail      1148400004   staff,itdep   john.smith@mycompany.com   John SMITH
```

## Group bindings

In the Admin guide, it as been explained how to bind a group to a user from the `crd` provider.
This capability is also possible for any user, whatever his provider is. 

For example, let say we have a user `oriley` in the LDAP server (While not defined in our `crd` provider):

```
$ kubectl sk user describe oriley --explain
USER     STATUS              UID          GROUPS        EMAILS                 COMMON NAMES   AUTH
oriley   passwordUnchecked   1148400003   itdep,staff   oriley@mycompany.com   Oliver RILEY   ldap

Detail:
PROVIDER   STATUS              UID          GROUPS        EMAILS                 COMMON NAMES
crd        userNotFound        0
ldap       passwordUnchecked   1148400003   staff,itdep   oriley@mycompany.com   Oliver RILEY
```

Let's say we want this user to able to manage both the cluster. For this, we need to setup a single GroupBinding:

```
$ kubectl sk user bind oriley system:masters
GroupBinding 'oriley.system.masters' created in namespace 'skas-system'.

$ kubectl sk user describe oriley --explain
USER     STATUS              UID          GROUPS                       EMAILS                 COMMON NAMES   AUTH
oriley   passwordUnchecked   1148400003   itdep,staff,system:masters   oriley@mycompany.com   Oliver RILEY   ldap

Detail:
PROVIDER   STATUS              UID          GROUPS           EMAILS                 COMMON NAMES
crd        userNotFound        0            system:masters
ldap       passwordUnchecked   1148400003   staff,itdep      oriley@mycompany.com   Oliver RILEY

```

Of course, this group binding could have been performed on the LDAP server. But this imply to have some Write access on it. 
And it could be a best practice to manage cluster authorization at cluster level. (We will see later a way to centralize authorization in a multi-clusters context). 


## Role binding

As it is possible to bind a group to a user defined in whatever provider, it is possible to bind a Kubernetes `role` (or `clusterRole`) to a group defined in the LDAP provider:

```
$ kubectl -n ldemo create rolebinding configurator-itdep --role=configurator --group=itdep
rolebinding.rbac.authorization.k8s.io/configurator-itdep created
```

(See the Admin Guide for the `configurator` role definition)

## Advanced configuration.

Up to now, in the configuration, the providers chain has been defined as below:

```
skMerge:
  providers:
    - name: crd
    - name: ldap
```

But each provider can support some optional attributes. Here is the complete 

```
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



