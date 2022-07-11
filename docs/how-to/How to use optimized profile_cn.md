# How to use a optimized profile
---  
## What's the optimized profile
KeenTune预置了多个典型的性能优化场景及和常用应用优化配置, 用户可以通过一键设置, 提升应用和系统性能。用户还可以将Auto-Tuning中产生的优化参数配置保存为profile，在多个环境中重复设置。  

## Which profile do we have
使用`keentune profile list`命令，可以查看 KeenTune 中内置的专家配置文件，显示为 available 的文件可以使用在当前环境中使用，显示为 active 的文件表示当前环境已经激活这个文件。
```s
>>> keentune profile list
[active]        cpu_high_load.conf
[available]     io_high_throughput.conf
[available]     mysql_tpcc.conf
[available]     net_high_throuput.conf
[available]     net_low_latency.conf
```

我们可以继续使用`keentune profile info`命令来查看每个专家配置文件的具体内容，包括调优参数域、调优参数取值等信息。目前KeenTune内置的`cpu_high_load.conf`,`io_high_throughput.conf`, `net_high_throuput.conf`, `net_low_latency.conf` 四个profile是分别针对cpu高负载，io高吞吐，网络高吞吐和网络低时延等常见的工作负载和调优需求的内核参数优化profile文件。`mysql_tpcc.conf`是针对tpcc测试对mysql参数和内核参数等进行优化。
```s
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

## Procedure of profile setting and rollback
选择好合适的profile之后，我们使用`keentune profile set`命令使其在当前环境中生效，需要注意的是，在设置参数之前请确认当前环境中有我们要配置的参数，例如如果涉及nginx参数请确保调优环境中已经安装了nginx。  
这里我们将`cpu_high_load.conf`文件参数配置到第一组的所有机器上，44个内核参数全部设置成功。`--group[?]` 参数指定了Target服务群组，设置参数时会给该群组下的所有机器都设置，关于Target Group相关信息参见[《Target Group配置》](./4.target_group_cn.md)
```s
>>> keentune profile set --group1 cpu_high_load.conf
[OK] Set cpu_high_load.conf successfully: [sysctl] successed 44/44
```

设置完参数之后如果对结果不满意，我们还可以使用`keentune profile rollback`命令将参数回滚，使被修改的参数恢复默认值。
```s
>>> keentune profile rollback
[ok] profile rollback successfully
```

KeenTune提供了一个非常有趣的能力让profile与Auto-Tuning共同进行性能优化。Profile中的配置是我们在标准环境中运行Auto-tuning获得的，在其他环境上未必有足够好的表现，我们可以通过动态调优对其进行二次优化。  
首先我们使用`keentune profile generate`命令将专家配置文件转换成用于动态调优的配置文件。并用这个配置文件启动一次[Auto-tuning调优](./3.Auto_tuning_cn.md)。
```s
>>> keentune profile generate --name cpu_high_load.conf --output cpu_high_load.json
[ok] /var/keentune/parameter/generate/cpu_high_load.json generate successfully
```