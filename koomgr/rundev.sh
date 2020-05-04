#! /bin/sh
MYDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd $MYDIR
set -e
make manager
export KUBECONFIG=${MYDIR}/rundev/kubeconfig
CMD="bin/manager  --config ${MYDIR}/rundev/config.yml --webhookCertDir ${MYDIR}/config/overlays/dev/cert --webhookHost localhost $@"
echo $CMD
$CMD
