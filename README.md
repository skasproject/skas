# SKAS: Simple Kubernetes Authentication System

SKAS is a Kubernetes extension aimed to handle users authentication and authorization.

Its main features are:

- Provide an Kubernetes authentication webhook and a kubectl extension, for seamless Kubernetes CLI integration (Without browser interaction)
- Provide a DEX connector to support all OIDC aware applications (Argocd, argo workflow, ...) 
- Allow definition of Users and Groups as Kubernetes custom resources.
- Support one or several LDAP server(s).
- Provide ability to combine users informations from several sources (LDAP, Local users database, ....). 
- Allow centralized user management in a multi-cluster environment.
- Allow delegation of the management for a subset of Users and/or Groups.
- Provide flexible architecture to handle sophisticated user management in a complex environment.

To go forward:

- [Installation](docs/installation.md)
- [Initial usage](docs/initial_usage.md)
- [Configure an LDAP server]()
- [Add a second LDAP Server]()
- [Delegate user database administration]()
- [Centralized user management in a multi-cluster context]() 
- [DEX integration. Argocd as an example]()





