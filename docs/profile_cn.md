# Keentune专家调优
## 专家调优介绍
KeenTune 是一款AI算法与专家知识库双轮驱动的操作系统全栈式智能优化产品，为主流的操作系统提供轻量化、跨平台的一键式性能调优，让应用在智能定制的运行环境发挥最优性能。专家调优是 KeenTune 的一种静态调优，预置多款典型高负载场景及应用的经验优化配置，用户可以通过一键设置，即可提升环境性能；该功能还可应用于将小规模环境调优出来的经验设置在大集群中扩展使用。  

## 使用方法
通过 keentune profile -h 命令，可查看与静态调优相关的所有可用命令
```sh
>>> keentune profile -h
Static tuning with expert profiles

Usage:
keentune profile [command] [flags]

Examples:
keentune profile delete --name tune_test.conf
keentune profile generate --name tune_test.conf --output gen_param_test.json
keentune profile info --name cpu_high_load.conf
keentune profile list
keentune profile rollback
keentune profile set --group1 cpu_high_load.conf

Available Commands:
delete      Delete a profile
generate    Generate a parameter configuration file from profile
info        Show information of the specified profile
list        List all profiles
rollback    Restore initial state
set         Apply a profile to the target machine

Flags:
-h, --help   help message

Use "keentune profile [command] --help" for more information about a command.
```
使用 keentune profile list 命令，查看 keentune 中内置的专家配置文件，显示为 available 的文件可以使用在当前环境中使用，显示为 active 的文件表示当前环境已经激活这个文件。  
```sh
>>> keentune profile list
[active]        cpu_high_load.conf
[available]     io_high_throughput.conf
[available]     mysql_tpcc.conf
[available]     net_high_throuput.conf
[available]     net_low_latency.conf
```
我们可以继续使用 keentune profile info 命令来查看每个专家配置文件的具体内容，包括调优参数域、调优参数取值等信息。以下为 cpu_high_load.conf 调优文件的详细内容，可以看到 cpu_high_load.conf 主要对内核参数(sysctl)进行了配置。  
```sh
>>> keentune profile info --name cpu_high_load.conf
[sysctl]
net.core.netdev_budget: 200
net.core.optmem_max: 143360
net.core.wmem_max: 61865984
net.core.wmem_default: 122880
kernel.sched_latency_ns: 40000000
kernel.sched_min_granularity_ns: 17000000
kernel.sched_wakeup_granularity_ns: 43000000
kernel.sched_cfs_bandwidth_slice_us: 24000
kernel.sched_migration_cost_ns: 4200000
kernel.sched_nr_migrate: 74
kernel.numa_balancing: 1
vm.watermark_scale_factor: 130
vm.min_free_kbytes: 634880
vm.swappiness: 7
kernel.pid_max: 3145728
kernel.shmmni: 7168
kernel.shmmax: 34359740000
kernel.shmall: 4294967200
kernel.core_uses_pid: 0
kernel.msgmni: 96000
kernel.msgmax: 32768
kernel.msgmnb: 466944
kernel.sem: 64000 2048000000 1000 64000
kernel.hung_task_timeout_secs: 720
kernel.nmi_watchdog: 0
kernel.sched_rt_runtime_us: 970000
kernel.timer_migration: 0
kernel.threads-max: 37355520
kernel.sysrq: 0
kernel.sched_autogroup_enabled: 1
kernel.randomize_va_space: 1
kernel.dmesg_restrict: 1
vm.overcommit_ratio: 50
vm.overcommit_memory: 0
vm.page-cluster: 6
vm.max_map_count: 3400000
vm.zone_reclaim_mode: 2
vm.drop_caches: 3
fs.inotify.max_user_watches: 32768
fs.nr_open: 516095
fs.file-max: 2048000
fs.aio-max-nr: 5529600
fs.inotify.max_user_instances: 18560
fs.suid_dumpable: 2
```
确认过专家配置文件的内容之后，我们使用 keentune profile set 命令使其在当前环境中生效，需要注意的是，在设置参数之前请确认当前环境中有我们要配置的参数，例如如果涉及nginx参数请确保调优环境中已经安装了nginx。这里我们将 cpu_high_load.conf 文件参数配置到第一组的所有机器上，44个内核参数全部设置成功。--groupx 参数指定了target服务群组，设置参数时会给该群组下的所有机器都设置，必选参数。
```sh
>>> keentune profile set --group1 cpu_high_load.conf
[OK] Set cpu_high_load.conf successfully: [sysctl] successed 44/44
```
设置完参数之后如果对结果不满意，我们还可以使用 keentune profile rollback 命令将参数回滚，使被修改的参数恢复默认值。
```sh
>>> keentune profile rollback
[ok] profile rollback successfully
```
keentune提供了一个非常酷的能力，如果对当前的专家配置不满意的话，我们可以通过动态调优对其进行二次优化，第一步就是使用 keentune profile generate 命令将专家配置文件转换成用于动态调优的配置文件。我们将 cpu_high_load.conf 文件转换成了 cpu_high_load.json 配置文件，这个文件可以直接用于启动一次动态调优。
```sh
>>> keentune profile generate --name cpu_high_load.conf --output cpu_high_load.json
[ok] /var/keentune/parameter/generate/cpu_high_load.json generate successfully
```