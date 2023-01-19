

# Kubernetes userdb object creation

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
