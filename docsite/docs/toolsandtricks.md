# Tools and Tricks

## Secret generator

As stated in [Advanced configuration](advancedconfiguration.md#use-a-kubernetes-secrets), there is the need to generate 
a random secret in the deployment. For this, one can use [kubernetes-secret-generator](https://github.com/mittwald/kubernetes-secret-generator),
a custom kubernetes controller.

Here is a manifest which, once applied, will create the secret `ldap2-client-secret` used the authenticate the communication between the two PODs of the two LDAP configuration referenced above.  

``` { .yaml .copy }
---
apiVersion: "secretgenerator.mittwald.de/v1alpha1"
kind: "StringSecret"
metadata:
  name: ldap2-client-secret
  namespace: skas-system
spec:
  fields:
    - fieldName: "clientSecret"
      encoding: "base64"
      length: "15"
```

## k9s

## Kubernetes dashboard

## reloader

## Tricks: Setup a meta helm chart

In [Advanced configuration chapter](/advancedconfiguration), we had setup the appropriate configuration by performing two closely related Helm deployment.

To ease automation, it could be useful to 'package' such kind of deployment by creating a 'meta chart', a chart which will embed other ones as dependencies.

Such chart will have the following layout.

```shell
$ tree
.
|-- Chart.yaml
|-- templates
|   `-- stringsecret.yaml
`-- values.yaml
```

> This example is based on the 'two ldap servers' configuration, with encryption and inter-pod authentication activated.


The `Chart.yaml` file define the meta-chart `skas-ldap2-meta`.There is two dependencies deploying the same helm chart, 
but with different values (See below). Note the `alias: skas2` on the second deployment.

??? abstract "Chart.yaml"

    ``` { .yaml .copy }
    apiVersion: v2
    name: skas-ldap2-meta
    version: 0.1.0
    dependencies:
    - name: skas
      version: 0.2.1
      repository: https://skasproject.github.io/skas-charts
    
    - name: skas
      alias: skas2
      version: 0.2.1
      repository: https://skasproject.github.io/skas-charts
    ```

The following will generate the shared secret allowing inter-pods authentication 

??? abstract " templates/stringsecret.yaml"

    ``` { .yaml .copy }
    ---
    apiVersion: "secretgenerator.mittwald.de/v1alpha1"
    kind: "StringSecret"
    metadata:
      name: ldap2-client-secret
      namespace: skas-system
    spec:
      fields:
        - fieldName: "clientSecret"
          encoding: "base64"
          length: "15"
    ```

And here is the `values.yaml` file. There is two blocks: `skas` and `skas2`, matching the name or alias in the `Chart.yaml` file.

??? abstract "values.yaml"

    ``` { .yaml .copy }
    # ======================================================== Main SKAS Pod configuration
    skas:
      clusterIssuer: cluster-issuer1
    
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
          - name: ldap1
            groupPattern: "dep1_%s"
          - name: ldap2
            groupPattern: "dep2_%s"
    
        providerInfo:
          crd:
            url: http://localhost:7012
          ldap1:
            url: http://localhost:7013
          ldap2:
            url: https://skas-skas2-ldap.skas-system.svc
            rootCaPath: /tmp/cert/ldap2/ca.crt
            insecureSkipVerify: false
            clientAuth:
              id: skMerge
              secret: ${LDAP2_CLIENT_SECRET}
    
        extraEnv:
          - name: LDAP2_CLIENT_SECRET
            valueFrom:
              secretKeyRef:
                name: ldap2-client-secret
                key: clientSecret
    
        extraSecrets:
          - secret: skas-skas2-ldap-cert
            volume: ldap2-cert
            mountPath: /tmp/cert/ldap2
    
      skLdap:
        enabled: true
        # --------------------------------- LDAP configuration
        ldap:
          host: ldap1.mydomain.internal
          insecureNoSSL: false
          rootCaData: "LS0tLS1CRUdJTiBDRVJUSUZ................................lRJRklDQVRFLS0tLS0K"
          bindDN: cn=Manager,dc=mydomain1,dc=internal
          bindPW: admin123
          groupSearch:
            baseDN: ou=Groups,dc=mydomain1,dc=internal
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
    
    # ======================================================== SKAS2 Pod configuration
    skas2:
      skAuth:
        enabled: false
    
      skMerge:
        enabled: false
    
      skCrd:
        enabled: false
    
      clusterIssuer: cluster-issuer1
    
      skLdap:
        enabled: true
        # --------------------------------- LDAP configuration
        ldap:
          host: ldap2.mydomain.internal
          insecureNoSSL: false
          rootCaData: "LS0tLS1CRUdJTiBDRVJUSUZ................................lRJRklDQVRFLS0tLS0K"
          bindDN: cn=Manager,dc=mydomain2,dc=internal
          bindPW: admin123
          groupSearch:
            baseDN: ou=Groups,dc=mydomain2,dc=internal
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
            port: 7113
            ssl: true
            services:
              identity:
                disabled: false
                clients:
                  - id: skMerge
                    secret: ${LDAP2_CLIENT_SECRET}
                protected: true
    
        extraEnv:
          - name: LDAP2_CLIENT_SECRET
            valueFrom:
              secretKeyRef:
                name: ldap2-client-secret
                key: clientSecret
    ```

These two block hold the same definition than the ones defined in the [Advanced configuration chapter](/advancedconfiguration). With two differences:

- `skas.skMerge.providerInfo.ldap2.url: https://skas-skas2-ldap.skas-system.svc`
- `skas.skMerge.extraSecrets[0].secret: skas-skas2-ldap-cert` 

To accommodate service and secret name change, due to aliasing of the second dependency. (Values was originally `https://skas2-ldap.skas-system.svc` and `https://skas2-ldap.skas-system.svc` )

Then, to launch the deployment, in the same folder as `Chart.yaml`, execute:

```shell
$ helm dependency build && helm -n skas-system upgrade -i skas .
```

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
$ kubectl sk init https://skas.ingress.ksprayX
Setup new context 'skas@ksprayX' in kubeconfig file '/tmp/kconfig'
```



