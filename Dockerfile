
#
# used to run kuttl tests from docker
# it is expected that a mount pt is provided where /opt/project is the root of the testsuite
# the root of the mount point needs a `kuttl-test.yaml` and a proper `kubeconfig` file which
# a process running in docker can reach.  If using kind, then `kind get kubeconfig --internal > kubeconfig`
# will provide the proper configuration for 0.7.0.  0.8.0 breaks this ability and is being worked on in the
# kind community.
# Assuming a test folders at mount root named "e2e" then:
# docker run -v $PWD:/opt/project kuttl e2e
# will run tests against the e2e folders.
#
# IF you want run tests from within a cluster, then specify `-e KUBECONFIG=`.  An empty KUBECONFIG will
# result in kuttl using the in-cluster-config.
# ex. docker run -e KUBECONFIG= -v $PWD:/opt/project kuttl e2e
# artifacts by default will land in the root of the mount point.

# kuttl builder
FROM golang:1.14 as builder

WORKDIR /go/src/kuttl
COPY . .

RUN go get -d -v ./...
RUN make cli

# release image with kubectl + kuttl
FROM debian:buster

RUN apt-get update && apt-get install -y curl wget gnupg2 apt-transport-https vim
RUN curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -

RUN echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" | tee -a /etc/apt/sources.list.d/kubernetes.list
RUN apt-get update
RUN  apt-get install -y kubectl

COPY --from=builder /go/src/kuttl/bin/kubectl-kuttl /usr/bin/kubectl-kuttl

WORKDIR /opt/project

# default configuration
ENV KUBECONFIG=kubeconfig

# standard kuttl test in entry point, flags and test folder can be provided as CMD
ENTRYPOINT ["kubectl", "kuttl", "test"]
