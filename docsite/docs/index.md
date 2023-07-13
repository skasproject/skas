# SKAS: Simple Kubernetes Authentication System

SKAS is a Kubernetes extension aimed to handle users authentication and authorization.

Its main features are:

- Provide an Kubernetes authentication webhook and a kubectl extension, for seamless Kubernetes CLI integration (Without browser interaction)
- Allow definition of Users and Groups as Kubernetes custom resources.
- Provide a DEX connector to support all OIDC aware applications (Argocd, argo workflow, ...)
- Support one or several LDAP server(s).
- Provide ability to combine users informations from several sources (LDAP, Local users database, ....).
- Allow centralized user management in a multi-cluster environment.
- Allow delegation of the management for a subset of Users and/or Groups.
- Provide flexible architecture to handle sophisticated user management in a complex environment.
- Can use a ReadOnly access to the LDAP/AD server(s). User profile can then be enriched with local information.



