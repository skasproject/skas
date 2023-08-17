
# Delegated users management

> _As the two configurations are quite similar, there is a lot of redundancy between this chapter and [Two LDAP servers configuration](/twoldapservers) chapter_

Aim of this configuration is the ability to delegate the management of a certain set of users and/or their group bindings.

But, we want to restrict the rights of the administrator of the delegated space. Especially , we don't want them to be able to promote themself to global system administrator.

![ldap](./images/draw5.png){ align=right width=350}

In this sample configuration, we will setup a separate user database for a department 'dep1'.

To achieve this, the solution is to create a specific namespace `dep1-userdb` which will host `skusers` and `groupBinding` SKAS resources

And to handle this namespace, we need to instantiate a second Identity Provider, of type `skCrd`.

For the reasons described in [Two LDAP servers configuration](/twoldapservers), we need to instantiate this POD as a separate Helm deployment, although using the same Helm Chart

This configuration requires two steps:

- Setup a new Helm deployment for `skCrd2` pod.
- Reconfigure the `skMerge` module of the main SKAS pod to connect to this new IDP.

![](./images/empty.png){width=700}

In the following, three variants of this configuration will be described. One with the connection in clear text, and two secured, with network encryption and inter-pod authentication.

## Clear text connection

### Auxiliary POD configuration

Here is a sample values file to configure the auxiliary POD:

??? abstract "values.skCrd2.yaml"

    ``` { .yaml .copy } 
    skAuth:
      enabled: false
    skMerge:
      enabled: false
    skLdap:
      enabled: false
    
    skCrd:
      enabled: true
      namespace: dep1-userdb
    
      adminGroups:
        - dep1-admin
    
      initialUser:
        login: dep1-admin
        passwordHash: $2a$10$ijE4zPB2nf49KhVzVJRJE.GPYBiSgnsAHM04YkBluNaB3Vy8Cwv.G  # admin
        commonNames: ["DEP1 administrator"]
        groups:
          - admin
    
      # By default, only internal (localhost) server is activated, to be called by another container running in the same pod.
      # Optionally, another server (external) can be activated, which can be accessed through a kubernetes service
      # In such case:
      # - A Client list should be provided to control access.
      # - ssl: true is strongly recommended.
      # - And protection against BFA should be activated (protected: true)
      exposure:
        internal:
          enabled: false
        external:
          enabled: true
          port: 7112
          ssl: false
          services:
            identity:
              disabled: false
              clients:
                - id: "*"
                  secret: "*"
              protected: true
    ```

- At the beginning of the file, we disable all other modules than `skCrd`.

- `skCrd.namespace:  dep1-userdb`  define the namespace this IDP will use to manage the user's information.

- Then, we define one adminGroup: `dep1-admin`. The Helm chart will setup RBAC rules to allow members of this group to access SKAS users resources in the namespace set above. 

- Then, we create an initial admin user `dep1-admin`, belonging to the group `admin`. More on this later.

Then, there is the `exposure` part, who define how this service will be exposed. (The default configuration is expose to `localhost` in clean text)

- `exposure.internal.enabled: false` shutdown the HTTP server bound on localhost.
- `exposure.external.enabled: true` set the HTTP server bound on the POD IP up. This on port 7112 with no SSL.
- Then the is the configuration of the `identity` service to expose:
    - `clients[]` is a mechanism to validate who is able to access this service, by providing an `id` and a `secret`
      (or password). the values "*" will disable this feature.
    - `protected: true` activate an internal mechanism against attacks of type 'Brut force', by introducing delays on
      unsuccessful connection attempts, and by limiting the number of simultaneous connections. There is no reason to
      disable it, except if you suspect a misbehavior.

To deploy this configuration:

```shell
helm -n skas-system install skas2 skas/skas --values ./values.skCrd2.yaml
```

