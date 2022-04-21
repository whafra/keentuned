[English](./keentuned/README.md)| [简体中文](./keentuned/README_ch.md) 

# KeenTune
## Introduction
---
KeenTune is a Linux full-stack intelligent optimization product with both AI algorithm and expert knowledge. It provides lightweight and cross-platform performance tuning service with easy operations, to provide a customized and optimized running environment for optimum performance of the application。

KeenTuned is the management component of KeenTune, including CLI and Daemon. CLI provides command line interfaces to users, which includes three kinds of commands: basic commands, static tuning related commands, and dynamic tuning related commands. Keentuned，as the management and control module, is responsible for monitoring other components, receiving and parsing commands from CLI, and scheduling related components in workflows.

## Installation
---
### 1. install requirements
```sh
$ sudo apt-get install python-setuptools golang -y
or
$ sudo yum install python-setuptools go -y

go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
```

### 2. install keentune & keentuned
```bash
$ cd keentuned
$ sudo sh keentuned_install.sh
```

## Quick Start
---
### Run keentune-target, keentune-brain, keentune-bench
vm 1
```shell
$ keentune-target
```
vm 2
```shell
$ keentune-brain
```
vm 3
```shell
$ keentune-bench
```

### Modify keentune configuration
```shell
$ vim /etc/keentune/conf/keentuned.conf

TARGET_IP = [ip of vm 1]
BRAIN_IP  = [ip of vm 2]
BENCH_IP  = [ip of vm 3]
```

### Run keentuned
```shell
$ keentuned
```

### Param Tune
keentune param tune --param [param conf] --bench [bench conf] --name [job name]   --iteration [number of iteration]  
e.g.  
```bash  
$ keentune param tune --param param_100.json --bench benchmark/wrk/bench_wrk_nginx_long.json --name tune_test --iteration 10 
```

Check the execution of the task by 'msg' command.
```bash  
$ keentune msg
```

### Profile Set
keentune profile set [profile name]  
e.g.
```bash
$ keentune profile set cpu_high_load.conf
```

## Code structure
---  
+ cli: KeenTune Console
+ daemon
    + api: API Definition
    + common: General method module
    + example: example of param conf and benchmark conf
    + modules: Function module

## Documentation
---
For more information, please refer to keentune help information or [KeenTune User Guide].
