The file [060-Deployment-strimzi-cluster-operator.yaml](strimzi-yaml-0.45.0/060-Deployment-strimzi-cluster-operator.yaml)
was altered to add the namespaces that we want strimzi kafka to watch
s/b value: "kafka,personae,strimzi"
(not valueFrom: fieldRef: fieldPath: metadata.namespace)

all myproject namespaces in yamls have to be sed replaced or find/replaced to strimzi

added the clusterrolebinding added-clusterrolebinding-operator-watched.yaml in config dir

github of strimzi files is:
https://github.com/strimzi/strimzi-kafka-operator