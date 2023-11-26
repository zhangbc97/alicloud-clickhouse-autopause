package main

import (
	pb "alicloud-clickhouse-autopause/clickhouse"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"sync"
	"time"
)

var globalLock = sync.Mutex{}

type DBInstanceStatus struct {
	RegionID     string
	DBInstanceID string

	DBInstanceAttribute pb.DescribeDBInstanceAttributeResponse

	lastConnTime time.Time
}

var instances = map[string]DBInstanceStatus{}

// GetDBInstanceConfig 给定RegionID和DBInstanceID，返回DBInstanceConfig
func GetDBInstanceConfig(config *ServerConfig, regionID string, dbInstanceID string) *DBInstance {
	for _, config := range config.DbInstances {
		if config.RegionID == regionID && config.DBInstanceID == dbInstanceID {
			return &config
		}
	}

	return nil
}

type server struct {
	pb.UnimplementedAliYunClickhouseServer

	config *ServerConfig
}

// SyncInstanceStatusTicket 启动一个定时任务，定时调用DescribeDBInstanceAttribute，更新数据库实例状态
func SyncInstanceStatusTicket(config *ServerConfig, instances map[string]DBInstanceStatus) {

	log.Default().Println("SyncInstanceStatusTicket")
	ticker := time.NewTicker(time.Duration(config.SyncStatusIntervalSeconds) * time.Second)

	go func(ticker *time.Ticker) {
		for {
			select {
			case <-ticker.C:
				globalLock.Lock()
				log.Default().Println("Sync instance status")
				for _, instance := range instances {
					instanceConfig := GetDBInstanceConfig(config, instance.DBInstanceAttribute.Data.RegionID, instance.DBInstanceAttribute.Data.DBInstanceID)
					// 调用DescribeDBInstanceAttribute
					res, err := pb.DescribeDBInstanceAttribute(instanceConfig.AccessKeyID, instanceConfig.AccessKeySecret, instanceConfig.RegionID, instanceConfig.DBInstanceID)

					if err != nil {
						log.Default().Println(err)
						continue
					}

					instance.DBInstanceAttribute = res
				}
				globalLock.Unlock()
			}
		}
	}(ticker)
}

func StopInstanceTimer(config *ServerConfig, instances map[string]DBInstanceStatus) {
	log.Default().Println("StopInstanceTimer")
	ticker := time.NewTicker(time.Duration(config.IdleCheckIntervalSeconds) * time.Second)

	go func(ticker *time.Ticker) {
		for {
			select {
			case <-ticker.C:
				globalLock.Lock()
				// 超过一定时间没有请求，就停止数据库实例
				log.Default().Println("Check instance status")
				for _, instance := range instances {
					if time.Now().Sub(instance.lastConnTime).Seconds() > float64(config.IdleSecondsBeforeStop) && instance.DBInstanceAttribute.Data.Status == "RUNNING" {
						instanceConfig := GetDBInstanceConfig(config, instance.DBInstanceAttribute.Data.RegionID, instance.DBInstanceAttribute.Data.DBInstanceID)
						// 调用StopDBInstance
						log.Default().Println("Start stop instance: %s, last conn time: %s", instanceConfig.DBInstanceID, instance.lastConnTime.String())
						_, err := pb.StopDBInstance(instanceConfig.AccessKeyID, instanceConfig.AccessKeySecret, instanceConfig.RegionID, instanceConfig.DBInstanceID, config.WaitStatusIntervalSeconds)

						if err != nil {
							log.Default().Println(err)
							continue
						}

						log.Default().Println("Stop instance success: %s", instanceConfig.DBInstanceID)
					}
				}
				globalLock.Unlock()
			}
		}
	}(ticker)
}

func (s *server) K(ctx context.Context, in *pb.KeepAliveRequest) (*pb.KeepAliveResponse, error) {

	globalLock.Lock()

	defer globalLock.Unlock()

	// 配置文件中查找DBInstanceConfig
	instance, ok := instances[in.DBInstanceID]

	if !ok {
		// 获取DBInstanceConfig
		config := GetDBInstanceConfig(s.config, in.RegionID, in.DBInstanceID)

		if config == nil {
			return &pb.KeepAliveResponse{Success: false}, nil
		}

		// 调用DescribeDBInstanceAttribute
		res, err := pb.DescribeDBInstanceAttribute(config.AccessKeyID, config.AccessKeySecret, config.RegionID, config.DBInstanceID)

		if err != nil {
			return &pb.KeepAliveResponse{Success: false}, nil
		}

		instance.DBInstanceAttribute = res
	}
	if instance.DBInstanceAttribute.Data.Status == "RUNNING" {
		instance.lastConnTime = time.Now()
		return &pb.KeepAliveResponse{Success: true}, nil
	} else {
		config := GetDBInstanceConfig(s.config, in.RegionID, in.DBInstanceID)
		// 走到这不需要判断config是否为空，因为上面已经判断过了
		log.Default().Println("Start start instance: %s", config.DBInstanceID)
		_, err := pb.StartDBInstance(config.AccessKeyID, config.AccessKeySecret, config.RegionID, config.DBInstanceID, s.config.WaitStatusIntervalSeconds)

		if err != nil {
			return &pb.KeepAliveResponse{Success: false}, nil
		}
	}
	instance.lastConnTime = time.Now()
	return &pb.KeepAliveResponse{Success: true}, nil
}

func main() {
	serverConfig, err := ReadConfig("config.yaml")

	if err != nil {
		log.Fatalf("ReadConfig failed, err: %v", err)
	}

	// 把两个定时器都启动
	SyncInstanceStatusTicket(&serverConfig, instances)
	StopInstanceTimer(&serverConfig, instances)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", serverConfig.Port))

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	pb.RegisterAliYunClickhouseServer(s, &server{
		config: &serverConfig,
	})

	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
