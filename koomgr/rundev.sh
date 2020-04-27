#! /bin/sh
MYDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd $MYDIR
set -e
make manager
CMD="bin/manager  --webhookCertDir config/overlays/dev/cert --webhookHost localhost --namespace koo-system $@"
echo $CMD
$CMD
