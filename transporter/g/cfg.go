package g

import (
	"encoding/json"
	"github.com/toolkits/file"
	"log"
	"strings"
	"sync"
)

type RpcConfig struct {
	Enabled bool
	Listen  string
}

type CheckerConfig struct {
	Enabled            bool
	Batch              int
	ConnTimeout        int
	CallTimeout        int
	MaxConcurrentConns int
	MaxIdle            int
	Replicas           int
	Cluster            map[string]string
	ClusterList        map[string]*ClusterNode
}

type InfluxDBConfig struct {
	Enabled            bool
	MaxConcurrentConns int
	Batch              int
	Address            string
	UserName           string
	Password           string
	Database           string
}

type GlobalConfig struct {
	MinStep  int
	Rpc      *RpcConfig
	Checker  *CheckerConfig  //!< policyChecker
	InfluxDB *InfluxDBConfig //!< influxdb
}

var (
	ConfigFile string
	config     *GlobalConfig
	configLock = new(sync.RWMutex)
)

func Config() *GlobalConfig {
	configLock.RLock()
	defer configLock.RUnlock()
	return config
}

func ParseConfig(cfg string) {
	if cfg == "" {
		log.Fatalln("use -c to specify configuration file")
	}

	if !file.IsExist(cfg) {
		log.Fatalln("config file:", cfg, "is not existent. maybe you need `mv cfg.example.json cfg.json`")
	}

	ConfigFile = cfg

	configContent, err := file.ToTrimString(cfg)
	if err != nil {
		log.Fatalln("read config file:", cfg, "fail:", err)
	}

	var c GlobalConfig
	err = json.Unmarshal([]byte(configContent), &c)
	if err != nil {
		log.Fatalln("parse config file:", cfg, "fail:", err)
	}

	// split cluster config
	c.Checker.ClusterList = formatClusterItems(c.Checker.Cluster)

	configLock.Lock()
	defer configLock.Unlock()
	config = &c

	log.Println("read config file:", cfg, "successfully")
}

type ClusterNode struct {
	Addrs []string `json:"addrs"`
}

func NewClusterNode(addrs []string) *ClusterNode {
	return &ClusterNode{addrs}
}

// map["node"]="host1,host2" --> map["node"]=["host1", "host2"]
func formatClusterItems(cluster map[string]string) map[string]*ClusterNode {
	ret := make(map[string]*ClusterNode)
	for node, clusterStr := range cluster {
		items := strings.Split(clusterStr, ",")
		nitems := make([]string, 0)
		for _, item := range items {
			nitems = append(nitems, strings.TrimSpace(item))
		}
		ret[node] = NewClusterNode(nitems)
	}

	return ret
}
