
## Features

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

DONE:
- Automatic kubeconfig.context setup

## All

- Ensure copyright message in all relevant location
- Reference and normalize os.exit() code
- Switch to logrus, to have more level (i.e WARNING)
- Display client.id in relevant message (Modify LoggingHandler)
- Rename sk-xxxx to skas-xxxxx ? or ska-xxxx
- Rename userDescribe to userExplain
- Rename userStatus to userIdentity
- In config file, for http client url. Set full url, including the path OR define by scheme:, host: and port (Currently, it is ambiguous, as a partial)
- liveliness, readiness probes on all modules
- Generalize the concept of service in the config of all id provider.
- Think about concept of domain. May be corresponding to a list of providers. Check login@domain through DEX
 
## sk-static

- Make user file dynamic (cf dexgate or certwatcher)

## sk-ldap

- Manage certificate with the same logic as sk-merge
- Hide ldap password in config

## sk-crd

- Add a service to change password

## sk-merge

- Manage rootCaPath in helm chart (Global and by provider)
- On startup, perform a scan to check underlying providers (A flag to disable)
- rename to sk-bind ?
- Add an optional providerList, to modify order of provider. Needed when appending a new provider to list as 'extraProvider' in helm chart
- Relay the changePassword service

## sk-auth

- If memory or another non-k8s storage is used in production, one may modify the runnable package (Start() instead of Run(), a standard logger, ...)
- Relay with authentication the changePassword service
  vi /etc/kubernetes/manifests/kube-apiserver.yaml
- Allow several kubeconfig definitions (Selected by ../v1/kubeconfig/<id>)
- kubeconfig configuration: Replace contextName and namespace by context.name and context.namespace
- Embed CLI binary with a download url

## sk-filer

- TODO
- Modify login protocol to support some error indication. (Invalid login/password vs unallowed vs ....)

## Doc:

Refer to topolvm docs for the following:
- Manually create a secret



