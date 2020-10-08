# kafka-demo
Trying out Kafka using Minkube, golang and more


# Install Kafka on Minikube
> [install minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)
```
minikube config set cpu 4
minikube config set memory 4096
minikube start
minikube enable addons helm-tiller
```

```
helm repo add incubator http://storage.googleapis.com/kubernetes-charts-incubator
helm install --name kafka incubator/kafka

==> v1/ConfigMap
NAME             DATA  AGE
kafka-zookeeper  3     1s

==> v1/Pod(related)
NAME               READY  STATUS             RESTARTS  AGE
kafka-0            0/1    Pending            0         0s
kafka-zookeeper-0  0/1    ContainerCreating  0         0s

==> v1/Service
NAME                      TYPE       CLUSTER-IP      EXTERNAL-IP  PORT(S)                     AGE
kafka                     ClusterIP  10.109.101.159  <none>       9092/TCP                    1s
kafka-headless            ClusterIP  None            <none>       9092/TCP                    0s
kafka-zookeeper           ClusterIP  10.107.106.226  <none>       2181/TCP                    1s
kafka-zookeeper-headless  ClusterIP  None            <none>       2181/TCP,3888/TCP,2888/TCP  1s

==> v1/StatefulSet
NAME             READY  AGE
kafka            0/3    0s
kafka-zookeeper  0/3    0s

==> v1beta1/PodDisruptionBudget
NAME             MIN AVAILABLE  MAX UNAVAILABLE  ALLOWED DISRUPTIONS  AGE
kafka-zookeeper  N/A            1                0                    1s
```


## Test Client
```
kubectl apply -f test-client

- create a new topic:
    kubectl exec -ti testclient -- ./bin/kafka-topics.sh --zookeeper kafka-zookeeper:2181 --topic test-sgune --group group-sgune --create --partitions 1 --replication-factor 1

    Created topic "test-sgune".

- list all kafkatopics with:
    kubectl exec -ti testclient -- ./bin/kafka-topics.sh --zookeeper kafka-zookeeper:2181 --list

- add messages to the topic sgune-test
    kubectl exec -ti testclient -- ./bin/kafka-console-producer.sh --broker-list kafka:9092  --topic test-sgune

- listen for messages on a topic:
    kubectl exec -ti testclient -- ./bin/kafka-console-consumer.sh --bootstrap-server kafka:9092 --topic test-sgune --from-beginning

- list consumer groups
    kubectl exec -ti testclient -- ./bin/kafka-consumer-groups.sh  --list --bootstrap-server kafka:9092

- describe a group
    kubectl exec -ti testclient -- ./bin/kafka-consumer-groups.sh  --bootstrap-server kafka:9092 --describe --group group-sgune

- delete a topic
    kubectl exec -ti testclient -- ./bin/kafka-topics.sh --zookeeper kafka-zookeeper:2181 --delete --topic test-sgune
  ```

## Message Producer
  ```
  kubectl exec -ti testclient -- ./bin/kafka-console-producer.sh --broker-list kafka:9092  --topic test-sgune
    >message 1
    >message 2
    >message 3
```

## Message Consumer
```
kubectl exec -ti testclient -- ./bin/kafka-console-consumer.sh --bootstrap-server kafka:9092 --topic test-sgune --from-beginning
    message 1
    message 2
    message 3
```


# Local Testing
## Zookeeper & Kafka Containers
```
docker run --name zookeeper  -p 2181:2181 -d zookeeper

docker logs -f kafka

export ZOOKEEPER_ADDR="$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' zookeeper)"

docker run -p 9092:9092 --name kafka  -e KAFKA_ZOOKEEPER_CONNECT="$ZOOKEEPER_ADDR:2181" -e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://127.0.0.1:9092 -d confluentinc/cp-kafka

docker logs -f kafka

```
> `export KAFKA_ADDR=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' kafka)`

> `export KAFKA_TOPIC=test-sgune`

Then you can run `./release/producer` and `./release/consumer` in different terminals

or

## Docker

Images : 
- producer - https://hub.docker.com/r/shreyasgune/kafka-producer 
- consumer - https://hub.docker.com/r/shreyasgune/kafka-consumer 


### Producer
```
docker build --rm -t shreyasgune/kafka-producer:latest -f producer/Dockerfile .

docker run -p 8080:8080 -e KAFKA_TOPIC=$KAFKA_TOPIC -e KAFKA_ADDR=$KAFKA_ADDR shreyasgune/kafka-producer:latest
```

### Consumer
```
docker build --rm -t shreyasgune/kafka-consumer:latest -f consumer/Dockerfile .

docker run -p 8081:8081 -e KAFKA_TOPIC=$KAFKA_TOPIC -e KAFKA_ADDR=$KAFKA_ADDR shreyasgune/kafka-consumer:latest
```

## Minikube


> Remember to find the IP address for the Kafka-svc deployed and use it for `KAFKA_ADDR` env var in `deployment.yaml`

```
kubectl apply -f consumer/k8s
kubectl apply -f producer/k8s
```

### Verify
- Producer
```
kubectl port-forward pod/kafka-producer-* 8080

curl GET 'localhost:8080/status'

curl --location --request POST 'localhost:8080/data?key=gojira&value=another%20world' \
--header 'Content-Type: application/x-www-form-urlencoded'

200 OK
{
    "gojira": "another world",
    "status": "Request has been pushed to the broker"
}

```

```
kubectl logs kafka-producer-* -f

[GIN-debug] POST   /data
[GIN-debug] GET    /status 
[GIN-debug] Listening and serving HTTP on :8080
gojira:another world pushed to broker[GIN] 2020/10/08 - 05:40:05 | 200 |  1.772857896s |       127.0.0.1 | POST     "/data?key=gojira&value=another%20world"

- Consumer
```
kubectl logs kafka-consumer-* -f

message at offset 0:  = message 1
message at offset 1:  = message 2
message at offset 2:  = message 3
message at offset 3:  = 
message at offset 4:  = 
message at offset 5: gojira = another world

```

```
kubectl port-forward pod/kafka-consumer-* 8081
curl GET 'localhost:8081/status'
```

> I have also kept `kafka-js/connect.js` to test local connection to kafka-broker, if you need it.