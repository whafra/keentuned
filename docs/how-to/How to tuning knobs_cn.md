# How to tuning knobs automatically
---  
## What knobs to be tuned
KeenTune默认支持对操作系统内核参数[sysctl](../daemon/examples/parameter/sysctl.json), [nginx运行参数](../daemon/examples/parameter/nginx.json), [mysql运行参数](../daemon/examples/parameter/my_cnf.json)等参数进行auto-tuning. 你还可以[自定义参数域](./customize_knobs.md)来进行auto-tuning.  

我们需要在keentuned.conf文件中定义我们想要调整的参数域  
```conf
[target-group-1]
# sysctl.json, nginx.json or my_cnf.json
# You can define several files and separated by ','
PARAMETER = sysctl.json, nginx.json
```
我们可以使用`keentune param list`命令来查看能使用的参数文件
```s
>>> keentune param list
Parameter List
        iperf.json
        my_cnf.json
        nginx.json
        sysctl.json
```

## Which benchmark we use
Auto-tuning需要一个benchmark工具来对性能进行评估，从而使得算法能够对参数配置进行选择和优化，KeenTune默认支持wrk,fio,iperf, 以及tpce和tpch等benchmark工具. 和参数一样，你也可以通过开发benchmark脚本来[自定义benchmark](./customize_benchmark.md)进行auto-tuning.

我们需要在keentuned.conf文件中定义我们想要使用的benchmark文件
```conf
[bench-group-1]
# use wrk-long benchmark script
BENCH_CONFIG = bench_wrk_nginx_long.json
```
我们也可以使用`keentune param list`命令来查看能使用的参数文件
```s
>>> keentune param list
Benchmark List
        iperf_bench.json
        bench_tpce_db.json
        bench_tpch_db.json
        bench_wrk_nginx_long.json
        bench_wrk_nginx_short.json
```

## Which auto-tuning algorithm
KeenTune在KeenTune-brain组件中内置了Random, TPE和HORD三种基础调优算法, 也可以开发新的算法组件将[自定义的算法](./customize_algorithm.md)加入到KeenTune中。

我们需要在keentuned.conf文件中定义我们想要使用的算法
```conf
[brain]
# use TPE auto-tuning algorithm
AUTO_TUNING_ALGORITHM = tpe
```

## Procedure of auto-tuning
确认好调优参数和发压配置文件后，我们可以使用 keentune param tune 命令来发起动态调优。--job 参数指定任务名，简写为 -j ，必选参数，注意任务名尽量保持唯一性；--iteration 参数指定调优轮次，简写为 -i，可选参数，默认值为100轮；--debug 参数指定为调试模式，可选参数，一般不用。
```s
>>> keentune param tune --job tune_test --iteration 10 --debug
[ok] Running Param Tune Success.

iteration: 10
name: tune_test

        see more details by log file: "/var/log/keentune/keentuned-param-tune-1652172864.log"
```

任务发起后，可以通过log日志查看具体任务执行情况。
```s
>>> cat /var/log/keentune/keentuned-param-tune-1652172864.log
```

当我们不想再继续运行某个动态调优任务时，可使用 keentune param stop 命令来停止正在执行的调优任务，同时会将环境恢复到默认参数配置。
```s
>>> keentune param stop
[Warning] Abort parameter optimization job.
```

动态调优任务跑完后，可通过 keentune param jobs 命令查看调优任务列表，此时可以看到刚跑完的 tune_test 任务已经在任务列表中，如果任务中途失败的话是不会出现在该列表上的。
```s
>>> keentune param jobs
 Tune Jobs
        tune_test
```

在动态调优过程中，如果某个调优任务调优效果比较好的话，我们想把此次调优的最优参数记录下来作为专家配置文件的话，可通过 keentune param dump 命令将最优参数固化为专家配置文件，以供其他用户参考配置。
```s
>>> keentune param dump --job tune_test
[Warning] Dump tune_test has already operated, overwrite? Y(yes)/N(no)y
[ok] dump successfully, file list:
        /var/keentune/profile/tune_test_group1.conf
```

如果我们不想再使用某个调优任务生成的专家配置文件了，可使用 keentune profile delete 命令来删除指定的专家配置文件，需要注意的是对于 keentune 内置的专家配置文件是不支持删除的。
```s
>>> keentune profile delete --name tune_test_group1.conf
[Warning] Are you sure you want to permanently delete job data 'tune_test_group1.conf' ?Y(yes)/N(no)y
[ok] tune_test_group1.conf delete successfully
```

动态调优完成后，如果想让环境恢复到最初的参数配置，可使用 keentune param rollback 命令来回滚参数，以使环境参数恢复到默认设置。
```s
>>> keentune param rollback
[ok] param rollback successfully
```

对于不再使用的任务，可使用 keentune param delete 命令来删除该任务，包括和该任务相关的参数以及生成的各种中间文件都会被删除。
```s
>>> keentune param delete --job tune_test
[Warning] Are you sure you want to permanently delete job data 'tune_test' ?Y(yes)/N(no)y
[ok] tune_test delete successfully
```