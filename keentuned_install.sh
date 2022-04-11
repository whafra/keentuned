#!/bin/bash
# keentuned configuration and installation script

#############################################手动配置部分开始##########################################
# 1. 配置文件修改部分，可以按需修改配置参数，也可以执行完脚本后，手动按需修改 /etc/keentune/conf/keentuned.conf。
# code_home 代码仓的位置, 写到daemon这一级，该值不需要修改。
code_home=`pwd`
# keentuned_home keentune依赖的配置文件位置
keentuned_home="/etc/keentune"
# output_dir keentune日志、调优结果等输出文件的位置
output_dir="/var/keentune"
# 执行的目标benchmark信息json文件，目录为 $keentuned_home/benchmark/ ，可以在其下创建自己的目录和文件，保证写入的目录和文件存在即可
bench_json_file="benchmark/wrk/bench_wrk_nginx_long.json"

# 执行的目标benchmark对应的Python3 脚本，目录为 $keentuned_home/benchmark/ ，可以在其下创建自己的目录和文件，保证写入的目录和文件存在即可
local_script_path="benchmark/wrk/ack_nginx_http_long_base.py"

# 注意：以下部分是修改配置文件的内容，

# 指定param tune调优的AI 算法
algorithm="tpe"
# 指定sensitize collect 敏感参数采集的算法
sensitive_algorithm="random"
# 指定调优的算法的运行的机器ip
brain_ip="localhost"
# 调优目标机器的ip
target_ip="localhost"
# benchmark打压机器的ip
bench_ip="localhost"

# 调优开始前的基线benchmark打压执行次数
base_round=2
# 每轮调优的benchmark打压执行次数
tune_round=1
# 最优配置的benchmark打压执行次数
check_round=1
# sensitize collect 敏感参数采集的benchmark打压执行次数
sensi_round=2

bench_conf="bench_wrk_nginx_long.json"
param_conf="sysctl.json"
bench_dest="localhost"

#############################################手动配置部分结束############################################

Check_Exec() {
    if [ $? -ne 0 ]; then 
    echo -e "exec [$1] failed";
    exit 1
    fi
}

# 2. Configuration parameter
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


# modify keentuned.conf
sed -i "s#KEENTUNED_HOME = .*#KEENTUNED_HOME = ${keentuned_home}#" $keentuned_home/conf/keentuned.conf
sed -i "s/BRAIN_IP = .*/BRAIN_IP = ${brain_ip}/" $keentuned_home/conf/keentuned.conf
sed -i "s/BENCH_SRC_IP = .*/BENCH_SRC_IP = ${bench_ip}/" $keentuned_home/conf/keentuned.conf
sed -i "17s/ALGORITHM = .*/ALGORITHM = ${algorithm}/" $keentuned_home/conf/keentuned.conf
sed -i "60s/ALGORITHM = .*/ALGORITHM = ${sensitive_algorithm}/" $keentuned_home/conf/keentuned.conf
sed -i "0,/TARGET_IP = .*/s//TARGET_IP = ${target_ip}/" $keentuned_home/conf/keentuned.conf

sed -i "s/BASELINE_BENCH_ROUND = .*/BASELINE_BENCH_ROUND = ${base_round}/" $keentuned_home/conf/keentuned.conf
sed -i "s/TUNING_BENCH_ROUND = .*/TUNING_BENCH_ROUND = ${tune_round}/" $keentuned_home/conf/keentuned.conf
sed -i "s/RECHECK_BENCH_ROUND = .*/RECHECK_BENCH_ROUND = ${check_round}/" $keentuned_home/conf/keentuned.conf
sed -i "0,/BENCH_DEST_IP = .*/s//BENCH_DEST_IP = ${bench_dest}/" $keentuned_home/conf/keentuned.conf
sed -i "s/^BENCH_ROUND = .*/BENCH_ROUND = ${sensi_round}/" $keentuned_home/conf/keentuned.conf

sed -i "0,/PARAMETER = .*/s//PARAMETER = ${param_conf}/" $keentuned_home/conf/keentuned.conf
sed -i "s/BENCH_CONFIG = .*/BENCH_CONFIG = ${bench_conf}/" $keentuned_home/conf/keentuned.conf

# modify bench_json_file
sed -i "s%\"local_script_path\":[^,]*%\"local_script_path\": \"${local_script_path}\"%" $keentuned_home/$bench_json_file

#3. Generate compiled files, and take effect global commands
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
cd $code_home/daemon
go install .
Check_Exec "install keentune service"

go install ../cli
Check_Exec "install keentune cli"

go_path=`go env | grep -i "GOPATH" | awk -F'"' '{print$2}'`

# -f ${go_path}/bin/cmd :Check the file whether exists; -f :represents a file; -d: represents a directory.
if [ -f ${go_path}/bin/daemon ]; then
    mv ${go_path}/bin/daemon -f /usr/bin/keentuned
    echo "keentuned command generate successfully"
else
    echo "daemon not exist"
    exit 1
fi

if [ -f ${go_path}/bin/cli ]; then
    mv ${go_path}/bin/cli -f /usr/bin/keentune
    echo "keentune command generate successfully"
else
    echo "cli not exist"
    exit 1
fi
