import os
import re
import sys
import json
import logging
import time
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import deleteDependentData
from common import checkServerStatus
from common import sysCommand

logger = logging.getLogger(__name__)


class TestMultiTarget(unittest.TestCase):
    def setUp(self) -> None:
        self.code_path = "/tmp/keentune-cluster-tmp/acops-new/keentuned/daemon/examples"
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_multi_target testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("test1")
        logger.info('the test_multi_target testcase finished')

    def run_param_tune(self):
        cmd = 'keentune param tune --param parameter/sysctl_target.json -i 1 --bench benchmark/wrk/bench_wrk_nginx_long_multi_target.json --job test1'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)

        path = re.search(r'\s+"(.*?)"', self.out).group(1)
        time.sleep(3)
        while True:
            with open(path, 'r') as f:
                res_data = f.read()
            if '[BEST] Tuning improvement' in res_data:
                break
            time.sleep(8)

        word_list = ["Step1", "Step2", "Step3", "Step4",
                    "Step5", "Step6", "[BEST] Tuning improvement"]
        result = all([word in res_data for word in word_list])
        self.assertTrue(result)

    def check_sysctl_params(self, server):
        path = "/var/keentune/parameter/test1/test1_best.json"
        with open(path, "r", encoding='UTF-8') as f:
            params = json.load(f)

        for data in params["parameters"]:
            param_name = data["name"]
            param_value = str(data["value"])
            cmd = "ssh {} 'sysctl -n {}'".format(server, param_name)
            sys_value = str(sysCommand(cmd)[1].strip('\n')).replace("\t", " ")
            if param_value != sys_value:
                self.status = 1
                break   
        else:
            self.status = 0

        self.assertEqual(self.status, 0)
        
    def test_multi_target(self):
        with open("common.py", "r", encoding='UTF-8') as f:
            data = f.read()
        target_ip = re.search(r"target_ip=\"(.*)\"", data).group(1)
        bench_ip = re.search(r"bench_ip=\"(.*)\"", data).group(1)

        if target_ip == "localhost":
            logger.info("this is standalone mode, don't need to run this use case")
        else:
            cmd = 'sh conf/restart_keentuned.sh {} {}'.format(target_ip, bench_ip)
            self.status, self.out, _  = sysCommand(cmd)
            self.assertEqual(self.status, 0)
            self.assertTrue(self.out.__contains__('restart keentuned server successfully!'))

            self.run_param_tune()
            self.check_sysctl_params(target_ip)
            self.check_sysctl_params(bench_ip)

            os.remove("{}/parameter/sysctl_target.json".format(self.code_path))
            os.remove("{}/benchmark/wrk/nginx_http_long_multi_target.py".format(self.code_path))
            os.remove("{}/benchmark/wrk/bench_wrk_nginx_long_multi_target.json".format(self.code_path))

