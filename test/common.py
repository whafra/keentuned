import os
import re
import subprocess
import time
import unittest
import logging

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


def checkServerStatus(server_list):
    result = False
    for server in server_list:
        cmd = "ps aux | grep {} |grep -v grep".format(server)
        status, _, _ = sysCommand(cmd)
        if status:
            logger.error("Please check {} server is running...".format(server))
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
