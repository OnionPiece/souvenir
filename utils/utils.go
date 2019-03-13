package utils

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"go.etcd.io/etcd/clientv3"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

func getEtcdTLSConfig() *tls.Config {
	certFile := flag.String("cert", os.Getenv("ETCD_CERT"), "A PEM eoncoded certificate file.")
	keyFile := flag.String("key", os.Getenv("ETCD_KEY"), "A PEM encoded private key file.")
	caFile := flag.String("CA", os.Getenv("ETCD_CA"), "A PEM eoncoded CA's certificate file.")

	cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		log.Fatal("Fail to load etcd-client cert and key, since: ", err)
	}

	caCert, err := ioutil.ReadFile(*caFile)
	if err != nil {
		log.Fatal("Fail to load etcd ca file, since: ", err)
	}
	rootCAs := x509.NewCertPool()
	rootCAs.AppendCertsFromPEM(caCert)

	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      rootCAs,
	}
	return cfg
}

func GetEtcdClient() *clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(os.Getenv("ETCD_ENDPOINTS"), ","),
		DialTimeout: 5 * time.Second,
		TLS:         getEtcdTLSConfig(),
	})
	if err != nil {
		log.Fatal("Fail to create etcd clientv3, since: %v", err)
	}
	return cli
}

func GetKeys(cli *clientv3.Client, prefix string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	res, err := cli.Get(ctx, prefix, clientv3.WithPrefix(), clientv3.WithKeysOnly())
	cancel()
	if err != nil {
		return "", err
	}
	if len(res.Kvs) == 0 {
		return "", nil
	} else {
		ret := make([]string, len(res.Kvs))
		ret = ret[:0]
		for _, kv := range res.Kvs {
			ret = append(ret, string(kv.Key))
		}
		return string(strings.Join(ret, ",")), nil
	}
}

func GetValue(cli *clientv3.Client, k string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	res, err := cli.Get(ctx, k)
	cancel()
	if err != nil {
		return "", err
	}
	if len(res.Kvs) == 0 {
		return "", nil
	} else {
		return string(res.Kvs[0].Value), nil
	}
}

func PutKeyValue(cli *clientv3.Client, k, v string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	_, err := cli.Put(ctx, k, v)
	cancel()
	return err
}

func DeleteKey(cli *clientv3.Client, k string, withPrefix bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	ops := []clientv3.OpOption{}
	if withPrefix {
		ops = append(ops, clientv3.WithPrefix())
	}
	_, err := cli.Delete(ctx, k, ops...)
	cancel()
	return err
}
