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
from common import getTuneTaskResult
from common import getTaskLogPath
from common import sysCommand


logger = logging.getLogger(__name__)


class TestMultiTarget(unittest.TestCase):
    def setUp(self) -> None:
        self.target, self.bench, self.brain = self.get_server_ip()
        self.port = "TARGET_PORT = 9873"
        self.scene_1 = "PARAMETER = sysctl_target.json"
        self.scene_2 = "PARAMETER = nginx.json, sysctl_target.json, nginx.json"
        self.scene_3 = "PARAMETER = sysctl_target.json, nginx.json, sysctl_target.json"
        self.code_path = "/etc/keentune"

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
        self.delete_tmp_file()
        logger.info('the test_multi_target testcase finished')

    def get_server_ip(self):
        with open("common.py", "r", encoding='UTF-8') as f:
            data = f.read()
        target = re.search(r"target_ip=\"(.*)\"", data).group(1)
        bench = re.search(r"bench_ip=\"(.*)\"", data).group(1)
        brain = re.search(r"brain_ip=\"(.*)\"", data).group(1)

        return target, bench, brain

    def delete_tmp_file(self):
        param_path = "{}/parameter/sysctl_target.json".format(self.code_path)
        script_path = "{}/benchmark/wrk/nginx_http_long_multi_target.py".format(self.code_path)
        bench_path = "{}/benchmark/wrk/bench_wrk_nginx_long_multi_target.json".format(self.code_path)
        file_list = [param_path, script_path, bench_path]

        for path in file_list:
            if os.path.exists(path):
                os.remove(path)
        
    def run_param_tune(self):
        cmd = 'keentune param tune -i 1 --job test1'
        path = getTaskLogPath(cmd)
        result = getTuneTaskResult(path)
        self.assertTrue(result)

    def check_sysctl_params(self, server):
        path = "/var/keentune/parameter/test1/test1_group{}_best.json".format(str(server[1]))
        with open(path, "r", encoding='UTF-8') as f:
            params = json.load(f)

        for data in params["parameters"]:
            if data["domain"] == "sysctl":
                param_name = data["name"]
                param_value = str(data["value"])
                cmd = "ssh {} 'sysctl -n {}'".format(server[0], param_name)
                sys_value = str(sysCommand(cmd)[1].strip('\n')).replace("\t", " ")
                if param_value != sys_value:
                    print("param_name is: %s" % param_name)
                    print("param_value is: %s" % param_value)
                    print("sys_value is: %s" % sys_value)
                    self.status = 1
                    break   
        else:
            self.status = 0

        self.assertEqual(self.status, 0)

    def run_multi_target(self, scene_cmd, bench_ip, server_list):
        if self.target == "localhost":
            logger.info("this is standalone mode, don't need to run this use case")
        else:
            cmd = 'sh conf/restart_keentuned.sh {} "{}" "{}"'.format(self.target, bench_ip, scene_cmd)
            self.status, self.out, _ = sysCommand(cmd)
            self.assertEqual(self.status, 0)
            self.assertIn("restart keentuned server successfully!", self.out)

            self.run_param_tune()
            for server in server_list:
                self.check_sysctl_params(server)
            return True
    
    def test_multi_target_01(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}, {}\n{}\n{}".format("localhost", self.target, self.port, self.scene_2)
        self.run_multi_target(scene_cmd, "localhost", [(self.target, 1)])

    def test_multi_target_02(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}, {}, {}, {}\n{}\n{}".format("localhost", self.target, self.bench, self.brain, self.port, self.scene_2)
        server_list = [(self.target, 1), (self.bench, 1), (self.brain, 1)]
        self.run_multi_target(scene_cmd, "localhost", server_list)

    def test_multi_target_03(self):
        group1_cmd = r"[target-group-1]\nTARGET_IP = {}\n{}\n{}".format("localhost", self.port, self.scene_2)
        group2_cmd = r"[target-group-2]\nTARGET_IP = {}, {}, {}\n{}\n{}".format(self.target, self.bench, self.brain, self.port, self.scene_3)
        scene_cmd = r"\n{}\n{}".format(group1_cmd, group2_cmd)
        server_list = [(self.target, 2), (self.bench, 2), (self.brain, 2)]
        self.run_multi_target(scene_cmd, "localhost", server_list)
    
    def test_multi_target_04(self):
        group1_cmd = r"[target-group-1]\nTARGET_IP = {}, {}\n{}\n{}".format("localhost", self.target, self.port, self.scene_2)
        group2_cmd = r"[target-group-2]\nTARGET_IP = {}, {}\n{}\n{}".format(self.bench, self.brain, self.port, self.scene_3)
        scene_cmd = r"\n{}\n{}".format(group1_cmd, group2_cmd)
        server_list = [(self.target, 1), (self.bench, 2), (self.brain, 2)]
        self.run_multi_target(scene_cmd, "localhost", server_list)

    def test_multi_target_05(self):
        group1_cmd = r"[target-group-1]\nTARGET_IP = {}\n{}\n{}".format("localhost", self.port, self.scene_1)
        group2_cmd = r"[target-group-2]\nTARGET_IP = {}\n{}\n{}".format(self.target, self.port, self.scene_2)
        group3_cmd = r"[target-group-3]\nTARGET_IP = {}\n{}\n{}".format(self.bench, self.port, self.scene_3)
        scene_cmd = r"\n{}\n{}\n{}".format(group1_cmd, group2_cmd, group3_cmd)
        server_list = [(self.target, 2), (self.bench, 3)]
        self.run_multi_target(scene_cmd, "localhost", server_list)

    def test_multi_target_06(self):
        group1_cmd = r"[target-group-1]\nTARGET_IP = {}\n{}\n{}".format("localhost", self.port, self.scene_1)
        group2_cmd = r"[target-group-2]\nTARGET_IP = {}\n{}\n{}".format(self.target, self.port, self.scene_2)
        group3_cmd = r"[target-group-3]\nTARGET_IP = {}, {}\n{}\n{}".format(self.bench, self.brain, self.port, self.scene_3)
        scene_cmd = r"\n{}\n{}\n{}".format(group1_cmd, group2_cmd, group3_cmd)
        server_list = [(self.target, 2), (self.bench, 3), (self.brain, 3)]
        self.run_multi_target(scene_cmd, "localhost", server_list)

    def test_multi_target_07(self):
        group1_cmd = r"[target-group-1]\nTARGET_IP = {}\n{}\n{}".format("localhost", self.port, self.scene_1)
        group2_cmd = r"[target-group-2]\nTARGET_IP = {}\n{}\n{}".format(self.target, self.port, self.scene_2)
        group3_cmd = r"[target-group-3]\nTARGET_IP = {}\n{}\n{}".format(self.bench, self.port, self.scene_3)
        group4_cmd = r"[target-group-4]\nTARGET_IP = {}\n{}\n{}".format(self.brain, self.port, self.scene_1)
        scene_cmd = r"\n{}\n{}\n{}\n{}".format(group1_cmd, group2_cmd, group3_cmd, group4_cmd)
        server_list = [(self.target, 2), (self.bench, 3), (self.brain, 4)]
        status = self.run_multi_target(scene_cmd, "localhost", server_list)

        if status:
            cmd = "echo y | keentune param dump --job test1"
            self.status, self.out, _ = sysCommand(cmd)
            self.assertEqual(self.status, 0)
            cmd = "keentune profile set --group1 test1_group1.conf --group2 test1_group2.conf --group3 test1_group3.conf --group4 test1_group4.conf"
            self.status, self.out, _ = sysCommand(cmd)
            self.assertEqual(self.status, 0)
            self.assertIn("Set test1_group1.conf successfully", self.out)
            self.assertIn("Set test1_group2.conf successfully", self.out)
            self.assertIn("Set test1_group3.conf successfully", self.out)
            self.assertIn("Set test1_group4.conf successfully", self.out)

    def test_multi_bench_01(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}\n{}\n{}".format(self.target, self.port, self.scene_3)
        bench_ip = "localhost, {}".format(self.bench)
        self.run_multi_target(scene_cmd, bench_ip, [(self.target, 1)])

    def test_multi_bench_02(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}\n{}\n{}".format(self.target, self.port, self.scene_3)
        bench_ip = "localhost, {}, {}, {}".format(self.target, self.bench, self.brain)
        self.run_multi_target(scene_cmd, bench_ip, [(self.target, 1)])

    def test_multi_target_bench_01(self):
        group1_cmd = r"[target-group-1]\nTARGET_IP = {}, {}\n{}\n{}".format("localhost", self.target, self.port, self.scene_2)
        group2_cmd = r"[target-group-2]\nTARGET_IP = {}, {}\n{}\n{}".format(self.bench, self.brain, self.port, self.scene_3)
        scene_cmd = r"\n{}\n{}".format(group1_cmd, group2_cmd)
        bench_ip = "localhost, {}".format(self.bench)
        server_list = [(self.target, 1), (self.bench, 2), (self.brain, 2)]
        self.run_multi_target(scene_cmd, bench_ip, server_list)

    def test_multi_target_bench_02(self):
        group1_cmd = r"[target-group-1]\nTARGET_IP = {}\n{}\n{}".format("localhost", self.port, self.scene_1)
        group2_cmd = r"[target-group-2]\nTARGET_IP = {}\n{}\n{}".format(self.target, self.port, self.scene_2)
        group3_cmd = r"[target-group-3]\nTARGET_IP = {}\n{}\n{}".format(self.bench, self.port, self.scene_3)
        group4_cmd = r"[target-group-4]\nTARGET_IP = {}\n{}\n{}".format(self.brain, self.port, self.scene_1)
        scene_cmd = r"\n{}\n{}\n{}\n{}".format(group1_cmd, group2_cmd, group3_cmd, group4_cmd)
        bench_ip = "localhost, {}, {}, {}".format(self.target, self.bench, self.brain)
        server_list = [(self.target, 2), (self.bench, 3), (self.brain, 4)]
        self.run_multi_target(scene_cmd, bench_ip, server_list)

