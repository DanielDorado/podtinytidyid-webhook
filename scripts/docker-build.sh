#!/usr/bin/env bash

# With CRC:
# docker build . -f docker/Dockerfile -t default-route-openshift-image-registry.apps-crc.testing/openshift/podtinytidyid:latest
# docker login default-route-openshift-image-registry.apps-crc.testing --username kubeadmin --password $(oc whoami -t)
# docker push default-route-openshift-image-registry.apps-crc.testing/openshift/podtinytidyid:latest

# With kind with local registry:
IMAGE=localhost:5001/podtinytidyid:0.0.1
docker build . -f docker/Dockerfile -t $IMAGE
docker push $IMAGE
