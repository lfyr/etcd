package etcd

import "time"

const DialTimeout = 5 * time.Second

func GetEtcdEndpoint() []string {
	return []string{"192.168.1.28:2379", "192.168.1.28:12379", "192.168.1.28:22379"}
}
