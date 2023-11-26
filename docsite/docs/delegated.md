# Delegated users management

> _As the two configurations are quite similar, there is a lot of redundancy between this chapter and 
[Two LDAP servers configuration](twoldapservers.md) chapter_

Aim of this configuration is the ability to delegate the management of a certain set of users and/or their group bindings.

But, we want to restrict the rights of the administrator of the delegated space. Especially , we don't want them to be 
able to promote themself to global system administrator.

![ldap](images/draw5.png){ align=right width=350}

In this sample configuration, we will set up a separate user database for a department 'dep1'.

To achieve this, the solution is to create a specific namespace, `dep1-userdb`, which will host `skusers` and
`groupBinding` SKAS resources.

To manage this namespace, we need to instantiate a second Identity Provider of type `skCrd`.

For the reasons described in [Two LDAP servers configuration](twoldapservers.md), we need to instantiate this POD as 
a separate Helm deployment, although using the same Helm Chart.

This configuration requires two steps:

- Set up a new Helm deployment for the `skas2` pod.
- Reconfigure the `skMerge` module of the main SKAS pod to connect to this new IDP.

![](./images/empty.png){width=700}

In the following, three variants of this configuration will be described: one with the connection in clear text and 
two secured options with network encryption and inter-pod authentication.

> _It is suggested that, even if your goal is to achieve a fully secured configuration, you begin by implementing the 
unsecured, simplest variant first. Then, you can incrementally modify it as described._

## Clear text connection

### Secondary POD configuration

Here is a sample values file for configuring the secondary POD:

??? abstract "values.skas2.yaml"

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

- At the beginning of the file, we disable all modules except `skCrd`.

- `skCrd.namespace: dep1-userdb` defines the namespace that this IDP will use to manage user information.

- Then, we define an adminGroup: `dep1-admin`. The Helm chart will set up RBAC rules to allow members of this group 
to access SKAS user resources in the specified namespace.

- Next, we create an initial admin user, `dep1-admin`, who belongs to the group `admin` (not `dep1-admin`. More on this later).

After that, there is the `exposure` section, which defines how this service will be exposed. (The default 
configuration exposes it to `localhost` in clear text).

- `exposure.internal.enabled: false` disables the HTTP server bound to localhost.
- `exposure.external.enabled: true` enables the HTTP server bound to the POD IP, on port 7112. SSL is disabled for 
this unsecure configuration.
- Then, there is the configuration of the `identity` service to expose:
    - `clients[]` is a mechanism for validating who can access this service by providing an id and a secret
      (or password). The value "*" will disable this feature.
    - `protected: true` activates an internal mechanism against brute force attacks by introducing delays on
      unsuccessful connection attempts and limiting the number of simultaneous connections. There is no reason to
      disable it unless you suspect misbehavior.

To deploy this configuration, execute:

```{.shell .copy}
helm -n skas-system install skas2 skas/skas --values ./values.skas2.yaml
```

> **Note the `skas2' release name**

The Helm chart will deploy the new pod(s), under the name `skas2`. it will also deploy an associated Kubernetes `service`.

### Main pod reconfiguration

The second step is to reconfigure the main pod. Here is a sample of the appropriate configuration:

??? abstract "values.skas.yaml"

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

- `providers` is a list of the connected providers, which allow you to define their behavior. For more information, refer to the
  [Identity Provider chaining: Provider configuration](chaining.md#provider-configuration) chapter.<br>The order of 
  providers in this list is important.
  > Note also the `groupPattern: "dep1-%s"`.

- `providerInfo` is a map that provides information on how to reach these providers.<br>For `crd`, we use the
  default `localhost` port.<br>For `crd_dep1` we use the service created by the `skas2` deployment.

The link between these two entries is of course the provider name.

> The crd provider must be the first in the list. Otherwise, an administrator of `dep1` may create a user with the same name 
as an existing administrator and gain authority over its password, to gain full Kubernetes access.

The reconfiguration must then be applied by executing this command:

```{.shell .copy}
helm -n skas-system upgrade skas skas/skas --values ./values.init.yaml --values ./values.skas.yaml
```

> _Don't forget to include the `values.init.yaml` file or merge it into the `values.skas.yaml` file. Additionally,
if you have other values files, make sure to include them in each upgrade._

> _Also, remember to restart the pod(s) after making these configuration changes. You can find more information on how
to do this in the [Configuration: Pod restart](configuration.md#pod-restart) section._

## Test and Usage

Now, you can test your configuration. You can log in using the `dep1-admin` user created in the previous deployment:"

```shell
$ kubectl sk login
> Login:dep1-admin
> Password:
> logged successfully..
```

>  _Please note that the password for this user is set to 'admin,' and it's highly recommended to change it as soon 
as possible using the `kubectl sk password` command._

Now, let's take a look at what our account looks like:

```shell
$ kubectl sk whoami
> USER         ID   GROUPS
> dep1-admin   0    dep1-admin
```

Note the group `dep1-admin`, which includes the prefix `dep1-` as configured in the `groupPattern: "dep1-%s"` setting 
above.

### User management

As `dep1-admin`, you have management access to users in the `dep1-userdb` namespace:

```shell
$ kubectl sk -n dep1-userdb user create fred --commonName "Fred Astair" --password "GtaunPMgP5f"
> User 'fred' created in namespace 'dep1-userdb'.

