
# INSTALLATION

## Installing with SKAS Helm Chart

The most straightforward and recommended method for installing the SKAS server is by using the provided Helm chart. 

Before you begin, make sure you meet the following prerequisites:

- **Certificate Manager:** Ensure that the Certificate Manager is deployed in your target Kubernetes cluster, and a `ClusterIssuer` is defined for certificate management.

- **Ingress Controller:** An NGINX ingress controller should be deployed in your target Kubernetes cluster.

- **Kubectl Configuration:** You should have a local client Kubernetes configuration with full administrative 
rights on the target cluster.

- **Helm:** Helm must be installed locally on your system.

Follow these steps to install SKAS using Helm:

- Add the SKAS Helm repository by running the following command:

    ``` {.shell .copy}
    helm repo add skas https://skasproject.github.io/skas-charts
    ```

- Create a dedicated namespace for SKAS:

    ```{.shell .copy}
    kubectl create namespace skas-system
    ```

- Deploy the SKAS Helm chart using the following command:

    ```{.shell .copy}
    helm -n skas-system install skas skas/skas \
        --set clusterIssuer=your-cluster-issuer \
        --set skAuth.exposure.external.ingress.host=skas.ingress.mycluster.internal \
        --set skAuth.kubeconfig.context.name=skas@mycluster.internal \
        --set skAuth.kubeconfig.cluster.apiServerUrl=https://kubernetes.ingress.mycluster.internal
    ```

Replace the values with your specific configuration:

- `clusterIssuer`: The ClusterIssuer from your Certificate Manager for certificate management.
- `skAuth.exposure.external.ingress.host`: The hostname used for accessing the SKAS service from outside the cluster.<br> 
  > Make sure to define this hostname in your DNS.
- `skAuth.kubeconfig.context.name`: A unique context name for this cluster in your local configuration.
- `skAuth.kubeconfig.cluster.apiServerUrl`: The API server URL from outside the cluster. You can find this information 
in an existing Kubernetes config file under `clusters[X].cluster.server`.

Alternatively, you can create a local YAML values file as follows:

???+ abstract "values.init.yaml"

    ``` {.yaml .copy}
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
    ```

And then install SKAS using this values file:

```{.shell .copy}
helm -n skas-system install skas skas/skas --values ./values.init.yaml
```

After a successful installation, verify the SKAS server pod is running:

```{.shell}
$ kubectl -n skas-system get pods
> NAME                    READY   STATUS    RESTARTS   AGE
> skas-746c54dc75-v8v2f   3/3     Running   0          25s
```

### Use another ingress controller instead of nginx

