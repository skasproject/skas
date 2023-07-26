# LDAP Setup

## Overview 

SKAS support the concept of 'Identity Provider'. An Identity provider (IDP) can be defined as a Database hosting 
users and groups and providing user authentication and information. 

Up to now, we have used a single IDP based on resources stored in Kubernetes, in the namespace `skas-system`.

SKAS provide another Identity Provider backed by an LDAP directory

> _The SKAS LDAP identity provider was strongly inspired from the LDAP connector of [DEX project](https://github.com/dexidp/dex).
Great thanks to all its contributors._

### Security considerations

SKAS  attempts to bind with the backing LDAP server using the admin and end user's plain text password.
Though some LDAP implementations allow passing hashed passwords, SKAS doesn't support hashing and instead strongly recommends that all administrators just use TLS.
This can often be achieved by using port 636 instead of 389, and by handling certificate authority.

SKAS currently allows insecure connections to ensure connectivity with the wide variety of LDAP implementations and to ease initial setup.
But such configuration should never be used in a production context, as they are actively leaking passwords.

## LDAP Configuration

### The helm values file

LDAP activation and configuration can be performed by adding specifics values during Helm (re)deployment.

Here is a template of such values: 

```yaml
skMerge:
  providers:
    - name: crd
    - name: ldap

skLdap:
  enabled: true

  # --------------------------------- LDAP configuration
  ldap:
    # The host and port of the LDAP server.
    # If port isn't supplied, it will be guessed based on the TLS configuration. 389 or 636.
    host:
    port:

    # Timeout on connection to ldap server. Default to 10
    timeoutSec: 10

    # Required if LDAP host does not use TLS.
    insecureNoSSL: false

    # Don't verify the CA.
    insecureSkipVerify: false

    # Connect to the insecure port then issue a StartTLS command to negotiate a
    # secure connection. If not supplied secure connections will use the LDAPS protocol.
    startTLS: false

    # Path to a trusted root certificate file, or Base64 encoded PEM data containing root CAs.
    rootCA:
    rootCAData:

    # If server require client authentication with certificate.
    #  Path to a client cert file and a private key file
    clientCert:
    clientKey:

    # BindDN and BindPW for an application service account. The connector uses these
    # credentials to search for users and groups.
    bindDN:
    bindPW:

    userSearch:
      # BaseDN to start the search from. For example "cn=users,dc=example,dc=com"
      baseDN:
      # Optional filter to apply when searching the directory. For example "(objectClass=person)"
      filter:
      # Attribute to match against the login. This will be translated and combined
      # with the other filter as "(<loginAttr>=<login>)".
      loginAttr:
      #  Can either be:
      # * "sub" - search the whole sub tree (Default)
      # * "one" - only search one level
      scope: "sub"
      # The attribute providing the numerical user ID
      numericalIdAttr:
      # The attribute providing the user's email
      emailAttr:
      # The attribute providing the user's common name
      cnAttr:

    groupSearch:
      # BaseDN to start the search from. For example "cn=groups,dc=example,dc=com"
      baseDN:
      # Optional filter to apply when searching the directory. For example "(objectClass=posixGroup)"
      filter: (objectClass=posixgroup)
      # Defaults to "sub"
      scope: "sub"
      # The attribute of the group that represents its name.
      nameAttr: cn
      # The filter for group/user relationship will be: (<linkGroupAttr>=<Value of LinkUserAttr for the user>)
      # If there is several value for LinkUserAttr, we will loop on.
      linkGroupAttr:
      linkUserAttr:
```

- `skMerge` is the SKAS module aimed to merge information from several Identity Provider. This merge obey to some rules which will be described later.
- `skMerge.providers` is the ordered list of Identity Providers. 
- `crd` is the name of the provider managing the user database in the `skas-system` namespace.
- `ldap` is the name of our LDAP provider, which will be configured under the `skLdap` subsection.
- `skLdap.enabled` must be set to `true` (It is `false` by default).
- `skLdap.ldap.*` is the LDAP configuration with all parameters and their description.

To apply this configuration:

```shell
$ helm -n skas-system upgrade skas https://github.com/skasproject/skas/releases/download/0.2.1/skas-0.2.1.tgz \
--values ./values.init.yaml --values --values ./values.ldap.yaml
```

> _Don't forget to add the `values.init.yaml`, or to merge it in the `values.ldap.yaml` file. Also, if you have others values file, they must be added on each upgrade_

In this configuration, there is two source of identity: Our original `skas-system` user database and the newly added `ldap` server. 
How these two sources are merged is the object of the next chapter. 

### Sample configurations

Here is a sample of configuration, aimed to connect to an OpenLDAP server

```
$ cat >./values.ldap.yaml <<"EOF"
skMerge:
  providers:
    - name: crd
    - name: ldap

skLdap:
  enabled: true
  # --------------------------------- LDAP configuration
  ldap:
    host: ldap.mydomain.internal
    insecureNoSSL: false
    rootCAData: "LS0tLS1CRUdJTiBDRVJUSUZ................................lRJRklDQVRFLS0tLS0K"
    bindDN: cn=Manager,dc=mydomain,dc=internal
    bindPW: admin123
    groupSearch:
      baseDN: ou=Groups,dc=mydomain,dc=internal
      filter: (objectClass=posixgroup)
      linkGroupAttr: memberUid
      linkUserAttr: uid
      nameAttr: cn
    timeoutSec: 10
    userSearch:
      baseDN: ou=Users,dc=mydomain,dc=internal
      cnAttr: cn
      emailAttr: mail
      filter: (objectClass=inetOrgPerson)
      loginAttr: uid
      numericalIdAttr: uidNumber
EOF
```

Note that, as the connection is using SSL, there is a need to provide a Certificate Authority. 
Such CA is provided here in `skLdap.ldap.rootCAData`, as a base64 encoded certificate file.

And here is a sample of configuration, aimed to connect to an FreeIPA LDAP server

```
$ cat >./values.ldap.yaml <<"EOF"
skMerge:
  providers:
    - name: crd
    - name: ldap

skLdap:
  enabled: true
  # --------------------------------- LDAP configuration
  ldap:
    host: ipa1.mydomain.internal
    port: 636
    rootCAData: "LS0tLS1CRUdJTiBDRU4zclBySE.........................JRklDQVRFLS0tLS0K"
    bindDN: uid=admin,cn=users,cn=accounts,dc=mydomain,dc=internal
    bindPW: ipaadmin
    userSearch:
      baseDN: cn=users,cn=accounts,dc=mydomain,dc=internal
      emailAttr: mail
      filter: (objectClass=inetOrgPerson)
      loginAttr: uid
      numericalIdAttr: uidNumber
      cnAttr: cn
    groupSearch:
      baseDN: cn=groups,cn=accounts,dc=mydomain,dc=internal
      filter: (objectClass=posixgroup)
      linkGroupAttr: member
      linkUserAttr: DN
      nameAttr: cn
EOF
```

Trick: To get the `rootCAData` from a FreeIPA server, log on this server and:

```
$ cd /etc/ipa
$ cat ca.crt  | base64 -w0
```


### Setup LDAP CA in a configMap

The `rootCAData` attribute could be a quite long string, which can be troublesome. 
An alternate solution is to store this in a configMap.

First, create the configMap with the CA file: 

```
kubectl -n skas-system create configmap ldap-ca.crt --from-file=./CA.crt
```

Then modify the values file as the following:

```
skMerge:
  providers:
    - name: crd
    - name: ldap

skLdap:
  enabled: true

  extraConfigMaps:
    - configMap: ldap-ca.crt
      volume: ldap-ca
      mountPath: /tmp/ca/ldap

  # --------------------------------- LDAP configuration
  ldap:
    host: ldap.ops.scw01
    insecureNoSSL: false
    rootCA: /tmp/ca/ldap/CA.crt
    .....
```

The `skLdap.extraConfigMaps` subsection instruct the POD to mount this configMap to the defined location. The property
`skLdap.ldap.rootCA` can now refer to the mounted value. Of course `skLdap.ldap.rootCAData` should be removed.



