
# Test docker image

docker run -v $(pwd)/sk-static:/config -p 7010:7010 ghcr.io/skasproject/skas:0.1.0 /sk-static --configFile config/config.yaml --usersFile config/users.yaml --bindAddr ":7010"

docker run -v $(pwd)/sk-ldap:/config -p 7011:7011 ghcr.io/skasproject/skas:0.1.0 /sk-ldap --configFile config/sampleconfigs/config-ldap-ops.yaml  --bindAddr ":7011"

docker run -v $(pwd)/sk-crd:/config -p 7012:7012 --env KUBECONFIG=config/tmp/kube.kspray1.scw01.yaml ghcr.io/skasproject/skas:0.1.0 /sk-crd --configFile config/config.yaml  --bindAddr ":7012"



# Kubernetes userdb initial object creation

> skcrd is a temporary project. Then some elements will be copied inside sk-crd. (sk-crd is not under the control of kubebuilder)

```
brew upgrade kubebuilder

cd ..../scratch
mkdir skcrd
cd skcrd/
go mod init skas/skcrd
kubebuilder init --domain skasproject.io
kubebuilder edit --multigroup=true
kubebuilder create api --group userdb --kind User --version v1alpha1
Create Resource [y/n]
y
Create Controller [y/n]
n
kubebuilder create api --group userdb --kind Group --version v1alpha1
Create Resource [y/n]
y
Create Controller [y/n]
n
kubebuilder create api --group userdb --kind GroupBinding --version v1alpha1
Create Resource [y/n]
y
Create Controller [y/n]
n


make manifests


git init
git add .
git status
git commit -m "initial commit"
git status

```

One can now create a goland project, (Create from existing sources). Then change gopath, and revert go.mod.

# Kubernetes session initial object creation

> sktoken is a temporary project. Then some elements will be copied inside sk-token. (sk-token will not be under the control of kubebuilder)

```
brew upgrade kubebuilder

cd ..../scratch
mkdir sktoken
cd sktoken/
go mod init skas/sktoken
kubebuilder init --domain skasproject.io
kubebuilder edit --multigroup=true
kubebuilder create api --group session --kind Token --version v1alpha1
Create Resource [y/n]
y
Create Controller [y/n]
n


make manifests


git init
git add .
git status
git commit -m "initial commit"
git status

```

One can now create a goland project, (Create from existing sources). Then change gopath, and revert go.mod.
