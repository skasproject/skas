# Architecture

## Overview

Here is the different modules involved for a SKAS authentication, right after installation: 

![Overview](./images/draw1.png){ align=left width=350}


SKAS is deployed as a Kubernetes Pod, this pod hosting three containers:

- `skAuth` which is in charge of delivering Kubernetes tokens and validate them.
- `skmerge`, which is in charge of building a consolidated identity from several identity providers. As in this configuration there is a single one, it appact as a simple passthrough
- `skCrd`, which is an identity provider storing user's information in the Kubernetes storage.

Arrow figure out the main communication flow between components. All of them are simple HTTP exchanges.

![](./images/empty.png)

For clarity, some connection has not been figured in this diagram:

- The `skCrd` module rely on the API server to store its user database, as Custom Resources.
- The `skAuth` module relay on the API server to store active tokens, as Custom Resources.

Here is a summary of exchange for an initial interaction

- The user issue a `kubectl` command (such as `kubectl get pods`). For this, a token is needed. It will be provided by the `kubectl-sk` client-go credential plugins.
- `kubectl-sk` prompt the user for login and password, then issue a `tokenCreate()` request to the `skAuth` module.
- The `skAuth` module issue a `getIdentity()` request with user credential. This request is forwarded to the `skCrd`module.
- The `skCrd` module retrieve user's information, check password validity and send information upward, to the `skMerge` module, which forward them to the `skAuth` module.
- The `skAuth` module generate a token and send it back to the `kubectl-sk` module. Which forward to `kubectl`.
- `kubectl` send the original request with the token to the Kubernetes API server
- The API Server send a `tokenReview()` request to the `skAuth` module, which reply with the user's informations (user id and groups).
- The API Server apply its RBAC rules on user's information to allow or deny requested operation. 

There is a more detailed description of this interaction as [sequence diagram](./architecture.md#sequence-diagrams).



## Sequence diagrams

Here is the sequence for a successful initial connexion.

``` mermaid
sequenceDiagram
  participant User
  participant kubectl
  participant kubectl-sk
  participant skAuth
  participant Api server
  autonumber
  User->>kubectl: User issue a<br>kubectl command
  kubectl->>kubectl-sk: kubectl launch<br>the kubectl-sk<br>credential plugin
  kubectl-sk->>kubectl-sk: kubeclt-sk lookup<br>for token in its <br>local cache.<br>NO in this case
  kubectl-sk->>User: kubectl-sk prompt<br>for login and password
  User-->>kubectl-sk: User provides its credential
  kubectl-sk->>skAuth: HTTP GET REQ:<br>getToken()
  skAuth->>skAuth: skAuth call<br>skMerge which<br>validate user<br>credentials and return<br>user's information<br>A token is generated.
  skAuth-->>kubectl-sk: Token in<br>HTTP response
  kubectl-sk-->>kubectl: kubectl-sk return<br>the token to kubectl<br>by printing it on<br>stdout and exit.
  kubectl->>Api server: kubectl issue the appropriate API call<br>with the provided bearer token
  Api server->>skAuth: The API Server<br>issue an<br>HTTP POST<br>with a<br>TokenReview<br>command
  skAuth-->>Api server: skAuth validate<br>the token and <br>provide user's<br>name and groups<br>in the response.
  Api server->>Api server: API Server validate<br>if user is allowed by<br>RBAC to perform<br>the requested action
  Api server-->>kubectl: API Server return the action result
  kubectl-->>User: kubectl display<br>the result and exits
```

Here is the sequence when a valid token is already present in client local cache:

``` mermaid
sequenceDiagram
  participant User
  participant kubectl
  participant kubectl-sk
  participant skAuth
  participant Api server
  autonumber
  User->>kubectl: User issue a<br>kubectl command
  kubectl->>kubectl-sk: kubectl launch<br>the kubectl-sk<br>credential plugin
  kubectl-sk->>kubectl-sk: kubeclt-sk lookup<br>for token in its <br>local cache.<br>YES in this case
  kubectl-sk->>kubectl-sk: Is the token still<br>valid against the<br>clientTokenTTL<br>YES in this case
  kubectl-sk-->>kubectl: kubectl-sk return<br>the token to kubectl<br>by printing it on<br>stdout and exit.
  kubectl->>Api server: kubectl issue the appropriate API call<br>with the provided bearer token
  Api server->>skAuth: The API Server<br>issue an<br>HTTP POST<br>with a<br>TokenReview<br>command
  skAuth-->>Api server: skAuth validate<br>the token and <br>provide user's<br>name and groups<br>in the response.
  Api server->>Api server: API Server validate<br>if user is allowed by<br>RBAC to perform<br>the requested action
  Api server-->>kubectl: API Server return the action result
  kubectl-->>User: kubectl display<br>the result and exits
```

Here is the sequence when a token is still valid, but the local cache (which is short lived) has expired:

``` mermaid
sequenceDiagram
  participant User
  participant kubectl
  participant kubectl-sk
  participant skAuth
  participant Api server
  autonumber
  User->>kubectl: User issue a<br>kubectl command
  kubectl->>kubectl-sk: kubectl launch<br>the kubectl-sk<br>credential plugin
  kubectl-sk->>kubectl-sk: kubeclt-sk lookup<br>for token in its <br>local cache.<br>YES in this case
  kubectl-sk->>kubectl-sk: Is the token still<br>valid against the<br>clientTokenTTL<br>NO in this case
  kubectl-sk->>skAuth: HTTP GET REQ:<br>validateToken()
  skAuth->>skAuth: skAuth check<br>if token is still valid.<br>Yes in this case
  skAuth-->>kubectl-sk: tokenValid response
  kubectl-sk-->>kubectl: kubectl-sk return<br>the token to kubectl<br>by printing it on<br>stdout and exit.
  kubectl->>Api server: kubectl issue the appropriate API call<br>with the provided bearer token
  Api server->>skAuth: The API Server<br>issue an<br>HTTP POST<br>with a<br>TokenReview<br>command
  skAuth-->>Api server: skAuth validate<br>the token and <br>provide user's<br>name and groups<br>in the response.
  Api server->>Api server: API Server validate<br>if user is allowed by<br>RBAC to perform<br>the requested action
  Api server-->>kubectl: API Server return the action result
  kubectl-->>User: kubectl display<br>the result and exits
```

And here is the sequence when a token has expired:

``` mermaid
sequenceDiagram
  participant User
  participant kubectl
  participant kubectl-sk
  participant skAuth
  participant Api server
  autonumber
  User->>kubectl: User issue a<br>kubectl command
  kubectl->>kubectl-sk: kubectl launch<br>the kubectl-sk<br>credential plugin
  kubectl-sk->>kubectl-sk: kubeclt-sk lookup<br>for token in its <br>local cache.<br>YES in this case
  kubectl-sk->>kubectl-sk: Is the token still<br>valid against the<br>clientTokenTTL<br>NO in this case
  kubectl-sk->>skAuth: HTTP GET REQ:<br>validateToken()
  skAuth->>skAuth: skAuth check<br>if token is still valid.<br>NO in this case
  skAuth-->>kubectl-sk: tokenInvalid<br>response
  kubectl-sk->>User: kubectl-sk prompt<br>for login and password
```



