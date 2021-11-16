import os
import re
import requests
import subprocess
import time
import unittest
import logging

from keentune_config import keentuned_ip, keentuned_port, brain_ip, \
    brain_port, bench_ip, bench_port,  target_ip, target_port

logger = logging.getLogger(__name__)


def sysCommand(cmd):
    result = subprocess.run(
        cmd,
        shell=True,
        close_fds=True,
        stderr=subprocess.PIPE,
        stdout=subprocess.PIPE
    )

    status = result.returncode
    output = result.stdout.decode('UTF-8', 'strict')
    error = result.stderr.decode('UTF-8', 'strict')

    return status, output, error


def getServerStatus(server):
    data_dict = {
        "keentuned": [keentuned_ip, keentuned_port],
        "keentune-brain": [brain_ip, brain_port],
        "keentune-target": [target_ip, target_port],
        "keentune-bench": [bench_ip, bench_port]
    }
    event = "sensitize_list" if server == "keentune-brain" else "status"
    url = "http://{}:{}/{}".format(data_dict[server][0], data_dict[server][1], event)
    res = requests.get(url, proxies={"http": None, "https": None})
    if res.status_code != 200:
        logger.error("Please check {} server is running...".format(server))
        result = 1
    else:
        result = 0
    return result


def checkServerStatus(server_list):
    result = False
    for server in server_list:
        status = getServerStatus(server)
        result = result or status
    return result


def deleteDependentData(param_name):
    cmd = "echo y | keentune param delete --job {}".format(param_name)
    sysCommand(cmd)
    cmd = 'echo y | keentune sensitize delete --data {}'.format(param_name)
    sysCommand(cmd)


def runParamTune():
    cmd = 'keentune param tune --param parameter/sysctl.json -i 1 --bench benchmark/wrk/bench_wrk_nginx_long.json --job test1'
    _, output, _ = sysCommand(cmd)
    path = re.search(r'\s+"(.*?)"', output).group(1)
    time.sleep(3)
    while True:
        with open(path, 'r') as f:
            res_data = f.read()
        if '[BEST] Tuning improvment' in res_data:
            break
        time.sleep(8)

    word_list = ["Step1", "Step2", "Step3", "Step4",
                 "Step5", "Step6", "[BEST] Tuning improvment"]
    res = all([word in res_data for word in word_list])
    result = 0 if res else 1
    return result


def runParamDump():
    cmd = 'echo y | keentune param dump -j test1 -o test1.conf'
    sysCommand(cmd)
    path = "/var/keentune/profile/test1.conf"
    res = os.path.exists(path)
    result = 0 if res else 1
    return result


def runProfileSet():
    cmd = 'keentune profile set --name test1.conf'
    sysCommand(cmd)
    cmd = 'keentune profile list'
    _, output, _ = sysCommand(cmd)
    res = re.search(r'\[(.*?)\].+test1.conf', output).group(1)
    result = 0 if res.__contains__('active') else 1
    return result


def runSensitizeCollect():
    cmd = 'keentune sensitize collect -i 10 --param parameter/sysctl.json --bench benchmark/wrk/bench_wrk_nginx_long.json --data test2'
    _, output, _ = sysCommand(cmd)
    path = re.search(r'\s+"(.*?)"', output).group(1)
    time.sleep(3)
    while True:
        with open(path, 'r') as f:
            res_data = f.read()
        if 'Sensitization collection finished' in res_data:
            break
        time.sleep(8)

    word_list = ["Step1", "Step2", "Step3", "Step4",
                 "Sensitization collection finished"]
    res = all([word in res_data for word in word_list])
    result = 0 if res else 1
    return result
