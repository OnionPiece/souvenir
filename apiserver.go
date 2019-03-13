package main

import (
	"encoding/json"
	"fmt"
	"github.com/lizk1989/souvenir/utils"
	"go.etcd.io/etcd/clientv3"
	"log"
	"net/http"
	"os"
	"strings"
)

type etcdHandler struct {
	client *clientv3.Client
}

func (h *etcdHandler) getNamespaceServices(w http.ResponseWriter, ns string) {
	key := fmt.Sprintf("/deletedPods/%s/", ns)
	svcs, err := utils.GetKeys(h.client, key)
	if err != nil {
		http.Error(w, "Failed to fetch data, datastore error", http.StatusInternalServerError)
		return
	} else if svcs == "" {
		http.Error(w, "No data found", http.StatusNotFound)
		return
	}
	svcs = strings.Replace(svcs, key, "", -1)
	ret := map[string]string{"services": svcs}
	jsonBytes, err1 := json.Marshal(ret)
	if err1 != nil {
		http.Error(w, "Failed to encode data to json", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(jsonBytes)
}

func (h *etcdHandler) doDelete(w http.ResponseWriter, ns, svc string) {
	key := ""
	withPrefix := false
	if svc != "" {
		key = fmt.Sprintf("/deletedPods/%s/%s", ns, svc)
	} else {
		key = fmt.Sprintf("/deletedPods/%s/", ns)
		withPrefix = true
	}
	err := utils.DeleteKey(h.client, key, withPrefix)
	if err != nil {
		http.Error(w, "Failed to delete, datastore error", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Deleted."))
}

func (h *etcdHandler) getServicePods(w http.ResponseWriter, ns, svc string) {
	key := fmt.Sprintf("/deletedPods/%s/%s", ns, svc)
	pods, err := utils.GetValue(h.client, key)
	if err != nil {
		http.Error(w, "Failed to fetch data, datastore error", http.StatusInternalServerError)
		return
	} else if len(pods) == 0 {
		http.Error(w, "No data found", http.StatusNotFound)
		return
	}
	ret := map[string]string{"pods": pods}
	jsonBytes, err1 := json.Marshal(ret)
	if err1 != nil {
		http.Error(w, "Failed to encode data to json", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(jsonBytes)
}

func (h *etcdHandler) handleFunc(w http.ResponseWriter, r *http.Request) {
	items := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	ns, svc := "", ""
	switch len(items) {
	case 2:
		if items[1] == "" {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}
		svc = items[1]
		fallthrough
	case 1:
		if items[0] == "" {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}
		ns = items[0]
	default:
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	if r.Method == http.MethodGet {
		if svc != "" {
			h.getServicePods(w, ns, svc)
		} else {
			h.getNamespaceServices(w, ns)
		}
	} else if r.Method == http.MethodDelete {
		h.doDelete(w, ns, svc)
	} else {
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		return
	}
}

func main() {
	log.SetOutput(os.Stdout)
	cli := utils.GetEtcdClient()
	defer cli.Close()

	handler := &etcdHandler{client: cli}

	http.HandleFunc("/", handler.handleFunc)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
