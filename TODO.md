
## Features

- Automatic kubeconfig.context setup
- Deployment in one pod
- Dex connector
- SkaGate
- Documentation
  - Installation
  - Usage
  - Advanced config
    - Another ldap
    - Authentication in another cluster
    - skaGate / k8s dashboard
    - dex / argocd
  - Architecture


## All

- Ensure copyright message in all relevant location
- Reference and normalize os.exit() code
- Switch to logrus, to have more level (i.e WARNING)
- Display client.id in relevant message
- Rename sk-xxxx to skas-xxxxx ? or ska-xxxx
- Rename userDescribe to userExplain
- Rename userStatus to userIdentity
- In config file, for http client url. Set full url, including the path OR define by scheme:, host: and port (Currently, it is ambiguous, as a partial)
- liveliness, readiness probes on all modules
- Generalize the concept of service in the config of all id provider.
 
## sk-static

- Make user file dynamic (cf dexgate or certwatcher)

## sk-ldap

- Manage certificate with the same logic as sk-merge

## sk-crd

- Add a service to change password

## sk-merge

- Manage rootCaPath in helm chart (Global and by provider)
- On startup, perform a scan to check underlying providers (A flag to disable)
- rename to sk-bind ?
- Add an optional providerList, to modify order of provider. Needed when appending a new provider to list as 'extraProvider' in helm chart
- Relay the changePassword service
- Enable service by default (Replace Enabled by Disabled)

## sk-auth

- If memory or another non-k8s storage is used in production, one may modify the runnable package (Start() instead of Run(), a standard logger, ...)
- Relay with authentication the changePassword service
  vi /etc/kubernetes/manifests/kube-apiserver.yaml

## Doc:

Refer to topolvm docs for the following:
- Manually create a secret



