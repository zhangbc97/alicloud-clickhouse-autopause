# alicloud-clickhouse-autopause

阿里云Clickhouse企业版自动启停服务

## 简介

阿里云目前已推出Clickhouse企业版，企业版采用存算分离架构，在不使用时可通过暂停实例来节省计算资源的费用。  
目前官方还不支持自动启停，本服务可以实现自动启停功能。

## 实现原理

本服务基于gRPC实现一个KeepAlive接口，在存在请求时需持续调用该接口进行续租，超过一定时间未收到续租请求则自动暂停实例。  
在实例暂停后，如果有请求则自动启动实例，在启动实例的过程中将不会返回，直到实例启动成功或者启动失败。

## 使用方式

- 编写配置文件

```yaml
port: 80

idleSecondsBeforeStop: 300       # 暂停服务前的空闲时间
idleCheckIntervalSeconds: 10     # 检查是否需要暂停服务的时间间隔
syncStatusIntervalSeconds: 60    # 定时同步最新状态的时间间隔
waitStatusIntervalSeconds: 300    # 启停操作等待的时间

# 在这里完善Clickhouse配置
dbInstances:
  - regionID: 'cn-beijing'
    dbInstanceID: ''

    accessKeyID: ''
    accessKeySecret: ''

    serverAddress: ''
    serverHttpPort: 8123
    serverTcpPort: 9000

```

- `docker run -d --name=alicloud-clickhouse-autopause -v /path/to/config.yaml:/config.yaml -p 80:80 zhangbc/alicloud-clickhouse-autopause:latest`

## 使用限制

- 目前服务只支持单点部署
- 该服务目前存在大量的全局锁，建议降低对KeepAlive接口的调用频率
