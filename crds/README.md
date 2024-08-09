This directory contains CRD files for the kuttl configuration files.

## Use of CRD files for kuttl

While there is no currently available K8S controller to handle the kuttl configuration files,
the CRD definitions may be handy for kuttl users to leverage coding assistance for K8S CRs in 
their favorite IDE.

For intellij IDEA, refer to https://www.jetbrains.com/help/idea/kubernetes.html#crd for instructions
on how to load the CRD files either from:
- a local clone on your desktop 
- remote github raw url pointing to the kuttl repository
- from a K8S cluster where you'd register the CRDs (by running `kubectl apply -f <crd_file.yaml>` 

## Maintaining CRD files

In order to leverage IDE supports for Json schema, the json schema were extracted into distinct files and manually inlined into the CRD files.

Future possible automated include use of [carvel ytt](https://carvel.dev/ytt/), see sample usage at https://github.com/crossplane/crossplane/issues/3197#issuecomment-1191479570
