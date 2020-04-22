#! /bin/sh
MYDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd $MYDIR
make manager
CMD="bin/manager  --certDir config/overlays/dev/cert --host localhost --namespace koo-system $@"
echo $CMD
$CMD
