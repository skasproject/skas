



# API type



## userIdentity

- Given a login without password, retrieve user's attributes (groups, emails, uid, commonNames, is there a password)
- Given a login and password, add password status to information above (checked/failed)

## userDescribe

- Given a login, provide user information for each identity provider an a consolidated version, in the form of userIdentity.
- Given a login and password, add password status to information above (checked/failed)

This interface is intended to 'explain' to an administrator how the userIdentity is build

## userLogin

- Given a login and a password, retrieve user's attributes.
  The difference with `userIdentity` is it does not provide any information without valid password. 
  Thus it can be more safely exposed to outside world.

## changePassword

- Given a login, an old password and a new password, change the user password if the old one match. 
  The request convey also the 'authority', and identifier of the identity provider who validate the old password.

## tokenCreate

- Given a login and password, create a token and send if back (If password is correct)

## tokenRenew

- Given a token, return if it is still valid. Also, 'touch' the token.

The expiration is controlled by a server and occurs after some (configurable) time of inactivity.

## tokenReview

- Same as tokenRenew, but with a protocol understood be the Kubernetes API server authentication web hook

## kubeconfig

- This API provide enough information to build a client kubeconfig file.

## API exposition:

- Each API can be exposed at three levels:
  - On localhost, accessible only from other container in the same pod.
  - Inside Kubernetes, by adding a kubernetes Service. 
  - Outside Kubernetes, by adding a kubernetes ingress controller.

For each module, every exposed API can be accessed on two ports:
- One bound on localhost, for intra-pod communication
- One bound on pod interface, to be exposed as a service. Always using SSL encrypted communication

Depending of the configuration, only one or both port can be activated.


# Modules

Each module is implemented by a container.

## sk-crd, sk-ldap, sk-static

These modules support the `userIdentity`.

The sk-crd module also support the `changePassword` interface

By default, these APIs are exposed on a port bound on 'localhost' in clear text, without authentication. Optionally,
it can be exposed on another port, using SSL and protected by a client ID/Secret.

## sk-merge

This modules support `userDescribe` and `changePassword` APIs.

By default, these interfaces are exposed on a port bound on 'localhost' in clear text, without authentication. Optionally,
it can be exposed on another port, using SSL and protected by a client ID/Secret.

The `userDescribe` response is built by requesting the `userIdentity` to underlying provider and by aggregating the responses.

The `changePassword` is forwarded to the appropriate identity provider

## sk-auth

The services provided by this module are intended to be accessed from inside kubernetes (Using a k8s service) and externally (Using an ingress)

All API are encrypted using SSL.

By default, the localhost listener is turned off, as unused.

This module expose a set of API to kubernetes and external worlds, by adding a service and an ingress.

### tokenCreate, tokenRenew, changePassword

These APIs are intended to be accessed by the skas client (kubectl-sk)

They may be protected by a client ID/Secret, to be configured in the client.

### tokenReview

This interface is encrypted but not protected, as it should be accessible from the Kubernetes API server, using [specific protocol](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#webhook-token-authentication) 

### kubeconfig

This API is intended to be accessed by the skas client, eventually protected by a client Id/Secret

### userDescribe

This API is intended to be accessed by the skas client. As its usage should be reserved to system admins, it require the request to be authenticated with a skas token or an http Basic authentication.

### userLogin

This API is intended to be accessed by some others application to validate a login/password. This will allow support of OIDC authentication using a DEX module. It should be protected by a client ID/Secret




