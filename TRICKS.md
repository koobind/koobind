
# Debugging build from SCRATCH container

On the node where the pod is running (here: loki-grafana) :

```
docker ps | grep loki-grafana-bd864f786-nt9rt
f91c662f37dd        grafana/grafana                                     "/run.sh"                4 hours ago         Up 4 hours                              k8s_grafana_loki-grafana-bd864f786-nt9rt_loki_1547d7f9-0ae2-45e4-8aba-d00ab523e98d_0
3308f14400a5        gcr.io/google_containers/pause-amd64:3.1            "/pause"                 4 hours ago         Up 4 hours                              k8s_POD_loki-grafana-bd864f786-nt9rt_loki_1547d7f9-0ae2-45e4-8aba-d00ab523e98d_0

docker run -it --network container:f91c662f37dd --pid container:f91c662f37dd --privileged centos bash

ps aux

ls /proc/1/root
```      
 
# git discard all changes

```
git checkout .
git clean -fd
```  
May be also:
```
git reset
git restore ...  
```

# Get a certificate

```
kubectl -n koo-system get secrets webhook-server-cert -o=jsonpath='{.data.ca\.crt}' | base64 -d
```

