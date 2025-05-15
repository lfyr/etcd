package etcd

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
)

func KvDemo() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   GetEtcdEndpoint(),
		DialTimeout: DialTimeout,
	})
	if err != nil {
		// handle error!
		log.Fatalln(err)
	}
	defer cli.Close()

	_, err = cli.Put(context.Background(), "key1", "value1")
	if err != nil {
		log.Fatalln(err)
	}

	putRes, err := cli.Put(context.Background(), "key1", "value11", clientv3.WithPrevKV())
	if err != nil {
		log.Fatalln(err)
	}

	if putRes.PrevKv != nil {
		log.Println("key1 old value:", putRes.PrevKv)
	}

	getRes, err := cli.Get(context.Background(), "key", clientv3.WithPrefix()) // WithPrefix 范围操作
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(getRes)

	response, err := cli.Delete(context.Background(), "key", clientv3.WithPrefix())
	if err != nil {
		return
	}
	fmt.Println(response.Deleted)
}