> **Note the `skas2' release name**

The Helm chart will deploy the new pod(s), under the name `skas2`. it will also deploy an associated Kubernetes `service`.

### Main pod reconfiguration

Second step is to reconfigure the main POD

Here is a sample of appropriate configuration:

??? abstract "values.main.yaml"

    ``` { .yaml .copy }
    skMerge:
      providers:
        - name: crd
        - name: crd_dep1
          groupPattern: "dep1-%s"
    
      providerInfo:
        crd:
          url: http://localhost:7012
        crd_dep1:
          url: http://skas2-crd.skas-system.svc
    ```


There is two entries aimed to configure a provider on the bottom of the `skMerge` module:

- `providers` is a list of the connected providers, which allow to define their behavior. The order is important here. Note the `groupPattern: "dep1-%s"`
  Refers to the [IDP chaining: Provider configuration](chaining.md#provider-configuration) chapter for more information.
- `providerInfo` is a map providing information on how to reach these providers.<br>For `crd`, we use the
  default `localhost` port.<br>For `crd_dep1` we use the service created by the `skas2` deployment.

The link between these two entries is of course the provider name.

Then, the reconfiguration must be applied:

```shell
$ helm -n skas-system upgrade skas skas/skas --values ./values.init.yaml \
--values ./values.main.yaml
```

> _Don't forget to add the `values.init.yaml`, or to merge it in the `values.main.yaml` file. Also, if you have others values file, they must be added on each upgrade_

> _And don't forget to restart the pod(s). See [Configuration: Pod restart](/configuration#pod-restart)_

If deploying two separate Charts is a constraint for you, you may setup a 'meta chart'. See [here](/toolsandtricks#tricks-setup-a-meta-helm-chart)

## Test and Usage

Then, you can now test your configuration:

You can login using the dep1-admin user that has been created by the previous deployment:

```shell
$ kubectl sk login
Login:dep1-admin
Password:
logged successfully..
```

>  _Password is `admin`. Of course, to change ASAP (`kubectl sk password`)_

Now, look at what our account look like:

```shell
$ kubectl sk whoami
USER         ID   GROUPS
dep1-admin   0    dep1-admin
```

Note the group: `dep1-admin`. The prefix `dep1-` has been added, as for any group of this provider. 
This is the result of  `groupPattern: "dep1-%s"` configuration above.

### User management

As `dep1-admin`, you can manage users in the namespace `dep1-userdb`:

```shell
$ kubectl sk -n dep1-userdb user create fred --commonName "Fred Astair" --password "GtaunPMgP5f"
User 'fred' created in namespace 'dep1-userdb'.

$ kubectl sk -n dep1-userdb user bind fred managers
GroupBinding 'fred.managers' created in namespace 'dep1-userdb'.

$ kubectl -n dep1-userdb get skusers
NAME         COMMON NAMES             EMAILS   UID   COMMENT   DISABLED
dep1-admin   ["DEP1 administrator"]
fred         ["Fred Astair"]                                   false

$ kubectl -n dep1-userdb get groupbindings
NAME               USER         GROUP
dep1-admin-admin   dep1-admin   admin
fred.managers      fred         managers
```

Then you can test the user 'fred'

```shell
$ kubectl sk login
Login:fred
Password:
logged successfully..

$ kubectl sk whoami
USER   ID   GROUPS
fred   0    dep1-managers
```

> _Note the group prefixed by `dep1-`. This will ensure no user managed by this identity provider can belong to some strategic groups such as `skas-admin` or `system:masters`.

Now, logged back to `dep1-admin`, ensure we are limited to our namespace:

```shell

$ kubectl sk login dep1-admin
Password:
logged successfully..

$ kubectl -n skas-system get skusers
Error from server (Forbidden): users.userdb.skasproject.io is forbidden: User "dep1-admin" cannot list resource "users" in API group "userdb.skasproject.io" in the namespace "skas-system"

