
## Features

- SkaGate
- A front for user management.
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
- Dex connector

## All

- Ensure copyright message in all relevant location
- ? Switch to logrus, to have more level (i.e WARNING) ?
- Display client.id in relevant message (Modify LoggingHandler)
- ? Rename sk-xxxx to skas-xxxxx ? or ska-xxxx  ?
- ? In config file, for http client url. Set full url, including the path OR define by scheme:, host: and port (Currently, it is ambiguous, as a partial) ?
- liveliness, readiness probes on all modules
- Think about concept of domain. May be corresponding to a list of providers. Check login@domain through DEX
- For certificates, provide a fallback when no cluster-issuer provided (cf topolvm)
- Provide schema for helm chart values.
- Add debug/trace on skhttp client.
- More info on version. cf dex
- systematize global rootCaPath, rootCaData
- Setup a system to force password change
- Think about a system who can safely delegate group binding inside a namespace
- Display client[].id on startup (Modify baseHandler by adding ClientManager ?)
- Add a protection against Brut force Attack 
- Check http handlers timeout (cf https://betterprogramming.pub/changes-in-go-1-20-b0a82d4b6c44, issue 6 )
- Tracing (open telemetry ?)

DONE:

- A sample helm chart to create a namespace, and an admin user and group.
- Rename userDescribe to userExplain
- Rename back userExplain to userDescribe
- Rename userStatus to userIdentity
- Rename tokenget to tokenCreate
- camelCase all url
- set uid as an int everywhere (In proto)
- helm: in rbac, add a rolebinding to an admin group (skas_admin by default)
- Sur un describe, distinguer la cause d'un password unchecked (Non fourni, ou non present dans le provider). Ceci pour pouvoir déterminer l'authoritée sans fournir le password.
- Generalize the concept of service in the config of all id provider.
- Service refactoring
  - Change the way we handle SSL: Always keep an non-ssl port on localhost, and add another port with SSL when required. When done, can remove localhost from certificate
  - In helm chart, use the fact default services config is open by default (Simplify some configmap) OR change the logic and make default to close everything.
    Default should be coherent: enabled and no check, or disabled and must set client * explicitly. (May be closed by default is better)
  - Two port should be managed. Each with its own set of services configuration. One intended to be bound on localhost and opened, for inside pod access.
    And one intended to be accessed externally, with default config to be closed.
- Refactor the provider configuration. To ease helm chart usage. (May be related to domain)
- Rename Identity.UserDetail.ProviderSpec to Identity.UserDetail.Provider


## sk-static

DONE

- Make user file dynamic (cf dexgate or certwatcher)

## sk-ldap

- Manage certificate with the same logic as sk-merge
- Hide ldap password in config

## sk-crd

DONE
 
- Add a service to change password

## sk-merge

- Manage rootCaPath in helm chart (Global and by provider)
- On startup, perform a scan to check underlying providers (A flag to disable)

DONE:

- There is still some UserStatus to change to UserIdentity
- Relay the changePassword service
- Display the list of provider as info on boot
- Add an optional providerList, to modify order of provider. Needed when appending a new provider to list as 'extraProvider' in helm chart

## sk-auth

- ? If memory or another non-k8s storage is used in production, one may modify the runnable package (Start() instead of Run(), a standard logger, ...) 
- ? Allow several kubeconfig definitions (Selected by ../v1/kubeconfig/<id>) ?
- ? Embed CLI binary with a download url ?
- Ability to add default namespace in kubeconfig init url
- kubeconfig: Always have a clientId (Missing at least for sk-client)

DONE:
- Relay userDescribe
- Use basic auth in complement to token auth on userDescribe
- Rename 'loginProvider' to 'downstreamProvider' (Or provider)
- Relay with authentication the changePassword service
- Relay login protocol to allow non-exposition of sk-merge for dex or alike
- In config, rename tokenConfig to token


## sk-client

- kubeconfig auth certitficate is not handled.
- Check from a system which has not the CA registered 
- Add "https://" on init if not present
- Reference and normalize os.exit() code
- Set a version number in saved config and tokenbag file name
- Optionally, request the client to provide id/secret on init (kubeconfig)

DONE
- kubeconfig configuration: Replace contextName and namespace by context.name and context.namespace
- Bug: If I got context with same name in two different config file, the local token bag is shared. Solution would be
  to add a same checksum of config file path in the token bag name (Or full path). Same for config.json

## packaging

- role: Test sk-auth is running before patching api-server
- Use go-releaser for the client

## sk-filter

- TODO ?
- Modify login protocol to support some error indication. (Invalid login/password vs unallowed vs ....)

## Doc:

Refer to topolvm docs for the following:
- Manually create a secret



