#!/bin/bash
keentune_code_dir=/tmp/keentune-cluster-tmp
target_server=$1
bench_server=$2

keentuned_path=$keentune_code_dir/acops-new/keentuned
install_conf=$keentuned_path/keentuned_install.sh
benchmark_script=$keentune_code_dir/acops-new/test/conf/nginx_http_long_multi_target.py
json_path=$keentune_code_dir/acops-new/test/conf/bench_wrk_nginx_long_multi_target.json
sysctl_path=$keentune_code_dir/acops-new/test/conf/sysctl_target.json

clear_keentune_env()
{
    ps -ef|grep -w 'keentuned'|grep -v grep|awk '{print $2}'| xargs -I {} kill -9 {}
}

restart_keentuned()
{
    sed -i "s/target_ip=.*/target_ip=\"${target_server},${bench_server}\"/g" $install_conf
    sed -i "s/multi_target_ip=.*/multi_target_ip=\"${bench_server}\"/g" $benchmark_script

    cp $sysctl_path $keentuned_path/daemon/examples/parameter
    cp $json_path $keentuned_path/daemon/examples/benchmark/wrk/
    cp $benchmark_script $keentuned_path/daemon/examples/benchmark/wrk/
    
    cd $keentuned_path
    ./keentuned_install.sh || ret=1
    keentuned > keentuned-multi_target.log 2>&1 &
}

clear_keentune_env
restart_keentuned
echo "restart keentuned server successfully!"