$ kubectl get --all-namespaces skusers
Error from server (Forbidden): users.userdb.skasproject.io is forbidden: User "dep1-admin" cannot list resource "users" in API group "userdb.skasproject.io" at the cluster scope
``` 

The `sk user describe` subcommand is also unauthorized, as it is, by definition, a cross provider feature.

```shell
$ kubectl sk user describe fred
Unauthorized!
``` 

### Default namespace

Providing the namespace to each command can be tedious. It can be set as the default one for both `kubectl` and `kubectl-sk` subcommands:

```shell
kubectl config set-context --current --namespace=dep1-userdb
``` 

Then:

```shell
$ kubectl get skusers
NAME         COMMON NAMES             EMAILS   UID   COMMENT   DISABLED
dep1-admin   ["DEP1 administrator"]
fred         ["Fred Astair"]                                   false
``` 

Or it can be set with an environment variable:

```shell
export SKAS_NAMESPACE="dep1-userdb"
``` 

**But this last method will apply only on `kubectl-sk` subcommands**

### User describe

As stated above, a `dev1-admin` user is not allowed use the `kubectl sk user describe` subcommand.

This control is performed by the `skAuth` module, with a list of allowed groups. 
Here is a modified version of the values file which allow `dep1-admin` members to perform a user describe subcommand. 

??? abstract "values.main.yaml"

    ``` { .yaml .copy }
    skAuth:
      # Members of these group will be allowed to perform 'kubectl_sk user describe'
      # Also, they will be granted by RBAC to access token resources
      adminGroups:
        - skas-admin
        - dep1-admin

    skMerge:
      providers:
        - name: crd
        - name: crd_dep1
          groupPattern: "dep1-%s"
    
      providerInfo:
        crd:
          url: http://localhost:7012
        crd_dep1:
          url: http://skas2-crd.skas-system.svc
    ```

**As stated in the comment, these users will also be able to view and delete session tokens.**

So, be conscious than setting a group as an admin for the `skAuth` module will allow some operation out of the strict 
perimeter defined be the namespace `dep1-userdb`. Anyway, members of such groups will still be prevented from listing or 
editing users and groupBinding out of their initial namespace. 

### The SKAS admin user

And what about the initial SKAS global `admin` user. Is it  able to manager also the `dep1-userdb` database ?

It's depend if you have bind this user to the `system:master` group. If so, it will have full cluster access:

```shell
$ kubectl get -n dep1-userdb skusers
NAME         COMMON NAMES             EMAILS   UID   COMMENT   DISABLED
dep1-admin   ["DEP1 administrator"]
fred         ["Fred Astair"]                                   false
```

You can remove this binding (and logout/login):

```shell
$ kubectl sk whoami
USER    ID   GROUPS
admin   0    skas-admin,system:masters

$ kubectl sk user unbind admin system:masters
GroupBinding 'admin.system.masters' in namespace 'skas-system' has been deleted.

$ kubectl sk logout
Bye!

$ kubectl sk login admin
Password:
logged successfully..

$ kubectl sk whoami
USER    ID   GROUPS
admin   0    skas-admin
```

Now, access should be denied 

```shell
$ kubectl get -n dep1-userdb skusers
Error from server (Forbidden): users.userdb.skasproject.io is forbidden: User "admin" cannot list resource "users" in API group "userdb.skasproject.io" in the namespace "dep1-userdb"
```

But, as SKAS admin, you can promote yourself as a member of the `dep1-admin` group.

```shell
$ kubectl sk user bind admin dep1-admin
GroupBinding 'admin.dep1-admin' created in namespace 'skas-system'.

$ kubectl sk logout
Bye!

$ kubectl sk login
Login:admin
Password:
logged successfully..

$ kubectl sk whoami
USER    ID   GROUPS
admin   0    dep1-admin,skas-admin