$ kubectl sk -n dep1-userdb user bind fred managers
> GroupBinding 'fred.managers' created in namespace 'dep1-userdb'.

$ kubectl -n dep1-userdb get skusers
> NAME         COMMON NAMES             EMAILS   UID   COMMENT   DISABLED
> dep1-admin   ["DEP1 administrator"]
> fred         ["Fred Astair"]                                   false

$ kubectl -n dep1-userdb get groupbindings
> NAME               USER         GROUP
> dep1-admin-admin   dep1-admin   admin
> fred.managers      fred         managers
```

Then you can test the user 'fred'

```shell
$ kubectl sk login
> Login:fred
> Password:
> logged successfully..

$ kubectl sk whoami
> USER   ID   GROUPS
> fred   0    dep1-managers
```

> _Note the group prefixed by `dep1-`. This will ensure no user managed by this identity provider can belong to some 
strategic groups such as `skas-admin` or `system:masters`.

Now, log back in as `dep1-admin` to ensure that we are limited to our namespace.

```shell

$ kubectl sk login dep1-admin
> Password:
> logged successfully..

$ kubectl -n skas-system get skusers
> Error from server (Forbidden): users.userdb.skasproject.io is forbidden: User "dep1-admin" cannot list resource "users" in API group "userdb.skasproject.io" in the namespace "skas-system"

$ kubectl get --all-namespaces skusers
> Error from server (Forbidden): users.userdb.skasproject.io is forbidden: User "dep1-admin" cannot list resource "users" in API group "userdb.skasproject.io" at the cluster scope
``` 

The `sk user describe` subcommand is also unauthorized because it is a cross-provider feature.

```shell
$ kubectl sk user describe fred
> Unauthorized!
``` 

### Default namespace

Providing the namespace for each command can be tedious. It can be set as the default for both `kubectl` and 
`kubectl-sk` subcommands:"

```{.shell .copy}
kubectl config set-context --current --namespace=dep1-userdb
```

Then:

```shell
$ kubectl get skusers
> NAME         COMMON NAMES             EMAILS   UID   COMMENT   DISABLED
> dep1-admin   ["DEP1 administrator"]
> fred         ["Fred Astair"]                                   false
``` 

Alternatively, it can be set using an environment variable

```{.shell .copy}
export SKAS_NAMESPACE="dep1-userdb"
``` 

**But this last method will only apply on `kubectl-sk` subcommands**

### User describe

As stated above, a `dep1-admin` user is not allowed to use the `kubectl sk user describe` subcommand.

This control is performed by the `skAuth` module, with a list of allowed groups. Here is a modified version of the values 
file that allows `dep1-admin` members to perform a user describe subcommand.

??? abstract "values.skas.yaml"

    ``` { .yaml .copy }
    skAuth:
      # Members of these group will be allowed to perform 'kubectl-sk user describe'
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

**As mentioned in the comment, these users will also be able to view and delete session tokens.**

So, please be aware that setting a group as an admin for the `skAuth` module will allow certain operations outside the strict
perimeter defined by the `dep1-userdb` namespace. However, members of such groups will still be prevented from listing or
editing users and groupBindings outside of their initial namespace.

### The SKAS admin user

And what about the initial SKAS global `admin` user? Is it able to manage the `dep1-userdb` database as well?

It depends on whether you have assigned this user to the `system:master` group. If you have, it will have full cluster access:

```shell
$ kubectl get -n dep1-userdb skusers
> NAME         COMMON NAMES             EMAILS   UID   COMMENT   DISABLED
> dep1-admin   ["DEP1 administrator"]
> fred         ["Fred Astair"]                                   false
```

You can remove this binding (then logout/login):

```shell
$ kubectl sk whoami
> USER    ID   GROUPS
> admin   0    skas-admin,system:masters

$ kubectl sk user unbind admin system:masters
> GroupBinding 'admin.system.masters' in namespace 'skas-system' has been deleted.

$ kubectl sk logout
> Bye!

$ kubectl sk login admin
> Password:
> logged successfully..

$ kubectl sk whoami
> USER    ID   GROUPS
> admin   0    skas-admin
```

Now, access should be denied:

```shell
$ kubectl get -n dep1-userdb skusers
> Error from server (Forbidden): users.userdb.skasproject.io is forbidden: User "admin" cannot list resource "users" in API group "userdb.skasproject.io" in the namespace "dep1-userdb"
```

But, as the SKAS admin, you can promote yourself as a member of the `dep1-admin` group.

```shell
$ kubectl sk user bind admin dep1-admin
> GroupBinding 'admin.dep1-admin' created in namespace 'skas-system'.

$ kubectl sk logout
> Bye!

$ kubectl sk login
> Login:admin
> Password:
> logged successfully..

$ kubectl sk whoami
> USER    ID   GROUPS
> admin   0    dep1-admin,skas-admin

$ kubectl get -n dep1-userdb skusers
> NAME         COMMON NAMES             EMAILS   UID   COMMENT   DISABLED
> dep1-admin   ["DEP1 administrator"]
> fred         ["Fred Astair"]                                   false
```

## Securing connection

It should be noted that unencrypted passwords will transit through the link between the two pods. Therefore, setting up encryption is a must-have.

### Secondary POD configuration

Here is the modified version for the `skas2` pod configuration:

??? abstract "values.skas2.yaml"

    ``` { .yaml .copy } 
    skAuth:
      enabled: false
    skMerge:
      enabled: false
    skLdap:
      enabled: false

    clusterIssuer: your-cluster-issuer

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

- There is a `clusterIssuer` definition to enable the generation of a certificate. (It is assumed here that
  `cert-manager` is deployed in the cluster).
- `exposure.external.ssl` is set to `true`. This will also lead to the generation of the server certificate.
- The `service.identity.clients` authentication is also activated. The `id` and `secret` values will have to be
  provided by the `skMerge` client.

To deploy this configuration:

```{.shell .copy}
helm -n skas-system install skas2 skas/skas --values ./values.skas2.yaml
```

> **Note the `skas2' release name**

The Helm chart will deploy the new pod(s) with the name `skas2`. It will also deploy an associated Kubernetes `service`
and submit a `cert-manager.io/v1/Certificate` request.

### Main pod reconfiguration

Here is the modified version for the main SKAS POD configuration:

??? abstract "values.skas.yaml"

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
- `clientAuth` provides information to authenticate against the `skas2` pod.
- `insecureSkipVerify` is set to `false`, as we want to check certificate validity.
- `rootCaPath` is set to access the `ca.crt`, the CA validating the `skas2` server certificate.

As stated above, during the deployment of the `skas2` secondary POD, a server certificate has been generated to allow
SSL enabled services. This certificate is stored in a secret (of type `kubernetes.io/tls`) named `skas2-crd-cert`.
Alongside the private/public key pair, it also contains the root Certificate Authority under the name`ca.crt`.

The `skMerge.extraSecrets` subsection instructs the POD to mount this secret at the defined location.
The property `skMerge.providerInfo.crd_dep1.rootCaPath` can now reference the mounted value.

Then, the reconfiguration must be applied:

```{.shell .copy}
helm -n skas-system upgrade skas skas/skas --values ./values.init.yaml --values ./values.skas.yaml
```

> _Don't forget to include the `values.init.yaml` file or merge it into the `values.skas.yaml` file. Additionally,
if you have other values files, make sure to include them in each upgrade._

