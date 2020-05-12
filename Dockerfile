# kuttl builder
FROM golang:1.14 as builder

WORKDIR /go/src/kuttl
COPY . .

RUN go get -d -v ./...
RUN make cli

# release image with kubectl + kuttl
FROM golang:1.14

RUN apt-get update && apt-get install -y curl wget gnupg2 apt-transport-https vim
RUN curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -

RUN echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" | tee -a /etc/apt/sources.list.d/kubernetes.list
RUN apt-get update
RUN  apt-get install -y kubectl

COPY --from=builder /go/src/kuttl/bin/kubectl-kuttl /usr/bin/kubectl-kuttl

WORKDIR /opt/project

ENV KUBECONFIG=kubeconfig
