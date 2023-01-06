# SKAS

Simple Kubernetes Authentication System


Aimed to manage identity not only in a Web context but also in CLI context, without browser based interaction.

## Components

Ths SKAS system is made of the following components:

- Identity provider: Theses module manage a single source of identity, such as LDAP, static files, ....
- Identity aggregator: This module will connect to several identity provider to provide a single identity access point, 
  by combining information for a given user.
- Identity consumer: These modules will connect to the identity aggregator to control usage of some resources.

## Identity providers

sk-static: 

sk-crd

sk-ldap

## Identity aggregator

sk-join

## Identity consumers:

- sk-k8s: An authentication webhook to control Kubernetes access  
- dex: A dex modified with a skas connector
- sk-gate: A web proxy, controlling access to an unauthenticated web site. Replace Dex/Dexgate association. 
  May also handle session token for Kubernetes dashboard access.

## Interfaces

SKAS components interaction involve two communication interface:

- login interface: Provided by sk-join.

  Request is only user login/password and response is user's attribute (name, email, groups, ...) if authenticated
  
  This interface will be used by all identity consumers.

- Provider interface: provided by all identity providers. 
  
  This interface is designed for the sk-join module to be able to consolidate information about a given user.

- Admin interface. Provided by sk-join module, to allow some admin operation, such as user information lineage 

## Interface security.

SKAS will provide mechanisms to secure all communication between components by using TLS, 
Server certificate validation and also client certificate validation.

All theses mechanisms can be relaxed.

For example, the login interface, intended to be used by external consumer, may be configured to be used without client authentication.

Another use case would be to group run components in containers grouped in the same pod. 
In such case, communication can use local (127.0.0.1) network, without TLS layer.

The login interface will have some mechanisms to prevent BFA (Brute Force Attack)

- A increasing delay on wrong password on same login.
- Temporary Black list an account after several unsuccessful retry with different password.
  (Several retry with same password will ne be considered as an attack, but logged as a miss configured client)