> _Also, remember to restart the pod(s) after making these configuration changes. You can find more information on how
to do this in the [Configuration: Pod restart](configuration.md#pod-restart) section._

You can now test again your configuration, as [described above](#test-and-usage)

## Using a Kubernetes secrets

There is still a security issue because the shared secret (`aSharedSecret`) is stored in plain text in both values
files, which could lead to it being accidentally committed to a version control system.

The best practice is to store the secret value in a Kubernetes `secret` resource, like this:


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

Where `data.clientSecret` is the secret encoded in base64.

> There are several solutions to generate such a secret value. One can use Helm with a random generator function. Another one is to use a [Secret generator](toolsandtricks.md#secret-generator)."

### Secondary POD configuration

To use this secret, here is the new modified version of the `skas2` POD configuration:

??? abstract "values.skas2.yaml"

    ``` { .yaml .copy } 
    skAuth:
      enabled: false
    skMerge:
      enabled: false
    skLdap:
      enabled: false

    clusterIssuer: your-cluster-issuer

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

- The `skCrd.extraEnv` subsection injects the secret value as an environment variable into the container.
- the `skCrd.exposure.external.services.identity.clients[0].secret` retrieves its value through this environment variable.

> Most of the values provided by the Helm chart end up inside a ConfigMap, which is then loaded by the SKAS executable. 
Environment variable interpolation occurs during this loading process.


### Main pod reconfiguration

Here is the modified version of the main SKAS pod configuration, which incorporates `secret` handling:

??? abstract "values.skas.yaml"

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

The modifications for the `skMerge` module are the same as those made for the SKAS2 POD.

## Set up a meta helm chart

Up to this point, we have configured our deployment by performing two closely related Helm deployments.
To simplify automation, it can be helpful to create a 'meta chart,' a chart that includes other charts as dependencies.

Such a chart will have the following layout:

```shell
.
|-- Chart.yaml
|-- templates
|   `-- stringsecret.yaml
`-- values.yaml
```

> In this example, we will implement encryption and inter-pod authentication.

The `Chart.yaml` file defines the meta-chart `skas-skas2-meta`. It has two dependencies that deploy the same Helm
chart but with different values (as shown below). Please note the `alias: skas2` in the second deployment.

??? abstract "Chart.yaml"

    ``` { .yaml .copy }
    apiVersion: v2
    name: skas-skas2-meta
    version: 0.1.0
    dependencies:
    - name: skas
      version: 0.2.2
      repository: https://skasproject.github.io/skas-charts
    
    - name: skas
      alias: skas2
      version: 0.2.2
      repository: https://skasproject.github.io/skas-charts
    ```

The following manifest will generate the shared secret required for inter-pod authentication.

??? abstract " templates/stringsecret.yaml"

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

And here is the global `values.yaml` file:

??? abstract "values.yaml"

    ``` { .yaml .copy }
    skas:
      skAuth:
        exposure:
          external:
            ingress:
              host: skas.ingress.kspray6
        kubeconfig:
          context:
            name: skas@kspray6
          cluster:
            apiServerUrl: https://kubernetes.ingress.kspray6

      skMerge:
        providers:
          - name: crd
          - name: crd_dep1
            groupPattern: "dep1-%s"
    
        providerInfo:
          crd:
            url: http://localhost:7012
          crd_dep1:
            url: https://skas-skas2-crd.skas-system.svc # Was https://skas2-crd.skas-system.svc
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
          - secret: skas-skas2-crd-cert  # Was skas2-crd-cert
            volume: skas2-cert
            mountPath: /tmp/cert/skas2
    
    
    skas2:
      skAuth:
        enabled: false
      skMerge:
        enabled: false
      skLdap:
        enabled: false
    
      clusterIssuer: your-cluster-issuer
    
      skCrd:
        enabled: true
        namespace: dep1-userdb
    
        adminGroups:
          - dep1-admin
    
        initialUser:
          login: dep1-admin
          passwordHash: $2a$10$ijE4zPB2nf49KhVzVJRJE.GPYBiSgnsAHM04YkBluNaB3Vy8Cwv.G  # admin
          commonNames: [ "DEP1 administrator" ]
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

There are two blocks: `skas` and `skas2`, matching the name or alias in the `Chart.yaml` file.

These two blocks hold the same definitions as the ones defined in the original configuration, with two differences:

- `skas.skMerge.providerInfo.crd_dep1.url: https://skas-skas2-crd.skas-system.svc`
- `skas.skMerge.extraSecrets[0].secret: skas-skas2-crd-cert`

This is to accommodate service and secret name changes due to aliasing of the second dependency.

Then, to launch the deployment, in the same folder as `Chart.yaml`, execute:

```{.shell .copy}
helm dependency build && helm -n skas-system upgrade -i skas .
```
