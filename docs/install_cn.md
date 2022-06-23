# Keentune安装手册
## YUM源配置
编辑系统的yum源配置
```sh
vi /etc/yum.repos.d/epel.repo
```
添加以下内容
```sh
[keentune]
name=keentune-os
baseurl=https://mirrors.openanolis.cn/anolis/8.6/Plus/x86_64/os/
gpgkey=https://mirrors.openanolis.cn/anolis/RPM-GPG-KEY-ANOLIS
enabled=1
```
重新生成yum缓存
```sh
yum clean all
yum makecache
```
---  
## Keentuned
### 1. 通过YUM安装
```sh
yum install keentuned
```
启动keentuned
```sh
systemctl start keentuned
```
### 2. 通过源码安装
准备golang编译环境
```sh
yum install go
go env
go version
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
```
下载代码
```sh
git clone https://gitee.com/anolis/keentuned.git
```
编译安装
```sh
cd keentuned
./keentuned_install.sh
```
启动keentuned
```
keentuned
```
---  
## Keentune-Target
安装依赖
```sh
yum install python36 python36-devel rust openssl-devel
pip3 install --upgrade pip
pip3 install tornado==6.1 pynginxconfig
```
### 1. 通过YUM安装
安装keentune-target
```sh
yum install keentune-target
```
启动keentune-target
```sh
systemctl start keentune-target
```
### 2. 通过源码安装
下载Keentune-target源码
```sh
git clone https://gitee.com/anolis/keentune_target.git
```
运行安装脚本
```sh
cd keentune_target
sudo python3 setup.py install
```
启动keentune-target
```sh
keentune-target
```
---  
## Keentune-Brain
安装依赖
```sh
yum install python36 python36-devel rust openssl-devel
pip3 install --upgrade pip
pip3 install numpy==1.19.5 POAP==0.1.26 tornado==6.1 hyperopt==0.2.5 ultraopt==0.1.1 bokeh==2.3.2 requests==2.25.1 pySOT==0.3.3 scikit_learn==0.24.2 paramiko==2.7.2 PyYAML==5.4.1 shap xgboost
```
### 1. 通过YUM安装
```sh
yum install keentune-brain
```
启动keentune-brain
```sh
systemctl start keentune-brain
```
### 2. 通过源码安装
下载Keentune-Brain源码
```sh
git clone https://gitee.com/anolis/keentune_brain.git
```
运行安装脚本
```sh
cd keentune_brain
sudo python3 setup.py install
```
启动keentune-brain
```sh
keentune-brain
```
---  
## Keentune-Bench
安装依赖
```sh
yum install python36 python36-devel rust openssl-devel
pip3 install --upgrade pip
pip3 install tornado==6.1 pynginxconfig
```
### 1. 通过YUM安装
```sh
yum install keentune-bench
```
启动keentune-bench
```sh
systemctl start keentune-bench
```
### 2. 通过源码安装
下载Keentune-Bench源码
```sh
git clone https://gitee.com/anolis/keentune_bench.git
```
运行安装脚本
```sh
cd keentune_bench
sudo python3 setup.py install
```
启动keentune-bench
```sh
keentune-bench
```