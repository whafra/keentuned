import os
import sys
import logging
import re
import subprocess
import time
import unittest

logger = logging.getLogger(__name__)
sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import sysCommand
from common import checkServerStatus
from common import deleteDependentData
from common import runSensitizeCollect
from common import getTaskLogPath
from common import getCollectTaskResult
from common import copyTmpFile


class TestSensitizeCollect(unittest.TestCase):
    @classmethod
    def setUpClass(self) -> None:
        logger.info("TestSensitizeCollect begin...")

    @classmethod
    def tearDownClass(self) -> None:
        deleteDependentData("sensitize1")
        logger.info("TestSensitizeCollect end...")

    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain", "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_sensitize_collect testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain", "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('the test_sensitize_collect testcase finished')

    def check_sensitize_collect_data(self, data):
        cmd = 'keentune sensitize list'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__(data))

    def test_sensitize_collect_RBT_lose_data_param(self):
        cmd = 'keentune sensitize collect -i 1'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('Incomplete or Unmatched command'))

    def test_sensitize_collect_RBT_lose_data_value(self):
        cmd = 'keentune sensitize collect -i 1 --data'
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__('flag needs an argument: --data'))

    def test_sensitize_collect_RBT_data_value_null(self):
        cmd = "keentune sensitize collect -i 1 --data ''"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('Incomplete or Unmatched command'))

    def test_sensitize_collect_RBT_data_value_empty(self):
        cmd = "keentune sensitize collect -i 1 --data ' '"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__("""--data find unexpected characters ' ' . Only "a-z", "A-Z", "0-9" and "_" are supported"""))

    def test_sensitize_collect_RBT_data_value_repeat(self):
        self.status = runSensitizeCollect("sensitize1")
        self.assertEqual(self.status, 0)
        cmd = 'keentune sensitize collect -i 1 --data sensitize1'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__("already exists"))
        deleteDependentData("sensitize1")

    def test_sensitize_collect_RBT_lose_iteration_value(self):
        cmd = "keentune sensitize collect -i --data sensitize3"
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__('invalid syntax'))

    def test_sensitize_collect_RBT_iteration_value_error(self):
        cmd = "keentune sensitize collect -i -o --data sensitize3"
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__('invalid syntax'))

    def test_sensitize_collect_RBT_iteration_value_null(self):
        cmd = "keentune sensitize collect -i '' --data sensitize3"
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__('invalid syntax'))

    def test_sensitize_collect_RBT_iteration_value_empty(self):
        cmd = "keentune sensitize collect -i ' ' --data sensitize3"
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__('invalid syntax'))

    def test_sensitize_collect_RBT_lose_debug_value(self):
        cmd = 'keentune sensitize collect -i 1 --data sensitize4 --debug'
        path = getTaskLogPath(cmd)
        while True:
            with open(path, 'r') as f:
                res_data = f.read()
            if 'Sensitization collection finished' in res_data:
                break
            time.sleep(8)

        word_list = ["Step1", "Step4", "Benchmark result:", "Sensitization collection finished"]
        result = all([word in res_data for word in word_list])
        self.assertTrue(result)
        deleteDependentData("sensitize4")

    def test_sensitize_collect_RBT_debug_value_error(self):
        cmd = "keentune sensitize collect -i 1 --data sensitize4 --debug -0"
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__("unknown shorthand flag: '0' in -0"))

    def test_sensitize_collect_RBT_lose_optional_param(self):
        cmd = 'keentune sensitize collect --data sensitize5'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)

        pattern = re.compile(r'iteration:\s*(\d+)')
        iteration = re.search(pattern, self.out).group(1)
        self.assertEqual(iteration, "100")
        
        cmd = 'keentune sensitize stop'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('Abort sensibility identification job'))
        time.sleep(5)
        deleteDependentData("sensitize5")

    def test_sensitize_collect_RBT_data_special_chars_01(self):
        cmd = 'keentune sensitize collect -i 1 --data _sensitize11'
        path = getTaskLogPath(cmd)
        result = getCollectTaskResult(path)
        self.assertTrue(result)
        self.check_sensitize_collect_data("_sensitize11")
        deleteDependentData("_sensitize11")

    def test_sensitize_collect_RBT_data_special_chars_02(self):
        cmd = 'keentune sensitize collect -i 1 --data /root/test'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__("--data find unexpected characters '/'"))

    def test_sensitize_collect_RBT_data_special_chars_03(self):
        cmd = 'keentune sensitize collect -i 1 --data test,a'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__("--data find unexpected characters ','"))

    def test_sensitize_collect_RBT_data_special_chars_04(self):
        cmd = 'keentune sensitize collect -i 1 --data  _test+1.1'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__("--data find unexpected characters '+' '.'"))

    def test_sensitize_collect_RBT_data_special_chars_05(self):
        cmd = 'keentune sensitize collect -i 1 --data sensitize3 a'
        path = getTaskLogPath(cmd)
        result = getCollectTaskResult(path)
        self.assertTrue(result)
        self.check_sensitize_collect_data("sensitize3")
        deleteDependentData("sensitize3")

    def test_sensitize_collect_RBT_iteration_value_negative(self):
        cmd = 'keentune sensitize collect -i -1 --data sensitize3'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__("--iteration must be positive integer, input: -1"))

    def test_sensitize_collect_RBT_iteration_value_zero(self):
        cmd = 'keentune sensitize collect -i 0 --data sensitize3'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__("--iteration must be positive integer, input: 0"))

    def test_sensitize_collect_RBT_iteration_value_float(self):
        cmd = 'keentune sensitize collect -i 5.2 --data sensitize3'
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__("Error: invalid argument"))

    def test_sensitize_collect_RBT_iteration_value_str(self):
        cmd = 'keentune sensitize collect -i "hello" --data sensitize3'
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__("Error: invalid argument"))

    def test_sensitize_collect_RBT_iteration_value_str_number(self):
        cmd = 'keentune sensitize collect -i "5" --data sensitize3'
        path = getTaskLogPath(cmd)
        result = getCollectTaskResult(path)
        self.assertTrue(result)
        self.check_sensitize_collect_data("sensitize3")
        deleteDependentData("sensitize3")

    def test_sensitize_collect_RBT_multi_config(self):
        cmd = "sh conf/reset_keentuned.sh {} '{}'".format("param", "sysctl.json, nginx.json")
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("restart keentuned server successfully!", self.out)

        cmd = 'keentune sensitize collect -i 2 --data sensitize3'
        path = getTaskLogPath(cmd)
        result = getCollectTaskResult(path)
        self.assertTrue(result)
        self.check_sensitize_collect_data("sensitize3")
        deleteDependentData("sensitize3")