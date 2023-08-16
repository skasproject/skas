
# Configuration

## Principle

As installation was performed using an helm chart, configuration will be performed by providing a 'values' file overriding 
the default [`values.yaml`](https://github.com/skasproject/skas/blob/main/helm/skas/values.yaml) of the helm chart

This is what was did in the initial configuration, which such a file:

```shell
$ cat >./values.init.yaml <<EOF
clusterIssuer: your-cluster-issuer

skAuth:
  exposure:
    external:
      ingress:
        host: skas.ingress.mycluster.internal
  kubeconfig:
    context:
      name: skas@mycluster.internal
    cluster:
      apiServerUrl: https://kubernetes.ingress.mycluster.internal
EOF
```

To apply a modified file, the `helm upgrade` command should be used.

```shell
$ helm -n skas-system upgrade skas skas/skas --values ./values.init.yaml
```

### Pod restart

For the new configuration to be taken in account, the `skas` pod(s) must be restarted. The best and simple way is to perform a 'rollout' on the skas deployment:

```shell
$ kubectl -n skas-system rollout restart deployment skas
deployment.apps/skas restarted
```

> _There is some solution to perform an automatic restart. See [reloader](toolsandtricks.md/#reloader)_

SKAS is a very flexible product and, as such, there is a lot of variables in the default `values.yaml` of the helm chart. 
Fortunately, default values are appropriate in most case. 

We will not describe in this chapter all the variables (You can refer to comments in the file) but will explicit some typical configuration variation. 


## Skas behavior

Here is a values file which redefine the most common variable related to SKAS behavior:

```
$ cat >./values.behavior.yaml <<"EOF"
# Default value. May be overridden by component
log: 
  mode: json # 'json' or 'dev'
  level: info

skAuth:
  # Define password requirement
  passwordStrength:
    forbidCommon: true    # Test against lists of common password
    minimumScore: 3       # From 0 (Accept anything) to 4

  tokenConfig:
    # After this period without token validation, the session expire
    inactivityTimeout: "30m"
    # After this period, the session expire, in all case.
    sessionMaxTTL: "12h"
    # This is intended for the client CLI, for token caching
    clientTokenTTL: "30s"


skCrd:
  initialUser:
    login: admin
    # passwordHash: $2a$10$ijE4zPB2nf49KhVzVJRJE.GPYBiSgnsAHM04YkBluNaB3Vy8Cwv.G  # admin
    commonNames: ["SKAS administrator"]
    groups:
      - skas-admin
EOF
```

- There is a `log` section, to adjust the level and to set the mode. By default, `log.mode` is set to `json`, 
aimed to be injected to a log management external system. To have a more 'human' form, `log.mode` can be set to `dev`.
- `skAuth.passwordStrength` will allow to modify the criteria of a valid password. 
- `skAuth.token.config` section will configure the token lifecycle.
- `skCrd.initialUser` will define the default admin user. Note the `passwordHash` has been commented out, 
otherwise password would be reset on each apply of these values.

> _The meaning of `skAuth` and `skCrd` subsection is described in the [Architecture](architecture.md) chapter._

Then, to apply a modified configuration:

```shell
$ helm -n skas-system upgrade skas skas/skas --values ./values.init.yaml \
--values ./values.behavior.yaml
```

We still need to add `values.init.yaml`, otherwise, corresponding default/empty values will be reset.

> _Don't forget to restart the pod(s). See [above](#pod-restart)_

## Kubernetes integration

Here is a values file which redefine the most common variable related to SKAS integration with Kubernetes:

```
$ cat >./values.k8s.yaml <<EOF

replicaCount: 1

# -- Annotations to be added to the pod
podAnnotations: {}

# -- Annotations to be added to all other resources
commonAnnotations: {}

image:
  pullSecrets: []
  repository: ghcr.io/skasproject/skas
  # -- Overrides the image tag whose default is the chart appVersion.
  tag:
  pullPolicy: IfNotPresent

# Node placement of SKAS pod(s) 
nodeSelector: {}
tolerations: []
affinity: {}

EOF
```

- `replicaCount` allow to define the number of pod replica for SKAS deployment. Note we are in an active-active configuration, with no need for a leader election mechanism.
- `podAnnotations` and `commonAnnotations` will allow to annotate pods and others SKAS resources, if required.
- `image` subsection will allow to define an alternate image version or location. Useful in an air-gap deployment, where SKAS image is stored in a private repository. 
- `nodeSelector`, `toleration` and `affinity` are usual Kubernetes properties related to the node placement of SKAS pod(s)

To apply a modified configuration:

```shell
$ helm -n skas-system upgrade skas skas/skas \
--values ./values.init.yaml --values ./values.behavior.yaml --values ./values.k8s.yaml
```

> _Don't forget to restart the pod(s). See [above](#pod-restart)_
