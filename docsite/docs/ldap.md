# LDAP Setup

## Overview 


SKAS supports the concept of an 'Identity Provider' (IDP). An Identity Provider can be defined as a database that hosts 
users and groups, providing user authentication and information.

Up to this point, we have used a single Identity Provider (IDP) based on resources stored in Kubernetes, 
within the `skas-system` namespace. SKAS also provides an alternative Identity Provider backed by an LDAP directory.

> _The SKAS LDAP identity provider code was heavily inspired by the LDAP connector of the [DEX project](https://github.com/dexidp/dex).
Many thanks to all of its contributors for their valuable contributions._

### Security considerations

SKAS attempts to bind with the underlying LDAP server using both the admin and user's plain text password. 
While some LDAP implementations permit the use of hashed passwords, SKAS does not support this feature. 
Instead, it strongly recommends that all administrators employ TLS (Transport Layer Security) for secure communication.

TLS can often be enabled by using port 636 instead of the default port 389 and configuring the appropriate certificate 
authority. This ensures that passwords are transmitted securely over the network.

SKAS currently permits insecure connections to ensure compatibility with a wide range of LDAP implementations and to 
simplify the initial setup process. However, it's important to note that these configurations should never be used 
in a production environment, as they can actively expose and compromise passwords, posing a significant security risk.

## LDAP Configuration

> _For a more comprehensive understanding of what this configuration entails, you can refer to the [Architecture Overview](./architecture.md#overview) documentation._

### The helm values file

LDAP activation and configuration will be performed by adding specific values during Helm (re)deployment. 

Here is a template of such values:

??? abstract "values.ldap.yaml"

    ```{.yaml .copy}
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
        rootCaPath:
        rootCaData:
    
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

- `skMerge` is the SKAS module aimed at merging information from several Identity Provider. This merging process follows 
specific rules, which are be described [here](./chaining.md).
- `skMerge.providers` is the ordered list of Identity Providers. 
- `crd` is the name of the provider managing the user database in the `skas-system` namespace.
- `ldap` is the name of our LDAP provider, which will be configured under the `skLdap` subsection.
- `skLdap.enabled` must be set to `true` (It is `false` by default). 
- `skLdap.ldap.*` is the LDAP configuration with all parameters and their description.

To apply this configuration, use the following command:

```{.shell .copy}
helm -n skas-system upgrade skas skas/skas --values ./values.init.yaml \
--values ./values.ldap.yaml
```

> _Don't forget to include the `values.init.yaml` file, or merge its content into the values.ldap.yaml file. 
Additionally, if you have other custom values files, make sure to include them in each `helm upgrade` command as well._

> _Remember to restart the SKAS pod(s) after applying any configuration changes. See [Configuration: Pod restart](configuration.md/#pod-restart)_

n this configuration, there are two sources of identity: our original `skas-system` user database and the newly added 
`ldap` server. Merging multiple sources of identity is an important aspect of SKAS identity management. 
The [next chapter](chaining.md) will delve into this topic.

After deployment, you can test your configuration using the `kubectl sk user describe <login> --explain` command. For more information, refer to the [next chapter](./chaining.md).

### Sample configurations

Here is a sample configuration for connecting to an OpenLDAP server:

??? abstract "values.openldap.yaml"

    ```
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
        rootCaData: "LS0tLS1CRUdJTiBDRVJUSUZ................................lRJRklDQVRFLS0tLS0K"
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
    ```

Note that, since the connection uses SSL, you need to provide the Certificate Authority of the LDAP Server. 
You can provide this CA in the field `skLdap.ldap.rootCaData` as a base64-encoded certificate file.

Here is a sample configuration for connecting to a FreeIPA LDAP server.

??? abstract "values.freeipa.yaml"

    ```
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
        rootCaData: "LS0tLS1CRUdJTiBDRU4zclBySE.........................JRklDQVRFLS0tLS0K"
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
    ```

Trick: To obtain the 'rootCaData' from a FreeIPA server, log onto this server and enter the following:

```{.shell .copy}
cd /etc/ipa && cat ca.crt  | base64 -w0
```

### Setup LDAP CA in a configMap

The `rootCaData` attribute can be a rather lengthy string, which can be troublesome. 
An alternative solution is to store it in a ConfigMap.

First, create the configMap with the CA file: 

```{.shell .copy}
kubectl -n skas-system create configmap ldap-ca.crt --from-file=./CA.crt
```

Then, modify the values file as follows:

??? abstract "values.ldap.yaml"

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
        host: ldap.mydomain.internal
        port: 636
        rootCaPath: /tmp/ca/ldap/CA.crt
        bindDN: ..........
    ```

- The `skLdap.extraConfigMaps` subsection instructs the POD to mount this configMap at the defined location. 
- The property `skLdap.ldap.rootCaPath` can now refer to the mounted value, and, of course, `skLdap.ldap.rootCaData`
should be removed.

You can apply this modification with the following command:

```{.shell .copy}
helm -n skas-system upgrade skas skas/skas --values ./values.init.yaml \
--values ./values.ldap.yaml
```

> _Remember to restart the SKAS pod(s) after applying any configuration changes. See [Configuration: Pod restart](configuration.md/#pod-restart)_