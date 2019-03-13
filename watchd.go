/*
  Ref:
    https://github.com/kubernetes/client-go/blob/release-6.0/examples/in-cluster-client-configuration/main.go
    https://github.com/kubernetes/client-go/blob/release-6.0/tools/record/event.go#L233
*/

package main

import (
	"fmt"
	"github.com/lizk1989/souvenir/utils"
	"go.etcd.io/etcd/clientv3"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func updateDeletedPodName(client *clientv3.Client, pod, ns, app string) {
	key := fmt.Sprintf("/deletedPods/%s/%s", ns, app)
	pods, err := utils.GetValue(client, key)
	ts := time.Now().Unix()
	if err != nil || pods == "" {
		utils.PutKeyValue(client, key, fmt.Sprintf("%d.%s", ts, pod))
	} else if !strings.Contains(pods, pod) {
		utils.PutKeyValue(client, key, fmt.Sprintf("%s,%d.%s", pods, ts, pod))
	}
}

func main() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// workaround for "curl -k"
	config.TLSClientConfig.CAFile = ""
	config.TLSClientConfig.Insecure = true
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	cli := utils.GetEtcdClient()
	defer cli.Close()

	watchIntf, err := clientset.CoreV1().Pods("").Watch((metav1.ListOptions{Watch: true}))
	if err != nil {
		panic(err.Error())
	}
	for event := range watchIntf.ResultChan() {
		if event.Type == watch.Deleted {
			meta := event.Object.(*apiv1.Pod).ObjectMeta
			name, ns, app := meta.Name, meta.Namespace, meta.Labels["app"]
			fmt.Printf("Found deleted pod %s under namespace %s, app %s\n", name, ns, app)
			updateDeletedPodName(cli, name, ns, app)
		}
	}
}
