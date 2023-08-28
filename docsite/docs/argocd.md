
# Argo CD integration with DEX

Argo CD embed an instance of DEX. It will take care its deployment and most if its configuration. 
In fact, only the `connectors:` part of the DEX config still remains to setup.

For this reason, the installation process described here will assume there is an Argo CD instance deployed in a 
'standard' way, with UI and ingress. (Such deployment is out of the scope of this documentation. Refer to the Argo CD 
documentation or/and use the [Argo CD helm chart](https://github.com/argoproj/argo-helm/tree/main/charts/argo-cd)) 

Then it should be proceeded to:

- The deployment of SKAS, with the configuration of a `login` service, for the use of the DEX instance embedded in Argo CD
- The configuration of this DEX instance, by patching the existing deployment.

## SKAS deployment

SKAS must be (re)configured to activate a `login` services. Here is the appropriate values file:

???+ abstract "values.skas2.yaml"

    ``` { .yaml .copy } 
    skAuth:
      exposure:
        external:
          services:
            login:
              disabled: false
              clients:
                - id: dex-argocd
                  secret: "aSharedSecret"
    ```

Note a client authentication has been setup, with a couple id/secret.

To apply this configuration:

```shell
$ helm -n skas-system upgrade -i skas skas/skas --values ./values.init.yaml --values ./values.skas.login.yaml
```

> _Don't forget to add the values.init.yaml, or to merge it in the values.skas.yaml file. Also, if you have others values file, they must be added on each upgrade_

> _And don't forget to restart the pod(s). See [Configuration: Pod restart](configuration.md#pod-restart)_

## Patching argo CD

Next step is the patch the Argo CD deployment. In fact, there is two patches:

- One to use the SKAS specific DEX image. (As DEX does not provide some extension mechanism to add connectors externally, a specific SKAS DEX image with a SKAS connector must be used).
- One to configure the `configMap` storing the DEX configuration.

Here is the patch file for the DEX image (It is a JSON RFC 6902 Patch): 

???+ abstract "dex-server-patch.json"

    ``` { .json .copy } 
    [
      { "op": "replace",
        "path": "/spec/template/spec/containers/0/image",
        "value": "ghcr.io/skasproject/dex:v2.37.0-skas-0.2.1"
      }
    ]
    ```

To be applied with the following command:

```shell
$ kubectl -n argocd patch deployment argocd-dex-server --type json --patch-file ./dex-server-patch.json
```

And here is the DEX configuration patch (It is a [Strategic Merge patch](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/#use-a-strategic-merge-patch-to-update-a-deployment)):

???+ abstract "argocd-cm-patch.yaml"

    ``` { .yaml .copy } 
    data:
      admin.enabled: "false"
      url: https://argocd.ingress.mycluster.internal
      dex.config: |
        connectors:
        - type: skas
          id: skas
          name: SKAS
          config:
            loginPrompt: "User"
            loginProvider:
              url: https://skas-auth.skas-system.svc
              rootCaData: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0......................09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
              insecureSkipVerify: false
              clientAuth:
                id: "dex-argocd"
                secret: "aSharedSecret"
    ```

The DEX `connectors:` config is similar to the one in DEX standalone configuration.

You must adjust:

- The `url: https://argocd.ingress.mycluster.internal` to point to your Argo CD server UI.
- The `admin.enabled` if you want to still be able to use the local Argo CD `admin` account.
- The `rootCaData:` is populated with the Certificate Authority of the
  `skAuth` service. To find its value, we can dig inside its certificate, which include its authority:
    ```shell
    $ kubectl -n skas-system get secret skas-auth-cert -o=jsonpath='{.data.ca\.crt}'
    ```

Then, this patch can be applied with the following command:

```shell
$ kubectl -n argocd patch configMap argocd-cm --type strategic --patch-file ./argocd-cm-patch.yaml
```

### Restart Pods

For the patch to be effective, some Argo CD pod should be restarted:

```shell
$ kubectl -n argocd rollout restart deployment argocd-dex-server && \
kubectl -n argocd rollout restart deployment argocd-server
```

## Test

Open your browser on the Argo CD UI. You should land with something like:

![](images/argocd1.png)

Click on the `LOG IN VIA SKAS` button. You should land on the DEX login page. Enter the a valid user login and password:

![](images/argocd2.png)

And you should land on the usual Argo CD Applications page. 

![](images/argocd3.png)

Click on the `User info` menu entry on the left to ensure we got the correct user information:

![](images/argocd4.png)

We can see the groups are correct. The Argo CD web UI choose to display the user's email in the Username field.
Our current user has no email defined, so the blank field.  

