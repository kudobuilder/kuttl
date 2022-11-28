#!/usr/bin/env bash
echo Installing kube-apiserver

sudo mkdir -p /usr/local/kubebuilder/bin

sudo curl -L https://dl.k8s.io/v1.16.4/kubernetes-server-linux-amd64.tar.gz -o /usr/local/kubebuilder/bin/kubernetes-server-linux-amd64.tar.gz
sudo tar xzvf /usr/local/kubebuilder/bin/kubernetes-server-linux-amd64.tar.gz -C /usr/local/kubebuilder/bin/ --strip-components=1

sudo cp /usr/local/kubebuilder/bin/server/bin/kube-apiserver /usr/local/kubebuilder/bin

sudo chmod +x /usr/local/kubebuilder/bin/kube-apiserver

export PATH=$PATH:/usr/local/kubebuilder/bin
