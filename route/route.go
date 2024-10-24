package route

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"go.etcd.io/etcd/client/v3"
)

// PathMappings 存储路由映射
var PathMappings = map[string]string{}

func InitEtcd() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"http://127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	// Fetch initial mappings from etcd
	resp, err := cli.Get(context.Background(), "/mall/route", clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
	}

	// 解析 JSON 映射
	for _, kv := range resp.Kvs {
		var mappings map[string]string
		err := json.Unmarshal(kv.Value, &mappings)
		if err != nil {
			log.Printf("Error unmarshalling JSON: %v", err)
			continue
		}
		// 将映射合并到 PathMappings
		for k, v := range mappings {
			PathMappings[k] = v
		}
	}

	// Watch for changes
	watchResp := cli.Watch(context.Background(), "/mall/route", clientv3.WithPrefix())
	go func() {
		for wresp := range watchResp {
			for _, ev := range wresp.Events {
				if ev.Type == clientv3.EventTypePut {
					// 新增或更新映射
					var newMapping map[string]string
					err := json.Unmarshal(ev.Kv.Value, &newMapping)
					if err != nil {
						log.Printf("Error unmarshalling JSON: %v", err)
						continue
					}
					for k, v := range newMapping {
						PathMappings[k] = v
					}
				} else if ev.Type == clientv3.EventTypeDelete {
					// 删除映射
					delete(PathMappings, string(ev.Kv.Key))
				}
			}
		}
	}()
}
