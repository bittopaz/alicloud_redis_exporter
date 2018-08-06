package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"encoding/json"

	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/common/log"
)

var (
	intervals = 100 * time.Millisecond

	metricList = []string{
		"MemoryUsage",
		"ConnectionUsage",
		"IntranetInRatio",
		"IntranetOutRatio",
		"IntranetIn",
		"IntranetOut",
		"FailedCount",
		"CpuUsage",
		"UsedMemory",
	}
)

type Auth struct {
	RoleName        string
	Region          string
	AccessKeyID     string
	AccessKeySecret string
}

type RedisInstance struct {
	Id               string  `json:"id"`
	MemoryUsage      float64 `json:"memoryUsage"`
	ConnectionUsage  float64 `json:"connectionUsage"`
	IntranetInRatio  float64 `json:"intranetInRatio"`
	IntranetOutRatio float64 `json:"intranetOutRatio"`
	IntranetIn       float64 `json:"intranetIn"`
	IntranetOut      float64 `json:"intranetOut"`
	FailedCount      float64 `json:"failedCount"`
	CpuUsage         float64 `json:"cpuUsage"`
	UsedMemory       float64 `json:"usedMemory"`
}

type aliResponse struct {
	Average float64
}

func GetValue(InstanceId string, metric string) float64 {
	defaultRegion, accessKeyID, accessKeySecret := getAuth()
	client, err := cms.NewClientWithAccessKey(defaultRegion, accessKeyID, accessKeySecret)
	HandleErr(err)
	request := cms.CreateQueryMetricLastRequest()
	request.Project = "acs_kvstore"
	request.Dimensions = fmt.Sprintf("{\"instanceId\":\"%s\"}", InstanceId)
	request.Metric = metric
	request.Domain = "cms.cn-shanghai.aliyuncs.com"
	response, err := client.QueryMetricLast(request)
	HandleErr(err)
	var re aliResponse
	HandleErr(json.Unmarshal([]byte(strings.Trim(response.Datapoints, "[]")), &re))
	time.Sleep(intervals)
	return re.Average
}

func checkInstanceList(id string) {
	if _, found := c.Get(id); found {
		c.Add(id, nil, time.Hour)
	} else {
		c.Set(id, nil, time.Hour)
	}
}

func StoreData() {
	var redisList []RedisInstance
	instanceList := c.Items()
	for id := range instanceList {
		if id == "redislist" {
			continue
		}
		memoryUsageValue := GetValue(id, "Memoryusage")
		connectionValue := GetValue(id, "ConnectionUsage")
		inratioValue := GetValue(id, "IntranetInRatio")
		outratioValue := GetValue(id, "IntranetOutRatio")
		intranetInValue := GetValue(id, "IntranetIn")
		intranetOutValue := GetValue(id, "IntranetOut")
		failedCountValue := GetValue(id, "FailedCount")
		cpuUsageValue := GetValue(id, "CpuUsage")
		usedMemoryValue := GetValue(id, "UsedMemory")

		redis := RedisInstance{
			Id:               id,
			MemoryUsage:      memoryUsageValue,
			ConnectionUsage:  connectionValue,
			IntranetInRatio:  inratioValue,
			IntranetOutRatio: outratioValue,
			IntranetIn:       intranetInValue,
			IntranetOut:      intranetOutValue,
			FailedCount:      failedCountValue,
			CpuUsage:         cpuUsageValue,
			UsedMemory:       usedMemoryValue,
		}
		redisList = append(redisList, redis)
	}
	c.Set("redislist", &redisList, cache.DefaultExpiration)
}

func readCache(id string) (result RedisInstance) {
	var redisList []RedisInstance
	if x, found := c.Get("redislist"); found {
		redisList = *x.(*[]RedisInstance)
	}

	for _, redis := range redisList {
		if redis.Id == id {
			result = redis
		}
	}
	return result
}

func getAuth() (defaultRegion string, accessKeyID string, accessKeySecret string) {
	var a Auth
	//get rolename
	cmdGetRoleName, err := http.Get("http://100.100.100.200/latest/meta-data/ram/security-credentials/")
	HandleErr(err)
	roleNameRaw, err := ioutil.ReadAll(cmdGetRoleName.Body)
	cmdGetRoleName.Body.Close()
	HandleErr(err)
	a.RoleName = string(roleNameRaw)

	//according to the rolename, get a json file.
	cmdGetJSON, err := http.Get("http://100.100.100.200/latest/meta-data/ram/security-credentials/" + a.RoleName)
	HandleErr(err)
	jsonRaw, err := ioutil.ReadAll(cmdGetJSON.Body)

	//convert json file to map
	var roleMap map[string]*json.RawMessage
	json.Unmarshal(jsonRaw, &roleMap)

	//extract related content from map
	json.Unmarshal(*roleMap["AccessKeyId"], &a.AccessKeyID)
	json.Unmarshal(*roleMap["AccessKeySecret"], &a.AccessKeySecret)
	a.Region = "cn-shanghai"
	return a.Region, a.AccessKeyID, a.AccessKeySecret
}

func HandleErr(err error) {
	if err != nil {
		log.Errorf("ERROR:%v\n", err)
		os.Exit(-1)
	}
}
