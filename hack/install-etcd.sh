#!/usr/bin/env bash
echo Installing etcd

ETCD_VER=v3.3.11
DOWNLOAD_URL=https://github.com/etcd-io/etcd/releases/download

sudo mkdir -p /usr/local/kubebuilder/bin
sudo curl -L ${DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz -o /usr/local/kubebuilder/bin/etcd-${ETCD_VER}-linux-amd64.tar.gz
sudo tar xzvf /usr/local/kubebuilder/bin/etcd-${ETCD_VER}-linux-amd64.tar.gz -C /usr/local/kubebuilder/bin/ --strip-components=1
sudo chmod +x /usr/local/kubebuilder/bin/etcd

export PATH=$PATH:/usr/local/kubebuilder/bin
