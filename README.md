# SKAS: Simple Kubernetes Authentication System

SKAS is a powerful Kubernetes extension designed to streamline user authentication and authorization processes.
Whether you're managing a single cluster or a complex multi-cluster environment,
SKAS offers a seamless solution to handle user authentication, integrating seamlessly with Kubernetes CLI and
supporting a range of identity providers.

SKAS boasts an array of essential features to simplify and enhance your Kubernetes authentication and
authorization experience:

- **Kubernetes Authentication Webhook and kubectl Extension:** SKAS provides a Kubernetes authentication webhook and an
  extension for kubectl, ensuring a smooth Kubernetes CLI integration without the need for browser interactions.

- **Custom Users and Groups:** Define users and groups as Kubernetes Custom Resources, giving you fine-grained control
  over access and permissions.

- **DEX Connector:** SKAS includes a DEX connector, making it compatible with all OIDC (OpenID Connect) aware
  applications such as Argocd and Argo Workflows, enabling secure and straightforward integration.

- **LDAP Integration:** Support for one or several LDAP servers allows you to leverage existing identity
  infrastructure seamlessly.

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

With SKAS, you can ensure a secure, streamlined, and efficient authentication and authorization process,
enabling your Kubernetes clusters to operate at their full potential.

This introduction provides a more comprehensive overview of SKAS and its main features. Depending on your requirements,
you can further expand on each feature with dedicated sections in your documentation. If you need assistance with any
specific sections or have more details to add, please let us know.



Please, find its documentation [here](http://www.skas.skasproject.com/)