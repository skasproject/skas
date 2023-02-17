
## API Server configuration (Authentication webhok)

The Authentication Webhook of the API server should be configured to reach our authentication module.

Depending of your installation, the directory mentioned below may differs.
Also, this procedure assume the API Server is managed by the Kubelet, as a static Pod. If your API Server is managed by another system (i.e. systemd), you should adapt accordingly.

**The following operations must be performed on all nodes hosting an instance of the Kubernetes API server**. Typically, all nodes of the control plane.

Also, these operations require `root`access on these node.

First, create a folder dedicated to `skas`:

```
# mkdir -p /etc/kubernetes/skas
```

Then, create the Authentication webhook configuration file in this folder (You can cut/paste the following):

```
# cat >/etc/kubernetes/skas/hookconfig.yaml <<EOF
apiVersion: v1
kind: Config
# clusters refers to the remote service.
clusters:
  - name: sk-auth
    cluster:
      certificate-authority: /etc/kubernetes/skas/sk-auth-cert-ca.crt        # CA for verifying the remote service.
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

> NB: It is assumed `kubectl` command was installed in this node.

```
# kubectl -n skas-system get secret sk-auth-cert -o=jsonpath='{.data.ca\.crt}' | base64 -d >/etc/kubernetes/skas/sk-auth-cert-ca.crt  
```

```
# ls -l /etc/kubernetes/skas
total 8
-rw-r--r--. 1 root root  620 May 11 12:36 hookconfig.yaml
-rw-r--r--. 1 root root 1220 May 11 12:58 sk_auth_ca.crt
```

Now, you must edit the API server manifest file (`/etc/kubernetes/manifests/kube-apiserver.yaml`) to load the `hookconfig.yaml` file:

```
# vi /etc/kubernetes/manifests/kube-apiserver.yaml
```

First step is to add two flags to the kube-apiserver command line:

- `--authentication-token-webhook-cache-ttl`: How long to cache authentication decisions.
- `--authentication-token-webhook-config-file`: The path to the configuration file we just setup

Here is what it should look like:

```
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

```
    volumeMounts:
    - mountPath: /etc/kubernetes/skas
      name: skas-config
    ....
```

And a corresponding new `volumes`  entry:

```
  volumes:
  - hostPath:
      path: /etc/kubernetes/skas
      type: ""
    name: skas-config
  ....
```

And another configuration parameter must be defined. The `dnsPolicy` must be set to `ClusterFirstWithHostNet`. Ensure such key does not already exists and add it:

```
  hostNetwork: true
  dnsPolicy: ClusterFirstWithHostNet 
```

This complete the API Server configuration. Saving the edited file will trigger a restart of the API Server.

For more information, the kubernetes documentation on this topic is [here](https://kubernetes.io/docs/reference/access-authn-authz/webhook/)

**Remember: Perform this on all nodes hosting an instance of API Server.**

### Troubleshooting

A small typo or incoherence in configuration may lead to API Server unable to restart.
If this is the case, you may have a look in the logs of the Kubelet (Remember, as a static pod, the API Server is managed by the Kubelet) in order to figure out what'is happen.

If you made a modification in this the `hookconfig.yaml` file, or if you update the CA file, will need to restart the API Server to reload the configuration.
But, the API Server is a 'static pod', a pod managed by the kubelet. As such, it can't be restarted as a standard pod.
The simplest way to trigger its effective reload is to modify the `/etc/kubernetes/manifests/kube-apiserver.yaml` file.
And you will need a real modification. Touch may not be enough. A common trick here is to modify slightly `the authentication-token-webhook-cache-ttl` flag value.
