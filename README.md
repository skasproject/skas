# SKAS: Simple Kubernetes Authentication System

SKAS is a powerful Kubernetes extension designed to streamline user authentication and authorization processes.

Whether you're managing a single cluster or a complex multi-cluster environment,
SKAS offers a solution to handle user authentication, integrating seamlessly with Kubernetes CLI and
supporting a range of identity providers.

## Main Features

- **Kubernetes Authentication Webhook and kubectl Extension:** SKAS provides a Kubernetes authentication webhook and an
  extension for kubectl, ensuring a smooth Kubernetes CLI integration without the need for browser interactions.

- **Users and Groups as Custom Resources:** Define users and groups as Kubernetes Custom Resources, giving you fine-grained control
  over access and permissions.

- **DEX support:** SKAS support DEX, making it compatible with all OIDC (OpenID Connect) aware
  applications such as Argocd and Argo Workflows.

- **LDAP Integration:** Support for one or several LDAP servers allows you to leverage existing identity
  infrastructure seamlessly.

- **LDAP facade:**  SKAS provides an 'LDAP server' interface, allowing LDAP clients from various applications to connect. (Experimental feature)

- **Unified User Information:** SKAS allows you to combine user information from multiple sources, including LDAP,
  local user databases, and more, providing a consolidated profile for users.

- **Centralized User Management:** In multi-cluster environments, SKAS simplifies user management,
  ensuring consistency and control across all clusters.

- **Delegated Management:** Delegate user and group management to specific individuals or teams, enhancing
  collaboration and reducing administrative burden.

- **Flexible Architecture:** SKAS is designed with flexibility in mind, accommodating sophisticated user management
  requirements in even the most complex environments.

- **ReadOnly LDAP/AD Access:** SKAS can operate with ReadOnly access to LDAP/AD servers, allowing user profiles to be
  enriched with local information, further enhancing the user experience.

Please, find its documentation [here](http://www.skas.skasproject.com/)

Version: 0.2.2

