# Dependencies of KeenTune
---  

## 添加 yum 源
1. 将以下keentune源追加进系统的yum源
```s
vim /etc/yum.repos.d/epel.repo

[keentune]
name=keentune-os
baseurl=https://mirrors.openanolis.cn/anolis/8.4/Plus/x86_64/os/
gpgkey=https://mirrors.openanolis.cn/anolis/RPM-GPG-KEY-ANOLIS
enabled=1
```
2. 重新生成缓存
```s
yum clean all
yum makecache
```

## keentuned 编译安装依赖项
```s
#go编译器安装
yum install go

#验证go编译器
go env
go version

#配置go环境
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
```

## keentune-brain 运行依赖项
```s
yum install python36 python36-devel rust openssl-devel

pip3 install --upgrade pip
pip3 install tornado==6.1
pip3 install pynginxconfig
pip3 install numpy==1.19.5
pip3 install POAP==0.1.26
pip3 install hyperopt==0.2.5
pip3 install ultraopt==0.1.1
pip3 install requests==2.25.1 
pip3 install pySOT==0.3.3 
pip3 install scikit_learn==0.24.2 
pip3 install shap 
pip3 install xgboost
```
如果安装shap和xgboost时报告gcc相关错误，使用以下步骤解决  
```s
yum install gcc gcc-c++ make cmake
pip3 install xgboost
pip3 install shap==0.35.0
```

## keentune-target 运行依赖项
```s
pip3 install tornado==6.1
pip3 install pynginxconfig
```

## keentune-bench 运行依赖项
```s
pip3 install tornado==6.1
```