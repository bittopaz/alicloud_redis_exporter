package main

import (
	"io/ioutil"
	"net/http"
	"os"

	"encoding/json"

	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
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
	SecurityToken   string
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
	InstanceId string  `json:"instanceId"`
	Average    float64 `json:"Average"`
}

func GetMetricMap(idList []string, metric string) map[string]float64 {
	result := make(map[string]float64)
	defaultRegion, accessKeyID, accessKeySecret, securityToken := getAuth()
	client, err := sdk.NewClientWithStsToken(
		defaultRegion,
		accessKeyID,
		accessKeySecret,
		securityToken,
	)
	cmsClient := cms.Client{
		Client: *client,
	}
	HandleErr(err)
	request := cms.CreateQueryMetricLastRequest()
	request.Project = "acs_kvstore"
	request.Metric = metric
	request.Domain = "metrics.cn-shanghai.aliyuncs.com"
	response, err := cmsClient.QueryMetricLast(request)
	HandleErr(err)
	var re []aliResponse
	HandleErr(json.Unmarshal([]byte(response.Datapoints), &re))
	for _, v := range re {
		if stringInSlice(v.InstanceId, idList) {
			result[v.InstanceId] = v.Average
		}
	}
	time.Sleep(intervals)
	return result
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
	var instanceList []string
	for id := range c.Items() {
		instanceList = append(instanceList, id)
	}
	memoryUsageMap := GetMetricMap(instanceList, "Memoryusage")
	connectionMap := GetMetricMap(instanceList, "ConnectionUsage")
	inratioMap := GetMetricMap(instanceList, "IntranetInRatio")
	outratioMap := GetMetricMap(instanceList, "IntranetOutRatio")
	intranetInMap := GetMetricMap(instanceList, "IntranetIn")
	intranetOutMap := GetMetricMap(instanceList, "IntranetOut")
	failedCountMap := GetMetricMap(instanceList, "FailedCount")
	cpuUsageMap := GetMetricMap(instanceList, "CpuUsage")
	usedMemoryMap := GetMetricMap(instanceList, "UsedMemory")
	for _, id := range instanceList {
		redis := RedisInstance{
			Id:               id,
			MemoryUsage:      memoryUsageMap[id],
			ConnectionUsage:  connectionMap[id],
			IntranetInRatio:  inratioMap[id],
			IntranetOutRatio: outratioMap[id],
			IntranetIn:       intranetInMap[id],
			IntranetOut:      intranetOutMap[id],
			FailedCount:      failedCountMap[id],
			CpuUsage:         cpuUsageMap[id],
			UsedMemory:       usedMemoryMap[id],
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

func getAuth() (defaultRegion string, accessKeyID string, accessKeySecret string, securityToken string) {
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
	json.Unmarshal(*roleMap["SecurityToken"], &a.SecurityToken)
	a.Region = "cn-shanghai"
	return a.Region, a.AccessKeyID, a.AccessKeySecret, a.SecurityToken
}

func removeFromSlice(s []string, target string) []string {
	for i, v := range s {
		if v == target {
			s = append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func HandleErr(err error) {
	if err != nil {
		log.Errorf("ERROR:%v\n", err)
		os.Exit(-1)
	}
}
