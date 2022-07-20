#!/bin/bash
algorithm=$1

keentuned_conf=/etc/keentune/conf/keentuned.conf

clear_keentune_env()
{
    ps -ef|grep -w 'keentuned'|grep -v grep|awk '{print $2}'| xargs -I {} kill -9 {}
}

restart_brain()
{
    sed -i "s/SENSITIZE_ALGORITHM.*=.*/SENSITIZE_ALGORITHM = ${algorithm}/g" $keentuned_conf
    keentuned > keentune_train_algorithm.log 2>&1 &
    sleep 5
}

clear_keentune_env
restart_brain
echo "restart brain server successfully!"