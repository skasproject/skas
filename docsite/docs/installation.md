
# INSTALLATION

## Install SKAS helm chart

The simplest and recommended method to install the skas server is to use the provided helm chart.

The following is assumed

- Certificate manager is deployed in the target cluster and a `ClusterIssuer` is defined.
- There is an nginx ingress controller deployed in the target cluster.
- You have a local client kubernetes configuration with full admin rights on target cluster.
- Helm is installed locally.

First, add the SKAS helm repo:

```shell
$ helm repo add skas https://skasproject.github.io/skas-charts
```

Then, create a dedicated namespace:

```shell
$ kubectl create namespace skas-system
```

Then, you can deploy the helm chart:

```shell
$ helm -n skas-system install skas skas/skas \
    --set clusterIssuer=your-cluster-issuer \
    --set skAuth.exposure.external.ingress.host=skas.ingress.mycluster.internal \
    --set skAuth.kubeconfig.context.name=skas@mycluster.internal \
    --set skAuth.kubeconfig.cluster.apiServerUrl=https://kubernetes.ingress.mycluster.internal
```

With the following values, adjusted to your context:

- `clusterIssuer`: The Certificate Manager `ClusterIssuer` used to generate the certificate for all ingress access.
- `skAuth.exposure.external.ingress.host`: The ingress hostname used to access the SKAS service from outside of the cluster. 
  > _You will also have to define this hostname in your DNS._

The two following values will be used on generation of user's k8s config files:

- `skAuth.kubeconfig.context.name`: The context name which will be used to identify this cluster in the local config file. Can be any name
- `skAuth.kubeconfig.cluster.apiServerUrl`: The API server URL, from outside of the cluster. If you don't know it, you can find the information on an existing config file, with the yaml path `cluters[X].cluster.server`.

As an alternate approach, you can create a local values yaml file:

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

And issue the helm command as:

```shell
$ helm -n skas-system install skas skas/skas --values ./values.init.yaml
```

Then, if the installation is successful, you should be able to see the 'skas' server pod:

```shell
$ kubectl -n skas-system get pods
NAME                    READY   STATUS    RESTARTS   AGE
skas-746c54dc75-v8v2f   3/3     Running   0          25s
```

### Use another ingress controller instead of nginx

