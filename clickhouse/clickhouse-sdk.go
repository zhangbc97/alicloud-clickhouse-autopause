package clickhouse

import (
	"encoding/json"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"log"
	"time"
)

func CallClickhouseAPI(ak string, sk string, apiName string, regionId string, DBInstanceId string, v any) error {
	config := sdk.NewConfig()

	credential := credentials.NewAccessKeyCredential(ak, sk)

	client, err := sdk.NewClientWithOptions(regionId, config, credential)
	if err != nil {
		panic(err)
	}

	request := requests.NewCommonRequest()

	request.Method = "POST"
	request.Scheme = "https" // https | http
	request.Domain = "clickhouse.aliyuncs.com"
	request.Version = "2023-05-22"
	request.ApiName = apiName
	request.QueryParams["RegionId"] = regionId
	request.QueryParams["DBInstanceId"] = DBInstanceId

	resp, err := client.ProcessCommonRequest(request)

	if err != nil {
		return err
	}

	if err := json.Unmarshal(resp.GetHttpContentBytes(), v); err != nil {
		return err
	}

	return nil

}

type DescribeDBInstanceAttributeResponse struct {
	RequestID string `json:"RequestId"`
	Data      struct {
		Description              string        `json:"Description"`
		EngineMinorVersion       string        `json:"EngineMinorVersion"`
		LatestEngineMinorVersion string        `json:"LatestEngineMinorVersion"`
		MaintainEndTime          string        `json:"MaintainEndTime"`
		DBInstanceID             string        `json:"DBInstanceId"`
		Bid                      string        `json:"Bid"`
		Engine                   string        `json:"Engine"`
		MaintainStartTime        string        `json:"MaintainStartTime"`
		Tags                     []interface{} `json:"Tags"`
		Status                   string        `json:"Status"`
		EngineVersion            string        `json:"EngineVersion"`
		ZoneID                   string        `json:"ZoneId"`
		VSwitchID                string        `json:"VSwitchId"`
		CreateTime               time.Time     `json:"CreateTime"`
		ScaleMax                 int           `json:"ScaleMax"`
		LockMode                 int           `json:"LockMode"`
		Nodes                    []struct {
			Ccu        int    `json:"CCU"`
			Memory     int    `json:"Memory"`
			CPUCores   int    `json:"CpuCores"`
			String     string `json:"String"`
			NodeStatus string `json:"NodeStatus"`
		} `json:"Nodes"`
		VpcID      string `json:"VpcId"`
		ChargeType string `json:"ChargeType"`
		ScaleMin   int    `json:"ScaleMin"`
		RegionID   string `json:"RegionId"`
		ExpireTime string `json:"ExpireTime"`
		AliUID     int64  `json:"AliUid"`
	} `json:"Data"`
}
type StopDBInstanceResponse struct {
	RequestID string `json:"RequestId"`
	Data      struct {
		TaskID         int    `json:"TaskId"`
		DBInstanceID   int    `json:"DBInstanceID"`
		DBInstanceName string `json:"DBInstanceName"`
	} `json:"Data"`
}

type StartDBInstanceResponse struct {
	RequestID string `json:"RequestId"`
	Data      struct {
		TaskID         int    `json:"TaskId"`
		DBInstanceID   int    `json:"DBInstanceID"`
		DBInstanceName string `json:"DBInstanceName"`
	} `json:"Data"`
}

func DescribeDBInstanceAttribute(ak string, sk string, regionId string, DBInstanceId string) (DescribeDBInstanceAttributeResponse, error) {
	res := DescribeDBInstanceAttributeResponse{}
	err := CallClickhouseAPI(ak, sk, "DescribeDBInstanceAttribute", regionId, DBInstanceId, &res)
	return res, err
}

func WaitForInstanceStatus(ak string, sk string, regionId string, DBInstanceId string, status string, maxWaitSeconds int) (bool, error) {
	startTime := time.Now()

	for {
		res, err := DescribeDBInstanceAttribute(ak, sk, regionId, DBInstanceId)
		if err != nil {
			return false, err
		}
		log.Default().Println("Instance status is " + res.Data.Status)
		if res.Data.Status == status {
			break
		}
		time.Sleep(5 * time.Second)
		if time.Now().Sub(startTime).Seconds() > float64(maxWaitSeconds) {
			return false, nil
		}

	}
	return true, nil
}

func StopDBInstance(ak string, sk string, regionId string, DBInstanceId string, maxWaitSeconds int) (bool, error) {
	res := StopDBInstanceResponse{}
	err := CallClickhouseAPI(ak, sk, "StopDBInstance", regionId, DBInstanceId, &res)

	if err != nil {
		return false, err
	}

	success, err := WaitForInstanceStatus(ak, sk, regionId, DBInstanceId, "STOPPED", maxWaitSeconds)

	if err != nil {
		return false, err
	}

	if success {
		return true, nil
	} else {
		return false, nil
	}

}
func StartDBInstance(ak string, sk string, regionId string, DBInstanceId string, maxWaitSeconds int) (bool, error) {
	// 只有状态是STOPPED的时候才能执行启动操作
	// 判断启动成功的标志是后端的HTTP状态码是200，实测接口延迟比较大，所以启动是否成功不用接口数据，后端通了就可以把流量切过去了

	// 从接口更新状态
	res, err := DescribeDBInstanceAttribute(ak, sk, regionId, DBInstanceId)

	if err != nil {
		return false, err
	}

	if res.Data.Status == "RUNNING" {
		return true, nil
	}

	// 执行开启动作
	startResponse := StartDBInstanceResponse{}
	err = CallClickhouseAPI(ak, sk, "StartDBInstance", regionId, DBInstanceId, &startResponse)

	if err != nil {
		return false, err
	}

	success, err := WaitForInstanceStatus(ak, sk, regionId, DBInstanceId, "RUNNING", maxWaitSeconds)

	if err != nil {
		return false, err
	}

	if success {
		return true, nil
	} else {
		return false, nil
	}

}
