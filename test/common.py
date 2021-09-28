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
    cmd = "echo y | keentune param delete --name {}".format(param_name)
    sysCommand(cmd)
    cmd = 'echo y | keentune sensitize delete --name {}'.format(param_name)
    sysCommand(cmd)


def runParamTune():
    cmd = 'keentune param tune --param_conf parameter/param_100.json -i 1 --bench_conf benchmark/wrk/bench_wrk_nginx_long.json --name test1'
    sysCommand(cmd)
    while True:
        cmd = "keentune msg --name 'param tune'"
        _, output, _ = sysCommand(cmd)
        if '[BEST] Tuning improvment' in output:
            break
        time.sleep(8)

    word_list = ["Step1", "Step2", "Step3", "Step4",
                 "Step5", "Step6", "[BEST] Tuning improvment"]
    res = all([word in output for word in word_list])
    result = 0 if res else 1
    return result


def runParamDump():
    cmd = 'echo y | keentune param dump -n test1 -o test1.conf'
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
    cmd = 'keentune sensitize collect -i 1 --param_conf parameter/param_100.json --bench_conf benchmark/wrk/bench_wrk_nginx_long.json --name test2'
    sysCommand(cmd)

    while True:
        cmd = "keentune msg --name 'sensitize collect'"
        _, output, _ = sysCommand(cmd)
        if 'Sensitization collection finished' in output:
            break
        time.sleep(8)

    word_list = ["Step1", "Step2", "Step3", "Step4",
                 "Sensitization collection finished"]
    res = all([word in output for word in word_list])
    result = 0 if res else 1
    return result
