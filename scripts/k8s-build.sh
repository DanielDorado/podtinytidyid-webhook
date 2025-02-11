#!/usr/bin/env bash

mkdir build -p

if ! [[ -d build/cert ]]; then
    echo "Directory  build/cert/ not found."
    echo "Creating files in build/cert."
    mkdir -p build/cert
    ./scripts/webhook-create-signed-cert.sh build/cert
fi

# deployment
cp manifests/podtinytidyid.deployment.yml build/podtinytidyid.deployment.yml

# service
cp manifests/podtinytidyid.service.yml build/podtinytidyid.service.yml

# RBAC
cp manifests/podtinytidyid.rbac.yml build/podtinytidyid.rbac.yml

# configmap
cat manifests/podtinytidyid.configmap.yml \
    <(cat manifests/podtinytidyid-conf.yml | sed 's/^/    /') > \
    build/podtinytidyid.configmap.yml

# secret
sed "s/<CRT-PEM>/$(cat build/cert/webhook-server-tls.crt | base64 -w0)/" \
    manifests/podtinytidyid.secret.tpl.yml \
    | sed "s/<KEY-PEM>/$(cat build/cert/webhook-server-tls.key | base64 -w0)/" \
    > build/podtinytidyid.secret.tpl.yml

# webhook
sed "s/<CA-BUNDLE>/$(cat build/cert/ca.crt | base64 -w0)/" \
    manifests/podtinytidyid.webhook.tpl.yml \
    > build/podtinytidyid.webhook.yml