If using another ingress controller, launch the helm chart with `--set ingressClass=xxxx`. As the value will not be 'nginx', no ingress resource will be 
created by the helm chart. It is up to you to setup your own ingress. 
([Here](https://github.com/skasproject/skas/blob/main/helm/skas/templates/sk-auth/exposure/ingress.yaml) is the nginx definition, as a starting point.)

Please note that the ingress is configured with `ssl-passthroughs`. The underlying service will handle SSL.

### No Certificate Manager

If you do not use Certificate Manager, launch the helm chart without `ClusterIssuer` definition. 
Then, the secret hosting the certificate for the services will be missing and will need to be created it manually. (The `skas` pod will fail)

- Prepare PEM encoded self-signed certificate and key files.The certificate must be valid for the following hostnames:
    - `skas-auth`
    - `skas-auth.skas-system.svc`
    - `localhost`
    - `skas.ingress.mycluster.internal` (To be adjusted to your the effective host name provided above)
- Base64-encode the CA cert (in its PEM format) and its key.
- Create Secret in `skas-system` namespace as follows:

```shell
$ kubectl -n skas-system create secret tls skas-auth-cert --cert=<CERTIFICATE FILE> --key=<KEY FILE>
```

Then, the `skas` pod should start successfully.

## API Server configuration.

The Authentication Webhook of the API server should be configured to reach our authentication module.

### Manual configuration

Depending of your installation, the directory mentioned below may differs (For information, the clusters used for test and documentation are built with [kubespray](https://github.com/kubernetes-sigs/kubespray))

Also, this procedure assume the API Server is managed by the Kubelet, as a static Pod. If your API Server is managed by another system (i.e. systemd), you should adapt accordingly.

> **The following operations must be performed on all nodes hosting an instance of the Kubernetes API server**. Typically, all nodes of the control plane.

Also, these operations require `root` access on these nodes.

First, create a folder dedicated to `skas`:

```
[root]$ mkdir -p /etc/kubernetes/skas
```

Then, create the Authentication webhook configuration file in this folder (You can cut/paste the following):

```shell
[root]$ cat >/etc/kubernetes/skas/hookconfig.yaml <<EOF
apiVersion: v1
kind: Config
# clusters refers to the remote service.
clusters:
  - name: sk-auth
    cluster:
      certificate-authority: /etc/kubernetes/skas/skas_auth_ca.crt        # CA for verifying the remote service.
      server: https://sk-auth.skas-system.svc:7014/v1/tokenReview # URL of remote service to query. Must use 'https'.

# users refers to the API server's webhook configuration.
users:
  - name: skasapisrv

# kubeconfig files require a context. Provide one for the API server.
current-context: authwebhook
contexts:
- context:
    cluster: sk-auth
    user: skasapisrv
  name: authwebhook
EOF
```

As you can see in this file, there is a reference to the certificate authority of the authentication webhook service. So, you must fetch it and copy to this location:

> NB: It is assumed `kubectl` command was installed in this node, with an administrator configuration.

```shell
[root]$ kubectl -n skas-system get secret skas-auth-cert -o=jsonpath='{.data.ca\.crt}' | base64 -d >/etc/kubernetes/skas/skas_auth_ca.crt  
```

```shell
[root]$ ls -l /etc/kubernetes/skas
total 8
-rw-r--r--. 1 root root  620 May 11 12:36 hookconfig.yaml
-rw-r--r--. 1 root root 1220 May 11 12:58 skas_auth_ca.crt
```

Now, you must edit the API server manifest file (`/etc/kubernetes/manifests/kube-apiserver.yaml`) to load the `hookconfig.yaml` file:

```shell
[root]$ vi /etc/kubernetes/manifests/kube-apiserver.yaml
```

First step is to add two flags to the kube-apiserver command line:

- `--authentication-token-webhook-cache-ttl`: How long to cache authentication decisions.
- `--authentication-token-webhook-config-file`: The path to the configuration file we just setup

Here is what it should look like:

```yaml
...
spec:
  containers:
  - command:
    - kube-apiserver
    - --authentication-token-webhook-cache-ttl=30s
    - --authentication-token-webhook-config-file=/etc/kubernetes/skas/hookconfig.yaml
    - --advertise-address=192.168.33.16
    - --allow-privileged=true
    - --anonymous-auth=True
...
```

And the second step will consists to map the node folder `/etc/kubernetes/skas` inside the API server pod, under the same path.
This is required as these files are accessed in the API Server container context.

For this, a new `volumeMounts` entry should be added:

```yaml
    volumeMounts:
    - mountPath: /etc/kubernetes/skas
      name: skas-config
    ....
```

And a corresponding new `volumes`  entry:

```yaml
  volumes:
  - hostPath:
      path: /etc/kubernetes/skas
      type: ""
    name: skas-config
  ....
```

And another configuration parameter must be defined. The `dnsPolicy` must be set to `ClusterFirstWithHostNet`. Ensure such key does not already exists and add it:

```yaml
  hostNetwork: true
  dnsPolicy: ClusterFirstWithHostNet 
```

This complete the API Server configuration. Saving the edited file will trigger a restart of the API Server.

For more information, the kubernetes documentation on this topic is [here](https://kubernetes.io/docs/reference/access-authn-authz/webhook/)

**Remember: Perform this on all nodes hosting an instance of API Server.**

### Using an Ansible role

If ansible is one of your favorite tool, you may automate these tedious tasks by using an ansible role.

You will find such a role [here](https://github.com/skasproject/skas/releases/download/0.2.1/skas-apiserver-role-0.2.1.tgz)

As the manual installation, you may need to modify it accordingly to you local context.

To use it, we assume you have an ansible configuration with an inventory defining the target cluster. Then:

- Download, and untar the role archive provided above in a folder which is part of the rolepath. 
- create a playbook file, such as :

```shell
$ cat >./skas.yaml <<EOF
- hosts: kube_control_plane  # This group must target all the nodes hosting an instance of the kubernetes API server
  tags: [ "skas" ]
  vars:
    skas_state: present
  roles:
  - skas-apiserver
EOF
```

- Launch the playbook:

```shell
$ ansible-playbook ./skas.yaml
```

The playbook will perform all the steps described in the manual installation above. This will generate a restart of the API server.

### Troubleshooting

A small typo or incoherence in configuration may lead to API Server unable to restart.
If this is the case, you may have a look in the logs of the Kubelet (Remember, as a static pod, the API Server is managed by the Kubelet) in order to figure out what'is happen.

If you made a modification in this the `hookconfig.yaml` file, or if you update the CA file, will need to restart the API Server to reload the configuration.
But, the API Server is a 'static pod', a pod managed by the kubelet. As such, it can't be restarted as a standard pod.
The simplest way to trigger its effective reload is to modify the `/etc/kubernetes/manifests/kube-apiserver.yaml` file.
And you will need a real modification. Touch may not be enough. A common trick here is to modify slightly `the authentication-token-webhook-cache-ttl` flag value.

## SKAS CLI installation.

SKAS provide a CLI interface as an [extension of kubectl](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/)

Installation is straightforward:

- Download the executable for your OS/Architecture [at this location](https://github.com/skasproject/skas/releases/tag/0.2.1)
- Name it `kubectl-sk` (To comply to the naming convention of kubectl extension)
- Make it executable
- Move it to a folder accessed by your PATH.

For example, on a Mac Intel:

```shell
$ cd /tmp
$ curl -L https://github.com/skasproject/skas/releases/download/0.2.1/kubectl-sk_0.2.1_darwin_amd64 -o ./kubectl-sk
$ chmod 755 kubectl-sk
$ sudo mv kubectl-sk /usr/local/bin
```

You can now check the extension is effective

```shell
$ kubectl sk
A kubectl plugin for Kubernetes authentication

Usage:
kubectl-sk [command]

Available Commands:
completion  Generate the autocompletion script for the specified shell
hash        Provided password hash, for use in config file
help        Help about any command
init        Add a new context in Kubeconfig file for skas access
login       Logout and get a new token
logout      Clear local token
password    Change current password
user        Skas user management
version     display skas client version
whoami      Display current logged user, if any

Flags:
-h, --help                help for kubectl-sk
--kubeconfig string   kubeconfig file path. Override default configuration.
--logLevel string     Log level (default "INFO")
--logMode string      Log mode: 'dev' or 'json' (default "dev")

Use "kubectl-sk [command] --help" for more information about a command.
```

There is also a command to list all available plugins:

```shell
$ kubectl plugin list
....
/usr/local/bin/kubectl-sk
....
```

SKAS is now fully installed. You can now move on the [User guide](./userguide.md). 


## SKAS Removal

When performing SKAS removal, the first step is to reconfigure the API server.

If you configured it manually, then you must remove the two entries `--authentication-token-webhook-cache-ttl` and `--authentication-token-webhook-config-file` from the API server manifest file (/etc/kubernetes/manifests/kube-apiserver.yaml)

If you configured it using the ansible role, just modify the playbook by setting `skas_state: absent`: 

```shell
$ cat >./skas.yaml <<EOF
- hosts: kube_control_plane  # This group must target all the nodes hosting an instance of the kubernetes API server
  tags: [ "skas" ]
  vars:
    skas_state: absent
  roles:
  - skas-apiserver
EOF
```

and launch it:

```shell
$ ansible-playbook ./skas.yaml
```

Then, you can uninstall the helm chart

```shell
$ helm -n skas-system uninstall skas
```

And delete the namespace

```shell
$ kubectl delete namespace skas-system
```
