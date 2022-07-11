# How to control pressure
---
## why we need controlling pressure of benchmark  
  KeenTune提供针对benchmark可能加压不足导致形成性能瓶颈的智能控压功能，主要包括：
● Benchmark选择与适配：预置多款标准benchmark工具，包括iperf3, wrk, sysbench (MySQL), fio等。针对支持的benchmark工具，当前基于专家知识已筛选部分可调参数，以便对benchmark提供的负载压力进行调整。其中可调参数已经内置默认benchmark参数配置。部分benchmark参数，如部署环境的cpu数量，内存大小，可以通过动态探测实际部署环境获得。
● 动态加压：应用KeenTune内置的调优算法，可以针对benchmark的预定调优目标，如吞吐，时延等指标，进行智能化控制，以达到最大压力。


## How to run benchmark pressure controlling  
#### iperf3
```s
keentune param tune --param parameter/iperf.json -i {number of iteration} --bench benchmark/iperf/iperf_bench.json --job {job_name}
```
#### sysbench
```s
keentune param tune --param sysbench.json --bench sysbench_mysql_read_write.json --job {job_name} --iteration {number of iteration}
```
#### wrk
```s
keentune param tune --param wrk.json --bench wrk_nginx_long.json   --job {job_name} --iteration {number of iteration}
```
根据相应的benchmark，需要预先在不同的环境拉起KeenTune服务，如对于iperf3，可在一个环境运行keentuned和keentune-brain服务，在另一个环境运行keentune-bench, keentune-target服务，以针对两个环境测试网络性能。相应的，需要在keentuned所在环境中的keentuned/keentuned_install.sh中修改target ip等配置。

### 智能控压结果查询
  任务发起后，可以通过log日志查看具体任务执行情况。具体查询：
+ 【智能控压中间数据】/var/keentune/data/tuning_data/tuning/{job name}
+ 【智能控压日志】/var/log/keentuned-param-tune-{job id}.log
+ 【智能控压推荐配置】/var/keentune/parameter/{job name}/{job name}_best.json