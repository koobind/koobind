#!/bin/bash

MYDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cat <<EOF >$MYDIR/generated/webhooks_patch.yaml
---
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  name: mutating-webhook-configuration
webhooks:
  - name: muser.kb.io
    clientConfig:
      caBundle: $(base64 -i $MYDIR/cert/tls.crt)
      url: https://koomgrdev:9443/mutate-directory-koobind-io-v1alpha1-user
  - name: mgroup.kb.io
    clientConfig:
      caBundle: $(base64 -i $MYDIR/cert/tls.crt)
      url: https://koomgrdev:9443/mutate-directory-koobind-io-v1alpha1-group
  - name: mbinding.kb.io
    clientConfig:
      caBundle: $(base64 -i $MYDIR/cert/tls.crt)
      url: https://koomgrdev:9443/mutate-directory-koobind-io-v1alpha1-binding
---
apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  name: validating-webhook-configuration
webhooks:
  - name: vuser.kb.io
    clientConfig:
      caBundle: $(base64 -i $MYDIR/cert/tls.crt)
      url: https://koomgrdev:9443/validate-directory-koobind-io-v1alpha1-user
  - name: vgroup.kb.io
    clientConfig:
      caBundle: $(base64 -i $MYDIR/cert/tls.crt)
      url: https://koomgrdev:9443/validate-directory-koobind-io-v1alpha1-group
  - name: vbinding.kb.io
    clientConfig:
      caBundle: $(base64 -i $MYDIR/cert/tls.crt)
      url: https://koomgrdev:9443/validate-directory-koobind-io-v1alpha1-binding
EOF