$ kubectl get -n dep1-userdb skusers
NAME         COMMON NAMES             EMAILS   UID   COMMENT   DISABLED
dep1-admin   ["DEP1 administrator"]
fred         ["Fred Astair"]                                   false
```

## Securing connection

It should ne noted than unencrypted passwords will transit through the link between the two pods. So, setting up encryption is a must have.

### Auxiliary POD configuration

Here is the modified version for the `skCrd2` pod configuration:

??? abstract "values.skCrd2.yaml"

    ``` { .yaml .copy } 
    skAuth:
      enabled: false
    skMerge:
      enabled: false
    skLdap:
      enabled: false

    clusterIssuer: cluster-issuer1

    skCrd:
      enabled: true
      namespace: dep1-userdb
    
      adminGroups:
        - dep1-admin
    
      initialUser:
        login: dep1-admin
        passwordHash: $2a$10$ijE4zPB2nf49KhVzVJRJE.GPYBiSgnsAHM04YkBluNaB3Vy8Cwv.G  # admin
        commonNames: ["DEP1 administrator"]
        groups:
          - admin
    
      # By default, only internal (localhost) server is activated, to be called by another container running in the same pod.
      # Optionally, another server (external) can be activated, which can be accessed through a kubernetes service
      # In such case:
      # - A Client list should be provided to control access.
      # - ssl: true is strongly recommended.
      # - And protection against BFA should be activated (protected: true)
      exposure:
        internal:
          enabled: false
        external:
          enabled: true
          port: 7112
          ssl: true
          services:
            identity:
              disabled: false
              clients:
                - id: "skMerge"
                  secret: "aSharedSecret"
              protected: true
    ```


The differences are the following:

- There is a `clusterIssuer` definition to be able to generate a certificate. (It is assumed here than `cert-manager` is deployed in the cluster)
- `exposure.external.ssl` is set to `true`. This will also leads the generation of the server certificate.
- The `service.identity.clients` authentication is also activated. The `id` and `secret` values will have to be provided by the `skMerge` client.

To deploy this configuration:

```shell
helm -n skas-system install skas2 skas/skas --values ./values.skCrd2.yaml
```

> **Note the `skas2' release name**

The Helm chart will deploy the new pod(s), under the name `skas2`. it will also deploy an associated Kubernetes `service`
and the `cert-manager.io/v1/Certificate` request.

### Main pod reconfiguration

Here is the modified version for the main SKAS POD configuration:

??? abstract "values.main.yaml"

    ``` { .yaml .copy }
    skMerge:
      providers:
        - name: crd
        - name: crd_dep1
          groupPattern: "dep1-%s"
    
      providerInfo:
        crd:
          url: http://localhost:7012
        crd_dep1:
          url: https://skas2-crd.skas-system.svc
          rootCaPath: /tmp/cert/skas2/ca.crt
          insecureSkipVerify: false
          clientAuth:
            id: skMerge
            secret: aSharedSecret
    
      extraSecrets:
        - secret: skas2-crd-cert
          volume: skas2-cert
          mountPath: /tmp/cert/skas2
    ```

The `providerInfo.crd_dep1` has been modified for SSL and authenticated connection:

- `url` begins with `https`.
- `clientAuth` provides information to authenticated against the `skCrd2` pod.
- `insecureSkipVerify` is set to false, as we want to check certificate validity.
- `rootCaPath` is set to access the `ca.crt`, the CA validating the `skCrd2` server certificate.

As stated above, during the deployment of the `skCrd2` auxiliary POD, a server certificate has been generated to allow
SSL enabled services. This certificate is stored in a secret (of type `kubernetes.io/tls`) named `skas2-crd-cert`.
Alongside the private/public key pair, it also contains the root Certificate authority under the name`ca.crt`.

The `skMerge.extraSecrets` subsection instruct the POD to mount this secret to the defined location.
The property `skMerge.providerInfo.crd_dep1.rootCaPath` can now refer to the mounted value.

Then, the reconfiguration must be applied:

```shell
$ helm -n skas-system upgrade skas skas/skas --values ./values.init.yaml \
--values ./values.main.yaml
```