If you are using an ingress controller other than NGINX, you can specify the ingress class by adding 
the --set ingressClass=xxxx flag when launching the Helm chart. In this case, the Helm chart won't create an ingress 
resource, and you will need to set up your own ingress. 
([Here](https://github.com/skasproject/skas/blob/main/helm/skas/templates/sk-auth/exposure/ingress.yaml) is the nginx definition, as a starting point.)

Please note that the ingress is configured with `ssl-passthroughs`. The underlying service will handle SSL.


### No Certificate Manager

If you are not using a Certificate Manager, you can still install SKAS. Follow these steps:

- Launch the helm chart without `ClusterIssuer` definition. Then, the secret hosting the certificate for the services 
will be missing, so the `skas` pod will fail
- Prepare a PEM-encoded self-signed certificate and key files. The certificate should be valid for the following hostnames:
    - `skas-auth`
    - `skas-auth.skas-system.svc`
    - `localhost`
    - `skas.ingress.mycluster.internal` (Adjust this to your actual hostname)
- Base64-encode the CA certificate (in PEM format) and its key.
- Create a secret in the skas-system namespace:
    ```{.shell .copy}
    $ kubectl -n skas-system create secret tls skas-auth-cert --cert=<CERTIFICATE FILE> --key=<KEY FILE>
    ```
- The skas pod should start successfully.

## API Server configuration.

The API server's Authentication Webhook must be configured to communicate with our authentication module.

### Manual configuration

Depending on your specific installation, the directory mentioned below may vary. For reference, the clusters used for testing and documentation purposes were built using [kubespray](https://github.com/kubernetes-sigs/kubespray).

Additionally, this procedure assumes that the API Server is managed by the Kubelet as a static Pod. If your API Server is managed by another system, such as systemd, you should make the necessary adjustments accordingly.

> _Please note that the following operations must be executed on all nodes hosting an instance of the Kubernetes API server, typically encompassing all nodes within the control plane._

These operations require 'root' access on these nodes._

To initiate the process, start by creating a dedicated folder for 'skas':"

```{.shell .copy}
mkdir -p /etc/kubernetes/skas
```

Next, create the Authentication Webhook configuration file within this directory. You can conveniently copy and paste the following configuration:

???+ abstract "/etc/kubernetes/skas/hookconfig.yaml"

    ```{.yaml .copy}
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
    ```

As indicated within this file, there is a reference to the certificate authority of the authentication webhook service. 
Consequently, you should retrieve it and place it in this location:

```{.shell .copy}
kubectl -n skas-system get secret skas-auth-cert \
-o=jsonpath='{.data.ca\.crt}' | base64 -d >/etc/kubernetes/skas/skas_auth_ca.crt  
```

> _Please ensure that the kubectl command is installed on this node with administrator configuration._

Inspect the folder's contents:
 
```{.shell}
$ ls -l /etc/kubernetes/skas
> total 8
> -rw-r--r--. 1 root root  620 May 11 12:36 hookconfig.yaml
> -rw-r--r--. 1 root root 1220 May 11 12:58 skas_auth_ca.crt
```

Now, you need to modify the API Server manifest file located at `/etc/kubernetes/manifests/kube-apiserver.yaml` to include the `hookconfig.yaml` file:"

```{.shell .copy}
vi /etc/kubernetes/manifests/kube-apiserver.yaml
```

The initial step involves adding two flags to the kube-apiserver command line:

- `--authentication-token-webhook-cache-ttl`: This determines the duration for caching authentication decisions.
- `--authentication-token-webhook-config-file`: This refers to the path of the configuration file we've just set up.

This is how it should appear:

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

The second step involves mapping the node folder `/etc/kubernetes/skas` inside the API server pod, using the same path. 
This mapping is necessary because these files are accessed within the context of the API Server container.

To achieve this, you should add a new volumeMounts entry as follows:

```yaml
    volumeMounts:
    - mountPath: /etc/kubernetes/skas
      name: skas-config
    ....
```

Additionally, you need to include a corresponding new volumes entry:

```yaml
  volumes:
  - hostPath:
      path: /etc/kubernetes/skas
      type: ""
    name: skas-config
  ....
```

Furthermore, you should define another configuration parameter. Specifically, you must set the `dnsPolicy` to 
`ClusterFirstWithHostNet`. Please verify that this key doesn't already exist and add or modify it accordingly:

```yaml
  hostNetwork: true
  dnsPolicy: ClusterFirstWithHostNet 
```

With these adjustments, you have completed the configuration for the API Server. Saving the edited file will trigger 
a restart of the API Server to take the changes into account.

For additional information, refer to the Kubernetes documentation on this topic, available [here](https://kubernetes.io/docs/reference/access-authn-authz/webhook/)

> **Please remember to carry out this procedure on all nodes that host an instance of the API Server.**

### Using an Ansible role

If Ansible is one of your preferred tools, you can automate these laborious tasks by utilizing an Ansible role.

You can obtain such a role [here](https://github.com/skasproject/skas/releases/download/0.2.2/skas-apiserver-role-0.2.2.tgz).

Similar to manual installation, you might need to customize it to suit your local context.

To utilize this role, we assume that you have an Ansible configuration in place, along with an inventory that defines the target cluster.

Additionally, this role utilizes the
[`kubernetes.core.k8s_info module`](https://docs.ansible.com/ansible/latest/collections/kubernetes/core/k8s_info_module.html).
Please review the requirements for this module

Then, follow these steps:

- Download and extract the role archive provided above into a folder that is part of the role path.
- Create a playbook file, for example:

???+ abstract "skas.yaml"

    ```{.yaml .copy }
    - hosts: kube_control_plane  # This group must target all the nodes hosting an instance of the kubernetes API server
      tags: [ "skas" ]
      vars:
        skas_state: present
      roles:
      - skas-apiserver
    ```

- Launch this playbook:

```{.shell .copy}
ansible-playbook ./skas.yaml
```

The playbook will execute all the steps outlined in the manual installation process detailed above. Consequently, 
this will trigger a restart of the API server.

### Troubleshooting

If there is a minor typo or a configuration inconsistency, it could potentially prevent the API Server from restarting. 
In such cases, it's advisable to examine the logs of the Kubelet. (Remember that, as a static pod, the API Server is 
managed by the Kubelet). These logs can provide insights into what might be causing the issue.

If you've made any modifications to the `hookconfig.yaml` file or updated the CA file, it's necessary to restart the 
API Server to apply the new configuration. However, since the API Server is a 'static pod' managed by the Kubelet, 
it can't be restarted like a standard pod.

The simplest method to effectively trigger a reload of the API Server is to make a modification to the 
`/etc/kubernetes/manifests/kube-apiserver.yaml` file. It's essential that this modification is a substantive change, 
as simply using the touch command may not suffice. A common approach is to make a slight modification to the 
`authentication-token-webhook-cache-ttl` flag value. This will prompt the API Server to reload its configuration 
and apply the changes.


## Installation of SKAS CLI

SKAS offers a command-line interface (CLI) as an [extension of kubectl](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/).

The installation process is straightforward:

- Download the executable that corresponds to your operating system and architecture [from this location](https://github.com/skasproject/skas/releases/tag/0.2.2).
- Rename the downloaded executable to `kubectl-sk` to adhere to the naming convention of kubectl extensions.
- Make the file executable.
- Move the `kubectl-sk` executable to a directory that is included in your system's PATH environment variable.

For instance, on a Mac with an Intel processor, you can use the following commands:

```{.shell .copy}
cd /tmp
curl -L https://github.com/skasproject/skas/releases/download/0.2.2/kubectl-sk_0.2.2_darwin_amd64 -o ./kubectl-sk
chmod 755 kubectl-sk
sudo mv kubectl-sk /usr/local/bin
```

Now, you can verify whether the extension is working as intended.

```{.shell .copy}
kubectl sk
```

It should display:

```
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

SKAS is now successfully installed. You can proceed with the [User guide](./userguide.md) for further instructions.

> _Depending on your cluster architecture, you may need to adjust your configuration for a safer and more resilient 
installation. Please refer to the [Configuration: Kubernetes Integration](configuration.md#kubernetes-integration) 
section for more information._

## SKAS Removal

When it comes to uninstalling SKAS, the initial step involves reconfiguring the API server. 
The approach depends on how you initially configured it:

If you configured it manually, remove the two entries, `--authentication-token-webhook-cache-ttl` and 
`--authentication-token-webhook-config-file`, from the API server manifest file located at 
`/etc/kubernetes/manifests/kube-apiserver.yaml`.

If you used the Ansible role for configuration, simply modify the playbook by setting `skas_state` to `absent`:

???+ abstract "skas.yaml"

    ```{.yaml .copy}
    - hosts: kube_control_plane  # This group must target all the nodes hosting an instance of the kubernetes API server
      tags: [ "skas" ]
      vars:
        skas_state: absent
      roles:
      - skas-apiserver
    ```

After making these changes, execute the playbook:

```{.shell .copy}
ansible-playbook ./skas.yaml
```

Once you have successfully reconfigured the Kubernetes API server, you can proceed to uninstall the Helm chart.

```{.shell .copy}
helm -n skas-system uninstall skas
```

And to delete the namespace

```{.shell .copy}
kubectl delete namespace skas-system
```
