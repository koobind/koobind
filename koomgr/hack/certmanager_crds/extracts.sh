
kubectl get customresourcedefinitions | grep cert-manager | cut -f 1 -d " " | while read a; do echo "---";  kubectl get customresourcedefinitions $a -o yaml ;done