> _Don't forget to add the `values.init.yaml`, or to merge it in the `values.ldap.yaml` file. Also, if you have others values file, they must be added on each upgrade_

> _And don't forget to restart the pod(s). See [Configuration: Pod restart](configuration.md/#pod-restart)_

You can now test again your configuration, as [described above](#test-and-usage)

## Use a Kubernetes secrets

There is still a security issue, as the shared secret (`aSharedSecret`) is in clear text in both values file. As such it may ends up in some version control system.

The good practice will be to store the secret value in a kubernetes `secret` resource, such as:

``` { .yaml .copy }
---
apiVersion: v1
kind: Secret
metadata:
  name: skas2-client-secret
  namespace: skas-system
data:
  clientSecret: Sk1rbkNyYW5WV1YwR0E5
type: Opaque
```

Where `data.clientSecret` is the secret encoded in base 64.

> There is several solution to generate such secret value. One can use Helm with some random generator function. Or use a [Secret generator](/toolsandtricks#secret-generator)

### Auxiliary POD configuration

To use this secret, here is the new modified version for the `skCrd2` POD configuration:

??? abstract "values.skCrd2.yaml"

    ``` { .yaml .copy } 
    skAuth:
      enabled: false
    skMerge:
      enabled: false
    skLdap:
      enabled: false

    clusterIssuer: cluster-issuer1

    skCrd:
      enabled: true
      namespace: dep1-userdb
    
      adminGroups:
        - dep1-admin
    
      initialUser:
        login: dep1-admin
        passwordHash: $2a$10$ijE4zPB2nf49KhVzVJRJE.GPYBiSgnsAHM04YkBluNaB3Vy8Cwv.G  # admin
        commonNames: ["DEP1 administrator"]
        groups:
          - admin
    
      # By default, only internal (localhost) server is activated, to be called by another container running in the same pod.
      # Optionally, another server (external) can be activated, which can be accessed through a kubernetes service
      # In such case:
      # - A Client list should be provided to control access.
      # - ssl: true is strongly recommended.
      # - And protection against BFA should be activated (protected: true)
      exposure:
        internal:
          enabled: false
        external:
          enabled: true
          port: 7112
          ssl: true
          services:
            identity:
              disabled: false
              clients:
                - id: "skMerge"
                  secret: ${SKAS2_CLIENT_SECRET}
              protected: true
             
      extraEnv:
        - name: SKAS2_CLIENT_SECRET
          valueFrom:
            secretKeyRef:
              name: skas2-client-secret
              key: clientSecret
    ```


The modifications are the following:

- The `skLdap.extraEnv` subsection inject the secret value as an environment variable in the container.
- the `exposure.external.services.identity.clients[0].secret` fetch its value through this environment variable.

> Most of the values provided by the helm chart ends up inside a configMap, which is then loaded by the SKAS executable. The environment variable interpolation occurs during this load.

### Main pod reconfiguration

Here is the modified version, with `secret` handling, for the main SKAS pod configuration:

??? abstract "values.main.yaml"

    ``` { .yaml .copy }
    skMerge:
      providers:
        - name: crd
        - name: crd_dep1
          groupPattern: "dep1-%s"
    
      providerInfo:
        crd:
          url: http://localhost:7012
        crd_dep1:
          url: https://skas2-crd.skas-system.svc
          rootCaPath: /tmp/cert/skas2/ca.crt
          insecureSkipVerify: false
          clientAuth:
            id: skMerge
            secret: ${SKAS2_CLIENT_SECRET}

      extraEnv:
        - name: SKAS2_CLIENT_SECRET
          valueFrom:
            secretKeyRef:
              name: skas2-client-secret
              key: clientSecret
    
      extraSecrets:
        - secret: skas2-crd-cert
          volume: skas2-cert
          mountPath: /tmp/cert/skas2
    ```

The modifications are the same as the SKAS2 POD


