package discovery

import (
	"context"
	"etcd/etcd"
	"fmt"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"sync"
)

type Service struct {
	Name     string
	IP       string
	Port     string
	Protocol string
}

func ServiceRegister(s *Service) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcd.GetEtcdEndpoint(),
		DialTimeout: etcd.DialTimeout,
	})
	if err != nil {
		// handle error!
		log.Fatalln(err)
	}
	defer cli.Close()

	var grantLease bool
	var leaseID clientv3.LeaseID
	ctx := context.Background()

	getRes, err := cli.Get(ctx, s.Name, clientv3.WithCountOnly())
	if err != nil {
		log.Fatalln(err)
	}

	// 判断是否已经注册
	if getRes.Count == 0 {
		grantLease = true
	}

	// 申请租约
	if grantLease {
		leaseRes, err := cli.Grant(ctx, 10)
		if err != nil {
			log.Fatalln(err)
		}
		leaseID = leaseRes.ID
	}

	kv := clientv3.NewKV(cli)
	txn := kv.Txn(ctx)
	_, err = txn.If(clientv3.Compare(clientv3.CreateRevision(s.Name), "=", 0)).
		Then(
			clientv3.OpPut(s.Name, s.Name, clientv3.WithLease(leaseID)),
			clientv3.OpPut(s.Name+".ip", s.IP, clientv3.WithLease(leaseID)),
			clientv3.OpPut(s.Name+".port", s.Port, clientv3.WithLease(leaseID)),
			clientv3.OpPut(s.Name+".protocol", s.Protocol, clientv3.WithLease(leaseID)),
		).
		Else(
			clientv3.OpPut(s.Name, s.Name, clientv3.WithIgnoreLease()),
			clientv3.OpPut(s.Name+".ip", s.IP, clientv3.WithIgnoreLease()),
			clientv3.OpPut(s.Name+".port", s.Port, clientv3.WithIgnoreLease()),
			clientv3.OpPut(s.Name+".protocol", s.Protocol, clientv3.WithIgnoreLease()),
		).
		Commit()
	if err != nil {
		log.Fatalln(err)
	}

	if grantLease {
		// 监听租约
		leaseKeepAlive, err := cli.KeepAlive(ctx, leaseID)
		if err != nil {
			log.Fatalln(err)
		}
		for lease := range leaseKeepAlive {
			fmt.Printf("leaseID:%x,ttl:%d\n", lease.ID, lease.TTL)
		}
	}
}

type Services struct {
	services map[string]*Service
	sync.RWMutex
}

var myServices = &Services{
	services: map[string]*Service{},
}

func ServiceDiscovery(svcName string) *Service {
	var s *Service = nil
	myServices.RLock()
	s, _ = myServices.services[svcName]
	myServices.RUnlock()
	return s
}

func WatchServiceName(svcName string) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcd.GetEtcdEndpoint(),
		DialTimeout: etcd.DialTimeout,
	})
	if err != nil {
		// handle error!
		log.Fatalln(err)
		return
	}
	defer cli.Close()

	getRes, err := cli.Get(context.Background(), svcName, clientv3.WithPrefix())
	if err != nil {
		log.Fatalln(err)
	}
	if getRes.Count > 0 {
		mp := sliceToMap(getRes.Kvs)
		s := &Service{}
		if kv, ok := mp[svcName]; ok {
			s.Name = string(kv.Value)
		}

		if kv, ok := mp[svcName+".ip"]; ok {
			s.IP = string(kv.Value)
		}

		if kv, ok := mp[svcName+".port"]; ok {
			s.Port = string(kv.Value)
		}

		if kv, ok := mp[svcName+".protocol"]; ok {
			s.Protocol = string(kv.Value)
		}

		myServices.Lock()
		myServices.services[svcName] = s
		myServices.Unlock()
	}
	rch := cli.Watch(context.Background(), svcName, clientv3.WithPrefix())
	for wres := range rch {
		for _, ev := range wres.Events {
			switch ev.Type {
			case clientv3.EventTypeDelete:
				myServices.Lock()
				delete(myServices.services, svcName)
				myServices.Unlock()
			case clientv3.EventTypePut:
				myServices.Lock()
				if _, ok := myServices.services[svcName]; !ok {
					myServices.services[svcName] = &Service{}
				}
				switch string(ev.Kv.Key) {
				case svcName:
					myServices.services[svcName].Name = string(ev.Kv.Value)
				case svcName + ".ip":
					myServices.services[svcName].IP = string(ev.Kv.Value)
				case svcName + ".port":
					myServices.services[svcName].Port = string(ev.Kv.Value)
				case svcName + ".protocol":
					myServices.services[svcName].Protocol = string(ev.Kv.Value)
				}
				myServices.Unlock()
			}
		}
	}
}

func sliceToMap(list []*mvccpb.KeyValue) map[string]*mvccpb.KeyValue {
	mp := make(map[string]*mvccpb.KeyValue, 0)
	for _, item := range list {
		mp[string(item.Key)] = item
	}
	return mp
}
