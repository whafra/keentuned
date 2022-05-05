#!/bin/bash
algorithm=$1

brain_conf=/etc/keentune/conf/brain.conf

clear_keentune_env()
{
    ps -ef|grep -w 'keentune-brain'|grep -v grep|awk '{print $2}'| xargs -I {} kill -9 {}
}

restart_brain()
{
    sed -i "s/explainer =.*/explainer = ${algorithm}/g" $brain_conf
    keentune-brain > keentune_brain_algorithm.log 2>&1 &
    sleep 5
}

clear_keentune_env
restart_brain
echo "restart brain server successfully!"