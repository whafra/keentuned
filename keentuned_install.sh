#!/bin/bash
# keentuned configuration and installation script

# code_home 代码仓的位置, 写到daemon这一级
code_home=`pwd`
# keentuned_home keentune依赖的配置文件位置
keentuned_home="/etc/keentune"


Check_Exec() {
    if [ $? -ne 0 ]; then 
    echo -e "exec [$1] failed";
    exit 1
    fi
}

# 1. Copy keentuned Configuration
if [ ! -d "$code_home" ]; then
    echo "code_home must be specified"
    exit 1
else
    echo "code_home is ${code_home}, begin to Compile keentune"
fi

if [ ! -d /etc/keentune/conf ]; then
    mkdir -p /etc/keentune/conf
    Check_Exec "mkdir /etc/keentune/conf"
fi

cp $code_home/keentuned.conf -f $keentuned_home/conf
Check_Exec "cp keentuned.conf"


/bin/cp -rf $code_home/daemon/examples/. $keentuned_home
if [ $? -ne 0 ]; then
    cp -rf $code_home/daemon/examples/. $keentuned_home
    Check_Exec "cp -rf examples to keentune home"
fi


#2. set go env
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct

echo "Step1. Complete environment configuration"

#3. go install
cd $code_home/daemon
go install .
Check_Exec "install keentune service"

go install ../cli
Check_Exec "install keentune cli"

go_path=`go env | grep -i "GOPATH" | awk -F'"' '{print$2}'`


if [ -f ${go_path}/bin/daemon ]; then
    mv ${go_path}/bin/daemon -f /usr/bin/keentuned
    echo "Step2. Build 'keentuned' daemon successfully"
else
    echo "daemon not exist"
    exit 1
fi

if [ -f ${go_path}/bin/cli ]; then
    mv ${go_path}/bin/cli -f /usr/bin/keentune
    echo "Step3. Build 'keentune' command successfully"
else
    echo "cli not exist"
    exit 1
fi

echo "Step4. Welcome to use!"
