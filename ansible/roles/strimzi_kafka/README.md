# strimzi_kafka

Ansible role for Block 2.3 that installs the Strimzi Cluster Operator and deploys a single-node KRaft Kafka cluster using `KafkaNodePool` and `Kafka` resources.

## Requirements

- Ansible
- `kubernetes.core` collection
- kubeconfig configured for the target cluster

Install the collection:

```bash
ansible-galaxy collection install kubernetes.core
```

## Role Variables

- `kafka_namespace`
- `strimzi_operator_install_url`
- `kafka_cluster_name`
- `kafka_version`
- `kafka_metadata_version`
- `kafka_replicas`
- `kafka_storage_size`
- `kafka_storage_class`

## Example

```bash
ansible-playbook ansible/playbooks/deploy-strimzi-kafka.yml
```
