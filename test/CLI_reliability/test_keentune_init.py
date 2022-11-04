import os
import re
import sys
import json
import logging
import time
import yaml
import unittest
from collections import defaultdict

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import deleteDependentData
from common import checkServerStatus
from common import getTuneTaskResult
from common import getTaskLogPath
from common import sysCommand


logger = logging.getLogger(__name__)


class TestKeentuneInit(unittest.TestCase):
    @classmethod
    def setUpClass(self) -> None:
        self.port = "TARGET_PORT = 9873"
        self.scene_1 = "PARAMETER = sysctl.json"
        self.scene_2 = "PARAMETER = nginx_conf.json"
        self.scene_3 = "PARAMETER = sysctl.json, nginx_conf.json"
        self.path = "/etc/keentune/conf/init.yaml"
        cmd = "echo y | cp /etc/keentune/conf/keentuned.conf /etc/keentune/conf/keentuned_bak.conf"
        status, _, _ = sysCommand(cmd)
        assert status == 0
        logger.info("TestKeentuneInit begin...")

    @classmethod
    def tearDownClass(self) -> None:
        cmd = "echo y | mv /etc/keentune/conf/keentuned_bak.conf /etc/keentune/conf/keentuned.conf"
        status, _, _ = sysCommand(cmd)
        assert status == 0
        logger.info("TestKeentuneInit end...")

    def setUp(self) -> None:
        self.target, self.bench, self.brain = self.get_server_ip()
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_keentune_init testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('the test_keentune_init testcase finished')

    def get_server_ip(self):
        with open("common.py", "r", encoding='UTF-8') as f:
            data = f.read()
        target = re.search(r"target_ip=\"(.*)\"", data).group(1)
        bench = re.search(r"bench_ip=\"(.*)\"", data).group(1)
        brain = re.search(r"brain_ip=\"(.*)\"", data).group(1)
        return target, bench, brain

    def check_yaml_file(self, data_dict):
        with open(self.path) as f:
            try:
                yaml_dict = yaml.safe_load(f)
            except yaml.YAMLError as exc:
                print("ingress nginx load yaml file failed, error is: {}".format(exc))

        yaml_dict.pop("brain")
        res_dict = defaultdict(dict)
        for group in yaml_dict:
            for data in yaml_dict[group]:
                if "target" in group:
                    res_dict[group][data["ip"]] = {"available": data["available"], "knobs": data["knobs"]}
                else:
                    res_dict[group][data["ip"]] = {"available": data["available"], "destination": data["destination"]}
        self.assertEqual(res_dict, data_dict)

    def run_multi_scense_init(self, scene_cmd, dest_ip, bench_ip, data_dict, msg=None):
        msg = msg if msg else "KeenTune Init success"
        if self.target == "localhost":
            logger.info("this is standalone mode, don't need to run this use case")
        else:
            cmd = 'sh conf/init_keentune.sh {} "{}" "{}"'.format(dest_ip, bench_ip, scene_cmd)
            self.status, self.out, _ = sysCommand(cmd)
            self.assertEqual(self.status, 0)
            self.assertIn("init keentune successfully!", self.out)

            cmd = 'keentune init'
            self.status, self.out, _ = sysCommand(cmd)
            self.assertEqual(self.status, 0)
            self.assertIn(msg, self.out)

            self.check_yaml_file(data_dict)
    
    def test_init_FUN_base(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}\n{}\n{}".format("localhost", self.port, self.scene_1)
        data_dict = {"target-group-1": {"localhost": {"available": True, "knobs": ["sysctl.json"]}},
                     "bench-group-1": {"localhost": {"available": True, "destination": {"ip": "localhost", "reachable": True}}}}
        self.run_multi_scense_init(scene_cmd, "localhost", "localhost", data_dict)

    def test_init_RBT_multi_target_01(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}, {}\n{}\n{}".format("localhost", self.target, self.port, self.scene_1)
        data_dict = {"target-group-1": {"localhost": {"available": True, "knobs": ["sysctl.json"]},
                                        self.target: {"available": True, "knobs": ["sysctl.json"]}},
                     "bench-group-1": {"localhost": {"available": True, "destination": {"ip": "localhost", "reachable": True}}}}
        self.run_multi_scense_init(scene_cmd, "localhost", "localhost", data_dict)

    def test_init_RBT_multi_target_02(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}, {}\n{}\n{}".format("localhost", self.target, self.port, self.scene_2)
        data_dict = {"target-group-1": {"localhost": {"available": True, "knobs": ["nginx_conf.json"]},
                                        self.target: {"available": True, "knobs": ["nginx_conf.json"]}},
                     "bench-group-1": {"localhost": {"available": True, "destination": {"ip": "localhost", "reachable": True}}}}
        self.run_multi_scense_init(scene_cmd, "localhost", "localhost", data_dict)

    def test_init_RBT_multi_target_03(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}, {}\n{}\n{}".format("localhost", self.target, self.port, self.scene_3)
        data_dict = {"target-group-1": {"localhost": {"available": True, "knobs": ["sysctl.json", "nginx_conf.json"]},
                                        self.target: {"available": True, "knobs": ["sysctl.json", "nginx_conf.json"]}},
                     "bench-group-1": {"localhost": {"available": True, "destination": {"ip": "localhost", "reachable": True}}}}
        self.run_multi_scense_init(scene_cmd, "localhost", "localhost", data_dict)
    
    def test_init_RBT_multi_target_04(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}, {}\n{}\n{}".format("localhost", self.target, self.port, self.scene_1)
        data_dict = {"target-group-1": {"localhost": {"available": True, "knobs": ["sysctl.json"]},
                                        self.target: {"available": True, "knobs": ["sysctl.json"]}},
                     "bench-group-1": {self.bench: {"available": True, "destination": {"ip": self.target, "reachable": True}}}}
        self.run_multi_scense_init(scene_cmd, self.target, self.bench, data_dict)

    def test_init_RBT_multi_target_05(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}, {}\n{}\n{}".format(self.target, self.bench, self.port, self.scene_1)
        data_dict = {"target-group-1": {self.target: {"available": True, "knobs": ["sysctl.json"]},
                                        self.bench: {"available": True, "knobs": ["sysctl.json"]}},
                     "bench-group-1": {"localhost": {"available": True, "destination": {"ip": "localhost", "reachable": True}}}}
        self.run_multi_scense_init(scene_cmd, "localhost", "localhost", data_dict)

    def test_init_RBT_multi_target_06(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}, {}\n{}\n{}".format(self.target, "1.2.3.4", self.port, self.scene_1)
        data_dict = {"target-group-1": {self.target: {"available": True, "knobs": ["sysctl.json"]},
                                        "1.2.3.4": {"available": False, "knobs": ["sysctl.json"]}},
                     "bench-group-1": {"localhost": {"available": True, "destination": {"ip": "localhost", "reachable": True}}}}
        msg = "target-group[1]: 1.2.3.4 offline"
        self.run_multi_scense_init(scene_cmd, "localhost", "localhost", data_dict, msg)

    def test_init_RBT_multi_target_07(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}, {}\n{}\n{}".format(self.target, "1.2.3.4", self.port, self.scene_1)
        data_dict = {"target-group-1": {self.target: {"available": True, "knobs": ["sysctl.json"]},
                                        "1.2.3.4": {"available": False, "knobs": ["sysctl.json"]}},
                     "bench-group-1": {self.bench: {"available": True, "destination": {"ip": self.target, "reachable": True}}}}
        msg = "target-group[1]: 1.2.3.4 offline"
        self.run_multi_scense_init(scene_cmd, self.target, self.bench, data_dict, msg)

    def test_init_RBT_multi_target_08(self):
        group1_cmd = r"[target-group-1]\nTARGET_IP = {}\n{}\n{}".format("localhost", self.port, self.scene_2)
        group2_cmd = r"[target-group-2]\nTARGET_IP = {}\n{}\n{}".format(self.target, self.port, self.scene_3)
        scene_cmd = r"\n{}\n{}".format(group1_cmd, group2_cmd)
        data_dict = {"target-group-1": {"localhost": {"available": True, "knobs": ["nginx_conf.json"]}},
                     "target-group-2": {self.target: {"available": True, "knobs": ["sysctl.json", "nginx_conf.json"]}},
                     "bench-group-1": {"localhost": {"available": True, "destination": {"ip": "localhost", "reachable": True}}}}
        self.run_multi_scense_init(scene_cmd, "localhost", "localhost", data_dict)

    def test_init_RBT_multi_target_09(self):
        group1_cmd = r"[target-group-1]\nTARGET_IP = {}, {}\n{}\n{}".format("localhost", self.target, self.port, self.scene_1)
        group2_cmd = r"[target-group-2]\nTARGET_IP = {}, {}\n{}\n{}".format(self.bench, self.brain, self.port, self.scene_3)
        scene_cmd = r"\n{}\n{}".format(group1_cmd, group2_cmd)
        data_dict = {"target-group-1": {"localhost": {"available": True, "knobs": ["sysctl.json"]},
                                        self.target: {"available": True, "knobs": ["sysctl.json"]}},
                     "target-group-2": {self.bench: {"available": True, "knobs": ["sysctl.json", "nginx_conf.json"]},
                                        self.brain: {"available": True, "knobs": ["sysctl.json", "nginx_conf.json"]}},
                     "bench-group-1": {self.bench: {"available": True, "destination": {"ip": self.target, "reachable": True}}}}
        self.run_multi_scense_init(scene_cmd, self.target, self.bench, data_dict)

    def test_init_RBT_multi_bench_01(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}\n{}\n{}".format("localhost", self.port, self.scene_1)
        data_dict = {"target-group-1": {"localhost": {"available": True, "knobs": ["sysctl.json"]}},
                     "bench-group-1": {self.bench: {"available": True, "destination": {"ip": "localhost", "reachable": True}}}}
        self.run_multi_scense_init(scene_cmd, "localhost", self.bench, data_dict)

    def test_init_RBT_multi_bench_02(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}\n{}\n{}".format(self.target, self.port, self.scene_1)
        data_dict = {"target-group-1": {self.target: {"available": True, "knobs": ["sysctl.json"]}},
                     "bench-group-1": {self.bench: {"available": True, "destination": {"ip": self.bench, "reachable": True}}}}
        self.run_multi_scense_init(scene_cmd, self.bench, self.bench, data_dict)

    def test_init_RBT_multi_bench_03(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}\n{}\n{}".format("localhost", self.port, self.scene_1)
        data_dict = {"target-group-1": {"localhost": {"available": True, "knobs": ["sysctl.json"]}},
                     "bench-group-1": {self.bench: {"available": True, "destination": {"ip": self.target, "reachable": True}}}}
        self.run_multi_scense_init(scene_cmd, self.target, self.bench, data_dict)

    def test_init_RBT_multi_bench_04(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}\n{}\n{}".format("localhost", self.port, self.scene_1)
        data_dict = {"target-group-1": {"localhost": {"available": True, "knobs": ["sysctl.json"]}},
                     "bench-group-1": {"localhost": {"available": True, "destination": {"ip": self.target, "reachable": True}}}}
        self.run_multi_scense_init(scene_cmd, self.target, "localhost", data_dict)

    def test_init_RBT_multi_bench_05(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}\n{}\n{}".format("localhost", self.port, self.scene_1)
        bench_ip = "localhost, {}".format(self.bench)
        data_dict = {"target-group-1": {"localhost": {"available": True, "knobs": ["sysctl.json"]}},
                     "bench-group-1": {"localhost": {"available": True, "destination": {"ip": "1.2.3.4", "reachable": False}},
                                       self.bench: {"available": True, "destination": {"ip": "1.2.3.4", "reachable": False}}}}
        msg = "bench destination 1.2.3.4 unreachable"
        self.run_multi_scense_init(scene_cmd, "1.2.3.4", bench_ip, data_dict, msg)

    def test_init_RBT_multi_bench_06(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}\n{}\n{}".format("localhost", self.port, self.scene_1)
        bench_ip = "localhost, {}".format(self.bench)
        data_dict = {"target-group-1": {"localhost": {"available": True, "knobs": ["sysctl.json"]}},
                     "bench-group-1": {"localhost": {"available": True, "destination": {"ip": "localhost", "reachable": True}},
                                       self.bench: {"available": True, "destination": {"ip": "localhost", "reachable": True}}}}
        self.run_multi_scense_init(scene_cmd, "localhost", bench_ip, data_dict)

    def test_init_RBT_multi_bench_07(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}\n{}\n{}".format("localhost", self.port, self.scene_1)
        bench_ip = "{}, {}".format(self.bench, self.target)
        data_dict = {"target-group-1": {"localhost": {"available": True, "knobs": ["sysctl.json"]}},
                     "bench-group-1": {self.target: {"available": True, "destination": {"ip": "localhost", "reachable": True}},
                                       self.bench: {"available": True, "destination": {"ip": "localhost", "reachable": True}}}}
        self.run_multi_scense_init(scene_cmd, "localhost", bench_ip, data_dict)

    def test_init_RBT_multi_bench_08(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}\n{}\n{}".format("localhost", self.port, self.scene_1)
        bench_ip = "{}, localhost".format(self.bench)
        data_dict = {"target-group-1": {"localhost": {"available": True, "knobs": ["sysctl.json"]}},
                     "bench-group-1": {"localhost": {"available": True, "destination": {"ip": self.target, "reachable": True}},
                                       self.bench: {"available": True, "destination": {"ip": self.target, "reachable": True}}}}
        self.run_multi_scense_init(scene_cmd, self.target, bench_ip, data_dict)

    def test_init_RBT_multi_bench_09(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}\n{}\n{}".format("localhost", self.port, self.scene_1)
        bench_ip = "{}, 1.2.3.4".format(self.bench)
        data_dict = {"target-group-1": {"localhost": {"available": True, "knobs": ["sysctl.json"]}},
                     "bench-group-1": {"1.2.3.4": {"available": False, "destination": {"ip": self.target, "reachable": False}},
                                       self.bench: {"available": True, "destination": {"ip": self.target, "reachable": True}}}}
        msg = "bench source 1.2.3.4 offline"
        self.run_multi_scense_init(scene_cmd, self.target, bench_ip, data_dict, msg)

    def test_init_RBT_multi_bench_10(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}\n{}\n{}".format("localhost", self.port, self.scene_1)
        bench_ip = "1.2.3.4, 5.6.7.8".format(self.bench)
        data_dict = {"target-group-1": {"localhost": {"available": True, "knobs": ["sysctl.json"]}},
                     "bench-group-1": {"1.2.3.4": {"available": False, "destination": {"ip": "localhost", "reachable": False}},
                                       "5.6.7.8": {"available": False, "destination": {"ip": "localhost", "reachable": False}}}}
        msg = "bench source 5.6.7.8 offline"
        self.run_multi_scense_init(scene_cmd, "localhost", bench_ip, data_dict, msg)

    def test_init_RBT_multi_target_bench_01(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}, {}, {}\n{}\n{}".format(self.target, "localhost", self.bench, self.port, self.scene_1)
        data_dict = {"target-group-1": {"localhost": {"available": True, "knobs": ["sysctl.json"]},
                                        self.target: {"available": True, "knobs": ["sysctl.json"]},
                                        self.bench: {"available": True, "knobs": ["sysctl.json"]}},
                     "bench-group-1": {"localhost": {"available": True, "destination": {"ip": self.target, "reachable": True}}}}
        self.run_multi_scense_init(scene_cmd, self.target, "localhost", data_dict)

    def test_init_RBT_multi_target_bench_02(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}, {}\n{}\n{}".format(self.target, "localhost", self.port, self.scene_1)
        bench_ip = "localhost, {}".format(self.bench)
        data_dict = {"target-group-1": {"localhost": {"available": True, "knobs": ["sysctl.json"]},
                                        self.target: {"available": True, "knobs": ["sysctl.json"]}},
                     "bench-group-1": {"localhost": {"available": True, "destination": {"ip": self.target, "reachable": True}},
                                       self.bench: {"available": True, "destination": {"ip": self.target, "reachable": True}}}}
        self.run_multi_scense_init(scene_cmd, self.target, bench_ip, data_dict)

    def test_init_RBT_multi_target_bench_03(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}, {}\n{}\n{}".format(self.target, "1.2.3.4", self.port, self.scene_1)
        bench_ip = "5.6.7.8, {}".format(self.bench)
        data_dict = {"target-group-1": {"1.2.3.4": {"available": False, "knobs": ["sysctl.json"]},
                                        self.target: {"available": True, "knobs": ["sysctl.json"]}},
                     "bench-group-1": {"5.6.7.8": {"available": False, "destination": {"ip": "localhost", "reachable": False}},
                                       self.bench: {"available": True, "destination": {"ip": "localhost", "reachable": True}}}}
        msg = "bench source 5.6.7.8 offline"
        self.run_multi_scense_init(scene_cmd, "localhost", bench_ip, data_dict, msg)

    def test_init_RBT_multi_target_bench_04(self):
        scene_cmd = r"\n[target-group-1]\nTARGET_IP = {}, {}\n{}\n{}".format("1.2.3.4", "5.6.7.8", self.port, self.scene_1)
        bench_ip = "1.2.3.4, 5.6.7.8".format(self.bench)
        data_dict = {"target-group-1": {"1.2.3.4": {"available": False, "knobs": ["sysctl.json"]},
                                        "5.6.7.8": {"available": False, "knobs": ["sysctl.json"]}},
                     "bench-group-1": {"1.2.3.4": {"available": False, "destination": {"ip": "localhost", "reachable": False}},
                                       "5.6.7.8": {"available": False, "destination": {"ip": "localhost", "reachable": False}}}}
        msg = "bench source 1.2.3.4 offline"
        self.run_multi_scense_init(scene_cmd, "localhost", bench_ip, data_dict, msg)
