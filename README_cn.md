[English](./keentuned/README.md)| [简体中文](./keentuned/README_cn.md) 

# KeenTune
## 简介
---
KeenTune 是一款AI算法与专家知识库双轮驱动的操作系统全栈式智能优化产品，为主流的操作系统提供轻量化、跨平台的一键式性能调优，让应用在智能定制的运行环境发挥最优性能。

KeenTuned 是KeenTune的调度管理组件，包含CLI和Deamon两个部分。CLI模块提供用户可用的命令行接口，命令分为基础命令、静态调优相关命令、动态调优相关命令三个部分。keentuned作为核心管控模块，负责监控其他组件、接收解析来自CLI的命令、按照业务处理顺序调度相关组件等。

## 安装
---
### 1. 安装依赖系统软件包
```sh
$ sudo apt-get install python-setuptools golang -y
or
$ sudo yum install python-setuptools go -y

go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
```
### 2. 安装 keentune & keentuned
```bash
$ cd keentuned
$ sudo sh keentuned_install.sh
```

## 快速使用指南
---
### 启动服务
#### 启动 keentune-target, keentune-brain, keentune-bench
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
#### 修改keentune配置
```shell
$ vim /etc/keentune/conf/keentuned.conf

TARGET_IP = [ip of vm 1]
BRAIN_IP  = [ip of vm 2]
BENCH_IP  = [ip of vm 3]
```
#### 启动keentuned
```shell
$ keentuned
```
### 动态参数调优
keentune param tune --param [param conf] --bench [bench conf] --name [job name]   --iteration [number of iteration]  
运行示例:  
```bash
$ keentune param tune --param param_100.json --bench benchmark/wrk/bench_wrk_nginx_long.json --name tune_test --iteration 10 
```
任务发起后，可以通过msg命令检查任务执行情况。  
```bash
$ keentune msg
```
### 静态参数调优
keentune profile set [profile name]  
运行示例:  
```bash
$ keentune profile set cpu_high_load.conf
```

## 代码结构
---  
+ cli: KeenTune 控制台模块
+ daemon: 通用方法模块
    + api: API接口定义
    + common: 通用方法模块
    + example: 示例param conf和benchmark conf
    + modules: 功能实现模块

## Documentation
---
其他命令使用详见 keentune help信息或[KeenTune用户指南]。
