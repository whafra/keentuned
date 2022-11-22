import os
import re
import sys
import logging
import time
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import deleteDependentData
from common import checkServerStatus
from common import sysCommand
from common import getTuneTaskResult
from common import getTaskLogPath
from common import runParamTune

logger = logging.getLogger(__name__)


class TestMultiScenes(unittest.TestCase):
    @classmethod
    def setUpClass(self) -> None:
        self.brain = self.get_server_ip()
        if self.brain != "localhost":
            status = sysCommand("scp conf/restart_brain.sh {}:/opt".format(self.brain))[0]
            assert status == 0
        
    @classmethod
    def tearDownClass(self) -> None:
        if self.brain != "localhost":
            status = sysCommand("ssh {} 'rm -rf /opt/restart_brain.sh'".format(self.brain))[0]
            assert status == 0

    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_multiple_scenes testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("param1")
        logger.info('the test_multiple_scenes testcase finished')

    @staticmethod
    def get_server_ip():
        with open("common.py", "r", encoding='UTF-8') as f:
            data = f.read()
        brain = re.search(r"brain_ip=\"(.*)\"", data).group(1)
        return brain

    def check_param_tune_job(self, name):
        cmd = 'keentune param jobs'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__(name))

    def run_sensitize_train(self, name):
        cmd = "echo y | keentune sensitize train --data {0} --job {0}".format(name)
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)

        path = re.search(r'\s+"(.*?)"', self.out).group(1)
        time.sleep(3)
        while True:
            with open(path, 'r') as f:
                res_data = f.read()
            if "identification results successfully" in res_data or "[ERROR]" in res_data:
                break
            time.sleep(8)

        word_list = ["Step1", "Step2", "Step3", "identification results successfully"]
        result = all([word in res_data for word in word_list])
        self.assertTrue(result)

        self.path = "/var/keentune/sensitize_workspace/{}/knobs.json".format(name)
        res = os.path.exists(self.path)
        self.assertTrue(res)

    def restart_brain_server(self, algorithm, flag):
        cmd = "sh conf/restart_brain.sh {} {}".format(algorithm, flag)
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('restart brain server successfully!'))

    def reset_keentuned(self, config, file):
        cmd = "sh conf/reset_keentuned.sh {} {}".format(config, file)
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("restart keentuned server successfully!", self.out)

    def test_param_tune_FUN_nginx(self):
        self.reset_keentuned("param", "nginx_conf.json")
        cmd = 'keentune param tune -i 10 --job param1'
        path = getTaskLogPath(cmd)
        result = getTuneTaskResult(path)
        self.assertTrue(result)
        self.check_param_tune_job("param1")
        self.reset_keentuned("param", "sysctl.json")

    def test_param_tune_FUN_tpe(self):
        self.restart_brain_server("tpe", "tune")
        status = runParamTune("param1")
        self.assertEqual(status, 0)

    def test_param_tune_FUN_hord(self):
        self.restart_brain_server("hord", "tune")
        status = runParamTune("param1")
        self.assertEqual(status, 0)

    def test_param_tune_FUN_random(self):
        self.restart_brain_server("random", "tune")
        status = runParamTune("param1")
        self.assertEqual(status, 0)

    def test_param_tune_FUN_lamcts(self):
        self.restart_brain_server("lamcts", "tune")
        status = runParamTune("param1")
        self.assertEqual(status, 0)

    def test_param_tune_FUN_bgcs(self):
        self.restart_brain_server("bgcs", "tune")
        status = runParamTune("param1")
        self.assertEqual(status, 0)
    
    def test_sensitize_train_FUN_lasso(self):
        self.restart_brain_server("lasso", "train")
        status = runParamTune("param1")
        self.assertEqual(status, 0)
        self.run_sensitize_train("param1")

    def test_sensitize_train_FUN_univariate(self):
        self.restart_brain_server("univariate", "train")
        status = runParamTune("param1")
        self.assertEqual(status, 0)
        self.run_sensitize_train("param1")

    def test_sensitize_train_FUN_gp(self):
        self.restart_brain_server("gp", "train")
        status = runParamTune("param1")
        self.assertEqual(status, 0)
        self.run_sensitize_train("param1")

    def test_sensitize_train_FUN_shap(self):
        self.restart_brain_server("shap", "train")
        status = runParamTune("param1")
        self.assertEqual(status, 0)
        self.run_sensitize_train("param1")

    def test_sensitize_train_FUN_explain(self):
        self.restart_brain_server("explain", "train")
        status = runParamTune("param1")
        self.assertEqual(status, 0)
        self.run_sensitize_train("param1")
