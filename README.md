# souvenir
Watch deleted pod(name, namespace, labels.app) in k8s/openshift cluster.

### Target

To watch K8S/OpenShift cluster deleted pods, and record their names, namespaces, services.

For CMP, support APIs to query services, and deleted pod names for a specified namespace. After CMP found a list of deleted pod names, then it can go to ElasticSearch checkout logs via deleted pod names.

### API

1. GET /NAMESPACE
  查询指定NAMESPACE下，“历史上”有哪些services。
  成功返回: '{"services": "svc1,svc2,..."}'
  
2. GET /NAMESPACE/SERVICE
  查询指定NAMESPACE下的指定SERVICE下，“历史上”有哪些已经被删除的Pods。
  成功返回: '{"pods": "timestamp1.pod1, timestamp2.pod2, ..."}'
  
3. DELETE /NAMESPACE
  删除指定NAMESPACE的记录。
  
4. DELETE /NAMESPACE/SERVICE
  删除指定NAMESPACE下指定SERVICE的记录。

### Steps to deploy

0. Build and push image.

1. Create secrets with etcd ca, cert, key for souvenir pods to access etcd as storage:

        oc create secret generic certs --from-file=ca=/etc/origin/master/master.etcd-ca.crt --from-file=cert=/etc/origin/master/master.etcd-client.crt --from-file=key=/etc/origin/master/master.etcd-client.key

2. Add cluster admin role to default SA for souvenir pods to watch K8S API for pods in all namespaces.

        oc adm policy add-cluster-role-to-user admin -z default

3. User yaml/souvenir.dc.yaml as template, fill the following fields with corrent values:
    - spec.template.metadata.namespace
    - spec.template.spec.containers.env.ETCD_ENDPOINTS
    - spec.template.spec.containers.image
    - spec.template.spec.imagePullSecrets.name
    - spec.triggers.imageChangeParams.from.namespace

4. Create pods with dc yaml, and expose it as service and as route.
