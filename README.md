This is a Go command that we include on our rabbitmq nodes that makes it easy and safe to perform common operations.

## Usage

### Show master

Shows the current master node.

```console
$ rabbitmq-clusterctl master
rabbit@master
```

### Join node

Joins the current node to the cluster.

```console
$ rabbitmq-clusterctl join
```

### Remove node

Removes the current node from the cluster.

```console
$ rabbitmqctl-clusterctl remove
```

### Promote node

Promotes this node to be the new master. You should ensure that all queues are synchronized on the node before promoting.

```console
$ rabbitmqctl-clusterctl promote
```
