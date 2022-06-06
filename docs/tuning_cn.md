# KeenTune智能调优
## 智能调优介绍
KeenTune 是一款AI算法与专家知识库双轮驱动的操作系统全栈式智能优化产品，为主流的操作系统提供轻量化、跨平台的一键式性能调优，让应用在智能定制的运行环境发挥最优性能。智能调优是 KeenTune 的一种动态调优，提供内核、编译器、runtime及TOP应用的配置参数集合，使用AI算法识别敏感参数，并进行参数组合的最优配置。

## 使用方法
通过 keentune param -h 命令，可查看与动态调优相关的所有可用命令
```sh
>>> keentune param -h
Dynamic parameter tuning with AI algorithms

Usage:
  keentune param [command] [flags]

Examples:
        keentune param delete --job tune_test
        keentune param dump --job tune_test
        keentune param jobs
        keentune param list
        keentune param rollback
        keentune param stop
        keentune param tune --job tune_test --iteration 10
        keentune param tune --job tune_test

Available Commands:
  delete      Delete the dynamic parameter tuning job
  dump        Dump the parameter tuning result to a profile
  jobs        List parameter optimizing jobs
  list        List parameter and benchmark configuration files
  rollback    Restore initial state
  stop        Terminate a parameter tuning job
  tune        Deploy and start a parameter tuning job

Flags:
  -h, --help   help message

Use "keentune param [command] --help" for more information about a command.
```
使用 keentune param list 命令，查看 keentune 中所支持的调优场景的配置文件列表，主要包括调优参数文件列表和发压配置文件列表，我们可以根据调优的场景来选择对应的调优参数文件和发压配置文件来启动动态调优。
```sh
>>> keentune param list
Parameter List
        iperf.json
        my_cnf.json
        nginx.json
        sysctl.json

Benchmark List
        bench_fio_disk_IOPS.json
        iperf_bench.json
        bench_tpce_db.json
        bench_tpch_db.json
        bench_wrk_nginx_long.json
        bench_wrk_nginx_short.json
```
确认好调优参数和发压配置文件后，我们可以使用 keentune param tune 命令来发起动态调优。--job 参数指定任务名，简写为 -j ，必选参数，注意任务名尽量保持唯一性；--iteration 参数指定调优轮次，简写为 -i，可选参数，默认值为100轮；--debug 参数指定为调试模式，可选参数，一般不用。
```sh
>>> keentune param tune --job tune_test --iteration 10 --debug
[ok] Running Param Tune Success.

iteration: 10
name: tune_test

        see more details by log file: "/var/log/keentune/keentuned-param-tune-1652172864.log"
```
任务发起后，可以通过log日志查看具体任务执行情况。
```sh
>>> cat /var/log/keentune/keentuned-param-tune-1652172864.log
```
当我们不想再继续运行某个动态调优任务时，可使用 keentune param stop 命令来停止正在执行的调优任务，同时会将环境恢复到默认参数配置。
```sh
>>> keentune param stop
[Warning] Abort parameter optimization job.
```
动态调优任务跑完后，可通过 keentune param jobs 命令查看调优任务列表，此时可以看到刚跑完的 tune_test 任务已经在任务列表中，如果任务中途失败的话是不会出现在该列表上的。
```sh
>>> keentune param jobs
 Tune Jobs
        tune_test
```
在动态调优过程中，如果某个调优任务调优效果比较好的话，我们想把此次调优的最优参数记录下来作为专家配置文件的话，可通过 keentune param dump 命令将最优参数固化为专家配置文件，以供其他用户参考配置。
```sh
>>> keentune param dump --job tune_test
[Warning] Dump tune_test has already operated, overwrite? Y(yes)/N(no)y
[ok] dump successfully, file list:
        /var/keentune/profile/tune_test_group1.conf
```
如果我们不想再使用某个调优任务生成的专家配置文件了，可使用 keentune profile delete 命令来删除指定的专家配置文件，需要注意的是对于 keentune 内置的专家配置文件是不支持删除的。
```sh
>>> keentune profile delete --name tune_test_group1.conf
[Warning] Are you sure you want to permanently delete job data 'tune_test_group1.conf' ?Y(yes)/N(no)y
[ok] tune_test_group1.conf delete successfully
```
动态调优完成后，如果想让环境恢复到最初的参数配置，可使用 keentune param rollback 命令来回滚参数，以使环境参数恢复到默认设置。
```sh
>>> keentune param rollback
[ok] param rollback successfully
```
对于不再使用的任务，可使用 keentune param delete 命令来删除该任务，包括和该任务相关的参数以及生成的各种中间文件都会被删除。
```sh
>>> keentune param delete --job tune_test
[Warning] Are you sure you want to permanently delete job data 'tune_test' ?Y(yes)/N(no)y
[ok] tune_test delete successfully
```