# sysbench安装使用手册
## sysbench介绍
SysBench是一个跨平台且支持多线程的模块化基准测试工具，用于评估系统在运行高负载的数据库时相关核心参数的性能表现。使用SysBench是为了绕过复杂的数据库基准设置，甚至在没有安装数据库的前提下，快速了解数据库系统的性能。

## sysbench安装
```sh
yum install sysbench -y
```

## 参数选择和运行
### 重点参数关注
+ --threads=N                     number of threads to use
+ --time=N                        limit for total execution time in seconds
+ --mysql-host=[LIST,...]          MySQL server host
+ --mysql-port=[LIST,...]          MySQL server port（默认为3306）
+ --mysql-user=STRING              MySQL user
+ --mysql-password=STRING          MySQL password []
+ --mysql-db=STRING                MySQL database name
### 示例
#### 准备数据
```sh
sysbench oltp_write_only --db-ps-mode=auto --mysql-host=XX --mysql-port=XX --mysql-user=XX --mysql-password=XX --mysql-db=sysdb --tables=100 --table_size=40000 --time=300 --report-interval=1 --threads=16 prepare
```
#### 运行
```sh
sysbench oltp_write_only --db-ps-mode=auto --mysql-host=XX --mysql-port=XX --mysql-user=XX --mysql-password=XX --mysql-db=sysdb --tables=100 --table_size=40000 --time=300 --report-interval=1 --threads=16 run
```
### 清理
```sh
sysbench oltp_write_only --db-ps-mode=auto --mysql-host=XX --mysql-port=XX --mysql-user=XX --mysql-password=XX --mysql-db=sysdb --tables=100 --table_size=40000 --time=300 --report-interval=1 --threads=16 cleanup
```

## 性能指标分析
重点关注以下指标
|指标|含义|示例|
|----|----|----|
|transactions|总事务数(每秒事务数)|4283989 (14279.42 per sec.)|
|queries|平均每秒执行次数｜25703962 (85676.60 per sec.)|
|avg|平均响应时间|1.12|
|95th percentile|95%响应时间|1.58|