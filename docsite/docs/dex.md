
# DEX integration

[DEX](https://dexidp.io/) is an OpenID Connect provider. As such, will serve OpenID Connect clients to provide Single Sign On services.

It does not host any user identity information by itself, but rely on other Identity Provider for this, through configurable `connectors`.

A connector has been developed for SKAS. As DEX does not provide some extension mechanism, adding a connector requires 
to patch the code. So, a specific DEX image with a SKAS connector has been build.

Deploying DEX in standalone mode require to operation:

- Reconfigure SKAS to open a service for the usage of DEX (`login` service)
- Deploying DEX itself, with the proper connector configuration.

In the following, three variants of this configuration will be described. One with the connection in clear text, and two secured, with network encryption and inter-pod authentication.

> _Even if your target is a fully secured configuration, we suggest you first implement the unsecured, simplest variant, and then modify it incrementally, as described._

## Clear text connection

### SKAS reconfiguration

The login service is provided by the `skAuth` SKAS module. It is disabled by default and must be enabled to be used. 
Refer to [Architecture/Modules and interface](architecture.md#modules-and-interfaces) for more information

Here is a values file to enable this service: 

???+ abstract "values.skas.login.yaml"

    ``` { .yaml .copy } 
    skAuth:
      exposure:
        external:
          services:
            login:
              disabled: false
    ```

Note than, by default, the `skAuth` module provide only SSL encrypted service. This will be the case of our `login` service.

To deploy this configuration:

```shell
$ helm -n skas-system upgrade -i skas skas/skas --values ../../values.init.yaml \
--values ./values.skas.login.yaml
```

> _Don't forget to add the `values.init.yaml`, or to merge it in the `values.skas.yaml` file. Also, if you have others values file, they must be added on each upgrade_

> _And don't forget to restart the pod(s). See [Configuration: Pod restart](configuration.md#pod-restart)_

### DEX deployment

For this sample, we will use the official [DEX Helm chart](https://github.com/dexidp/helm-charts/tree/master/charts/dex). 
This will require some configuration, by providing a specific value file.

Here is such file (Some value will need to be adjusted to your context)

??? abstract "values.dex.yaml"

    ``` { .yaml .copy } 
    image:
      repository: ghcr.io/skasproject/dex
      tag: v2.37.0-skas-0.2.1
    
    config:
      issuer: http://dex.ingress.mycluster.internal
      storage:
        type: kubernetes
        config:
          inCluster: true
      web:
        http: 0.0.0.0:5556
      logger:
        level: info
        format: text
      oauth2:
        skipApprovalScreen: true
      connectors:
        - type: skas
          id: skas
          name: SKAS
          config:
            loginPrompt: "User"
            loginProvider:
              url: https://skas-auth.skas-system.svc
              insecureSkipVerify: true
    
      staticClients:
        - id: example-app
          redirectURIs:
            - 'http://127.0.0.1:5555/callback'
          name: 'Example App'
          secret: ZXhhbXBsZS1hcHAtc2VjcmV0
    
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop:
          - ALL
      readOnlyRootFilesystem: false
      runAsNonRoot: true
      runAsUser: 1000
      seccompProfile:
        type: RuntimeDefault
    
    ingress:
      enabled: true
      className: nginx
      hosts:
        - host: dex.ingress.mycluster.internal
          paths:
            - path: /
              pathType: ImplementationSpecific
              backend:
                service:
                  name: dex
                  port:
                    number: 5556
    ```

Here are some comment about this values file:

- The `image` section target the SKAS patched image of DEX.
- The `config` section is the DEX configuration file. Refer to the 
[sample DEX config file](https://github.com/dexidp/dex/blob/master/examples/config-dev.yaml) for more explanation.
- The `config.issuer` value must be adjusted to your local DNS name. Note it is an unsecured URL.
- The `config.connnectors[0]` is the SKAS specific section. 
- The `config.connnectors[0].config.loginProvider.url` value targets the `skAuth` service. A noted above, 
this is an SSL encrypted service.
- The `config.connnectors[0].config.loginProvider.insecureSkipVerify` is set to `true`. As the targeted service is 
using HTTPS, we skip the certificate authority validation for this first sample.
- the `staticClients` section define a first OIDC client with parameter compatible with the `example-app` described below. 
- The `securityContext` section explicit some security constraints. This may be useful if your cluster implements 
some security restriction on running PODs. 
- The `ingress` section should be adjusted, at least for the `host:` url and maybe more if you use another ingress controller than nginx.  

As we will use the public DEX helm chart, its repo must first be defined:

```shell
$ helm repo add dex https://charts.dexidp.io
```

Then we can proceed to the DEX deployment:

```shell
$ helm -n skas-system upgrade -i dex dex/dex --values ./values.dex.yaml
```

If everything is OK, you should have two PODs running:

```shell
$ kubectl -n skas-system get pods
NAME                    READY   STATUS    RESTARTS   AGE
dex-54b4698bcd-9wbz6    1/1     Running   0          5h5m
skas-5cc75b8ff9-pw7nd   3/3     Running   0          6h23m
```

In case of problems, you may be want to check the resulting configuration. Unfortunately, this Helm chart store it in 
a secret. This means the configuration values are encoded en base64.

To display it, you can type:

```shell
$ kubectl get secret -n skas-system dex -o jsonpath="{ $.data.config\.yaml }" | base64 -d
```

If you modify some value in the `values.skas.login.yaml` file, execute again the Helm deployment command and restart the DEX Pod:

```shell
$ kubectl -n skas-system rollout restart deployment dex
```

### Testing

By convention all OIDC provider must provide a `well-known` endpoint which describe its other endpoints and other configuration values.

You can test this endpoint with the following command: 

```shell
$ curl http://dex.ingress.mycluster.internal/.well-known/openid-configuration
{
  "issuer": "http://dex.ingress.mycluster.internal",
  "authorization_endpoint": "http://dex.ingress.mycluster.internal/auth",
  "token_endpoint": "http://dex.ingress.mycluster.internal/token",
  "jwks_uri": "http://dex.ingress.mycluster.internal/keys",
  ....
```

This will ensure at least DEX is started and you ingress is functional.

> _Of course, you must adjust the URL to your context._ 

To go further, DEX provide a raw [example-app](https://github.com/dexidp/dex/tree/master/examples/example-app).
The main purpose of this application is to provide a starting point for developer to integrate an OIDC client in their 
code. But it also provide an interactive tool to test an OIDC service.

For your convenience, we have setup a [repository to host binaries of this application, for several OS and processor](https://github.com/skasproject/dex-example-app/releases/tag/2.37.0).

For example, to download/install this binary for a Mac Intel:

```shell
$ cd /tmp
$ curl -L https://github.com/skasproject/dex-example-app/releases/download/2.37.0/example-app_2.37.0_darwin_amd64 -o ./example-app
$ sudo chmod 755 example-app
$ sudo mv example-app /usr/local/bin
```

```shell
$ example-app --issuer  http://dex.ingress.mycluster.internal
2023/08/28 18:50:07 listening on http://127.0.0.1:5555
```

Now, launch your browser to this provided link (`http://127.0.0.1:5555`). You should land on a page like this:

![](images/example-app1.png)

Click on the `Login` button. You should then land on a login page:

![](images/example-app2.png)

Enter a valid SKAS user account ('admin' for example) and you should land on a page like this:

![](images/example-app3.png)

This is not really 'user friendly', but it is a test application.

You can have a look on the log of SKAS and DEX logs. Also, you can test an invalid login.

> Of course, for this to work, you must fully preserve the configuration of `staticClients` in the DEX config file

DEX github repo also provide this `example-app` as a container. You can launch it as:

```shell
$ docker run -p 5555:5555 ghcr.io/dexidp/example-app:latest  example-app --issuer  http://dex.ingress.kspray6 --listen http://0.0.0.0:5555
2023/08/28 17:29:04 listening on http://0.0.0.0:5555
```

## Securing connection

The previous configuration has a major security issue: The login and password information are entered through a clear text connection. 

This following configuration will fix this point and also add authentication between DEX and SKAS and validate the SKAS certificate.

### SKAS reconfiguration

Here is the modified values file for SKAS reconfiguration.

???+ abstract "values.skas.login.yaml"

    ``` { .yaml .copy } 
    skAuth:
      exposure:
        external:
          services:
            login:
              disabled: false
              clients:
                - id: dex
                  secret: "aSharedSecret"
    ```

The service authentication has been activated.

To deploy this configuration, use the same command as previously:

```shell
$ helm -n skas-system upgrade -i skas skas/skas --values ../../values.init.yaml \
--values ./values.skas.login.yaml
```

> _And restart the POD_

### DEX deployment

And here is the modified values file for DEX deployment

??? abstract "values.dex.yaml"

    ``` { .yaml .copy }     
    image:
      repository: ghcr.io/skasproject/dex
      tag: v2.37.0-skas-0.2.1
    
    config:
      issuer: https://dex.ingress.mycluster.internal
      storage:
        type: kubernetes
        config:
          inCluster: true
      web:
        http: 0.0.0.0:5556
      logger:
        level: info
        format: text
      oauth2:
        skipApprovalScreen: true
      connectors:
        - type: skas
          id: skas
          name: SKAS
          config:
            loginPrompt: "User"
            loginProvider:
              url: https://skas-auth.skas-system.svc
              rootCaPath: ""
              rootCaData: "LS0tLS1CRUdJTiBDRVJU.......................ENFUlRJRklDQVRFLS0tLS0K"
              insecureSkipVerify: false
              clientAuth:
                id: "dex"
                secret: "aSharedSecret"
    
      staticClients:
        - id: example-app
          redirectURIs:
            - 'http://127.0.0.1:5555/callback'
          name: 'Example App'
          secret: ZXhhbXBsZS1hcHAtc2VjcmV0
    
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop:
          - ALL
      readOnlyRootFilesystem: false
      runAsNonRoot: true
      runAsUser: 1000
      seccompProfile:
        type: RuntimeDefault
    
    ingress:
      enabled: true
      className: nginx
      annotations:
        cert-manager.io/cluster-issuer: your-cluster-issuer
        nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
      hosts:
        - host: dex.ingress.mycluster.internal
          paths:
            - path: /
              pathType: ImplementationSpecific
      tls:
        - secretName: dex-server-tls
          hosts:
            - dex.ingress.kspray6
    ```

The modification are the following:

- The `config.issuer` endpoint is now using HTTPS.
- The `config.connectors[0].config.loginProvider.rootCaData` is populated with the Certificate Authority of the 
  `skAuth` service. To find its value, we can dig inside its certificate, which include its authority:
  ```shell
  $ kubectl -n skas-system get secret skas-auth-cert -o=jsonpath='{.data.ca\.crt}'
  ```
- The `config.connectors[0].config.loginProvider.clientAuth` is set to authenticate with the id/secret defined above for `skAuth` login service.
- The `ingress` is now configured to handle SSL connection (And force SSL for non-SSL connection)

To apply this new configuration, use the same command as previously:

```shell
$ helm -n skas-system upgrade -i dex dex/dex --values ./values.dex.yaml
```

And restart the POD:

```shell
$ kubectl -n skas-system rollout restart deployment dex
```

> Another security improvement would be to configure the ingress in SSL passthroughs and to configure the DEX Pod to
handle SSL itself. This would ensure real end to end encryption. Unfortunately, this is not possible with the current
Helm Chart and refactoring it is out of the scope of this documentation.

### Testing

You can test again this endpoint with the following command. Note the https:// now on URLs

```shell
$ curl https://dex.ingress.mycluster.internal/.well-known/openid-configuration
{
  "issuer": "https://dex.ingress.mycluster.internal",
  "authorization_endpoint": "https://dex.ingress.mycluster.internal/auth",
  "token_endpoint": "https://dex.ingress.mycluster.internal/token",
  "jwks_uri": "https://dex.ingress.mycluster.internal/keys",
  "userinfo_endpoint": "https://dex.ingress.mycluster.internal/userinfo", 
  ....
```

And use again the `example-app` application. Note the 'https://' on the issuer.

```shell
$ example-app --issuer  https://dex.ingress.mycluster.internal
2023/08/28 18:50:07 listening on http://127.0.0.1:5555
```

#### Got a certificate issue ?

You may get the following on the curl request:

```shell
$ curl https://dex.ingress.mycluster.internal/.well-known/openid-configuration
curl: (60) SSL certificate problem: unable to get local issuer certificate
More details here: https://curl.haxx.se/docs/sslcerts.html
....
```

This is the case if the DEX issuers certificate has been signed by an authority which is not recognized by your workstation.

Solution is to retrieve this certificate:

```shell
$ kubectl -n skas-system get secret dex-server-tls -o=jsonpath='{.data.ca\.crt}' | base64 -d >./CA.crt
```

And to provide it to the Curl command:

```shell
$ curl https://dex.ingress.mycluster.internal/.well-known/openid-configuration \
--cacert ./CA.crt
{
  "issuer": "https://dex.ingress.mycluster.internal",
  "authorization_endpoint": "https://dex.ingress.mycluster.internal/auth",
  "token_endpoint": "https://dex.ingress.mycluster.internal/token",
  "jwks_uri": "https://dex.ingress.mycluster.internal/keys",
  "userinfo_endpoint": "https://dex.ingress.mycluster.internal/userinfo", 
  ....
```

You will encounter the same issue with the `example-app` test application. Here also, provide the certificate:

```shell
$ example-app --issuer https://dex.ingress.kspray6 --issuer-root-ca ./CA.crt
2023/08/29 09:31:53 listening on http://127.0.0.1:5555
```

## Using a Kubernetes secret

There is still a security issue, as two shared secreta (aSharedSecret and the staticClients secret) are in clear text 
in both values file. As such, they may ends up in some version control system.

So, let's store these values in Kubernetes secrets and access them using environment variables.

Here is a secret aimed to be shared between DEX and SKAS. Its value can be randomly generated, as it is accessed by both party. 

???+ abstract "dex-client-secret.yaml"

    ``` { .yaml  .copy}
    apiVersion: v1
    kind: Secret
    type: Opaque
    metadata:
      name: dex-client-secret
      namespace: skas-system
    data:
      DEX_CLIENT_SECRET: cGZRM3lXSTBBN2M3aGJE
    ```

> There is several solutions to generate such secret value. One can use Helm with some random generator function. Or use a [Secret generator](toolsandtricks.md#secret-generator)

And here is the secret shared between DEX (In `config.staticClients[0]` and the `example-app` application binary)

???+ abstract "example-app-secret.yaml"

    ``` { .yaml  .copy}
    apiVersion: v1
    kind: Secret
    type: Opaque
    metadata:
      name: example-app-secret
      namespace: skas-system
    data:
      EXAMPLE_APP_SECRET: WlhoaGJYQnNaUzFoY0hBdGMyVmpjbVYw    # Result of printf "ZXhhbXBsZS1hcHAtc2VjcmV0" | base64
    ```

> _Its value is hard-coded in `example-app`, so must not be changed (Or you must pass the new value as `--client-secret` parameter on launch)._

> _Note than both secret are formatted in a way compatible with `spec.containers[X].envFrom`. This is required by the DEX Helm chart._

### SKAS reconfiguration

Here is the modified values file for SKAS reconfiguration.

??? abstract "values.skas.login.yaml"

    ``` { .yaml  .copy}
    skAuth:
      exposure:
        external:
          services:
            login:
              disabled: false
              clients:
                - id: dex
                  secret: ${DEX_CLIENT_SECRET}
    
      extraEnv:
        - name: DEX_CLIENT_SECRET
          valueFrom:
            secretKeyRef:
              name: dex-client-secret
              key: DEX_CLIENT_SECRET
    ```

The modifications are the following:

- The `skAuth.extraEnv` subsection inject the secret value as an environment variable in the container.
- the `skAuth.exposure.external.services.identity.clients[0].secret` fetch its value through this environment variable.

> Most of the values provided by the helm chart ends up inside a configMap, which is then loaded by the SKAS executable.
The environment variable interpolation occurs during this load.

### DEX deployment

And here is the modified values file for DEX deployment

??? abstract "values.dex.yaml"

    ``` { .yaml  .copy}
    image:
      repository: ghcr.io/skasproject/dex
      tag: v2.37.0-skas-0.2.1
    
    config:
      issuer: https://dex.ingress.mycluster.internal
      storage:
        type: kubernetes
        config:
          inCluster: true
      web:
        http: 0.0.0.0:5556
      logger:
        level: info
        format: text
      oauth2:
        skipApprovalScreen: true
      connectors:
        - type: skas
          id: skas
          name: SKAS
          config:
            loginPrompt: "User"
            loginProvider:
              url: https://skas-auth.skas-system.svc
              rootCaPath: ""
              rootCaData: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUdTekNDQkRPZ0F3SUJBZ0lKQU4zclBySE5JRmZBTUEwR0NTcUdTSWIzRFFFQkN3VUFNSFV4Q3pBSkJnTlYKQkFZVEFrWlNNUTR3REFZRFZRUUlEQVZRWVhKcGN6RU9NQXdHQTFVRUJ3d0ZVR0Z5YVhNeEdUQVhCZ05WQkFvTQpFRTl3Wlc1RVlYUmhVR3hoZEdadmNtMHhGakFVQmdOVkJBc01EVWxVSUVSbGNHRnlkRzFsYm5ReEV6QVJCZ05WCkJBTU1DbU5oTG05a2NDNWpiMjB3SGhjTk1qRXdPREU0TURreU16QTFXaGNOTXpFd09ERTJNRGt5TXpBMVdqQjEKTVFzd0NRWURWUVFHRXdKR1VqRU9NQXdHQTFVRUNBd0ZVR0Z5YVhNeERqQU1CZ05WQkFjTUJWQmhjbWx6TVJrdwpGd1lEVlFRS0RCQlBjR1Z1UkdGMFlWQnNZWFJtYjNKdE1SWXdGQVlEVlFRTERBMUpWQ0JFWlhCaGNuUnRaVzUwCk1STXdFUVlEVlFRRERBcGpZUzV2WkhBdVkyOXRNSUlDSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQWc4QU1JSUMKQ2dLQ0FnRUE2cFBqejVCb1FYczZ2NE02NHBwc1JvTEFBeHMxQmovRmJGZTBzVzIycG9XL1d0L1NHSWVVVTdCUgpmZUhJOUdNbE53WmlkSXV0ZEU5d0N1a2pIbVVLbVhmRUx4MXlhamdTSm5PSmR1cWdCZHBHNTVwLzhtQ2lubVk2Cm1Pdis2V0hGMHFIYjVEZjZTTHhpSFNkTEZQVWtwV3IrbmI2T0JxaERsZ2JNUjA5WVZwV1ZHTlFqalQvSWdhbmIKRG12S1Z2S0VXTEk2cGZGVDhxWW5rTnNxamQ3T1NiZGdaVlRPTGh4YVJIZ2xVUlc3dGNvaXIyYW8rWFRNSkJaVAo4elRmS1BOVmcwK3c5ODBmVVY0dCtSZElFdXREbTdFa1JCcXBXNkZtZktFOFlhb2thWHJxMjgrVUFWSFpMd1BxCk1jVnFTeVNzaVR0bTBSYXhjcG9aYUQ5SjdWWlB1UGxlOUluUE1sNTJYaE1pZHBlNG9SYW1JRjJlUnNOZExkVTYKQklPcTBtNHRaTkk1QnRwbXBTZVlMOXBBMmtGL3UwT2Z1VWNUbVZTSlBGMkMybVJETVpmMVMxVFVGYnVIK1N2ZgoraTQ4bFVoSjIvajlURVkxRk1DM0oxMkVBUXk0YXpBa0FXWkdKUDBBdzBpdFBjUkJVMkJ0ZXV1VWhhQlNWTU9JCkxxSGFhTXRhZzJCUXcwblBhTDhabFNRcVJyakF0NnRaUDhqTnNpRFBxSE9SOVFDb29EbGZoWUJ5T3l2Ry9FWHEKWXpVUUV3NXF2NkdiSzJLYWs5U0s0ckhqRGF6V1l3a0Mza1grbkxiREFmcUNMNkhpWUMyL0ZiQzVwVmlyM0o5RwppM0JIVFBTRk9rQ2t3QkJMNGE4ZWxDRWRmajEvTlRxNDYzNzRiQU1jSHcvV2dqdzhCT0VDQXdFQUFhT0IzVENCCjJqQVBCZ05WSFJNQkFmOEVCVEFEQVFIL01CMEdBMVVkRGdRV0JCUk90dmNMS0E4UFhOUGs1bmRFL0Y5SldKb3gKbWpDQnB3WURWUjBqQklHZk1JR2NnQlJPdHZjTEtBOFBYTlBrNW5kRS9GOUpXSm94bXFGNXBIY3dkVEVMTUFrRwpBMVVFQmhNQ1JsSXhEakFNQmdOVkJBZ01CVkJoY21sek1RNHdEQVlEVlFRSERBVlFZWEpwY3pFWk1CY0dBMVVFCkNnd1FUM0JsYmtSaGRHRlFiR0YwWm05eWJURVdNQlFHQTFVRUN3d05TVlFnUkdWd1lYSjBiV1Z1ZERFVE1CRUcKQTFVRUF3d0tZMkV1YjJSd0xtTnZiWUlKQU4zclBySE5JRmZBTUEwR0NTcUdTSWIzRFFFQkN3VUFBNElDQVFESgpxZGI3Myt4cWFqclNuaHoxOTlWZGR3RUVvWGVSTi9jbkY0ZUdQODk0dURBSCtvcWYvVDNhUExZaWxHdnVoZElwCmUrUFk4Z2dsdUJRa3hzd1pDQjFzSFNGUFVHOFZPWmNQVU1SZGV1TVVqTUczcEhRT3J4N2VMV1hYRXNnblJ3MTcKcjIvei92L3VVVmovaW15Z0cwQWRkV0t2Y2ZEZ3AwcHNlUFRMY0xaRkdURU1nN3Y2RWswWFRLMXlEdlhzWmliUQpWcTdVMEE1SE5nNm40SzByNFBycTNQTTdCZWVpQnpkY21yaDR4MkIzcXkvWDY3SXF5K2pTMFZZS2NmVFkwQ25FClR1KzE5cjJlSGY0ZGM4VXIzbzJSamptQUJ5cHBHYVQ0RDdkZ0g0a0hkeDJoN2NmMWhTeVozU3lMang4VkZFdzIKbmhFVjcrUVBXM2g1RExDaXMwelcyY2pyRFhKblIxT3dpR3NqWmgyWUZlMytiUUNiQU5sV0F0dVNaZTg0ZkVlLwpJeHRqLzhBMlAvd0thaitWeGIrWFZKd3l0YzJJbUhYUTdMcjZ6MlMzcFJUTmcyREQ3V2d2WkZvVk5WWkllR0xaCmJDYkVjdmpPQkdDSU1DK0tyV0dMYlQzaTFlMUxpY2k5MWFxWGNIcDlyRVpTbE8va1BHZjVnWDZGSmNqNmpWbzcKUDZLQ2xCbUloVllITXVlb3JIN09VRmw4bWRzVmF5eE1COGR6bHI0OXlRUXpocWlmM3l3TEpRRXBDbENzYnEvZApKMkQ5M0JUQTh6NWN0bzRJNW9DdGZRMkdqbGtmRUpHODYzZ2NJVC8zaWV1M0FJLytMQVRGTzcrVFlWcVlZOFNJCndEUVZ4czF3T3BIWk9FZWtmTzRmS1cxMkJRK2YrSzltK2owSVNGelVDQT09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
              insecureSkipVerify: false
              clientAuth:
                id: "dex"
                secret: ${DEX_CLIENT_SECRET}
    
      staticClients:
        - id: example-app
          redirectURIs:
            - 'http://127.0.0.1:5555/callback'
          name: 'Example App'
          secretEnv: EXAMPLE_APP_SECRET
    
    envFrom:
      - secretRef:
          name: dex-client-secret
      - secretRef:
          name: example-app-secret
    
    
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop:
          - ALL
      readOnlyRootFilesystem: false
      runAsNonRoot: true
      runAsUser: 1000
      seccompProfile:
        type: RuntimeDefault
    
    ingress:
      enabled: true
      className: nginx
      annotations:
        cert-manager.io/cluster-issuer: your-cluster-issuer
        nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
      hosts:
        - host: dex.ingress.kspray6
          paths:
            - path: /
              pathType: ImplementationSpecific
      tls:
        - secretName: dex-server-tls
          hosts:
            - dex.ingress.mycluster.internal
    ```

DEX handle environment variable extension in two different ways, depending of the subsection:

- Environment variable are expanded the usual way in the `connectors` definition.
- This is not the case for other parts of the configuration, such as `staticClients`. So two new attributes have been
added in a staticClient definition: `idEnv` and `secretEnv`. Only this last is used in our case.

Both secret values are injected in the Pod by the `envFrom:` subsection