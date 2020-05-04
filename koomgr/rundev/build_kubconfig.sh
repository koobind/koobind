#! /bin/sh

SERVER="https://kspray1:6443"
NAMESPACE=koo-system

MYDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

SECRET_NAME=$(kubectl -n ${NAMESPACE} get serviceaccounts default -o jsonpath='{.secrets[0].name}')
CA=$(kubectl -n ${NAMESPACE} get secret/$SECRET_NAME -o jsonpath='{.data.ca\.crt}')
TOKEN=$(kubectl -n ${NAMESPACE} get secret/$SECRET_NAME -o jsonpath='{.data.token}' | base64 --decode)

cat >${MYDIR}/kubeconfig <<EOF
  apiVersion: v1
  kind: Config
  clusters:
  - name: default-cluster
    cluster:
      certificate-authority-data: ${CA}
      server: ${SERVER}
  contexts:
  - name: default-context
    context:
      cluster: default-cluster
      namespace: ${NAMESPACE}
      user: default-user
  current-context: default-context
  users:
  - name: default-user
    user:
      token: ${TOKEN}
EOF
