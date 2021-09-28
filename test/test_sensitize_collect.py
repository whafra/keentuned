import os
import sys
import time
import unittest
import logging

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import sysCommand
from common import checkServerStatus
from common import deleteDependentData

logger = logging.getLogger(__name__)

class TestSensitizeCollect(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain", "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_sensitize_collect testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain", "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("test2")
        logger.info('the test_sensitize_collect testcase finished')

    def test_sensitize_collect(self):
        cmd = 'keentune sensitize collect -i 1 --param_conf parameter/param_100.json --bench_conf benchmark/wrk/bench_wrk_nginx_long.json --name test2'
        status, _, _ = sysCommand(cmd)
        self.assertEqual(status, 0)

        while True:
            cmd = "keentune msg --name 'sensitize collect'"
            self.status, self.out, _ = sysCommand(cmd)
            if 'Sensitization collection finished' in self.out:
                break
            time.sleep(8)

        word_list = ["Step1", "Step2", "Step3", "Step4", "Sensitization collection finished"]
        result = all([word in self.out for word in word_list])
        self.assertEqual(self.status, 0)
        self.assertTrue(result)

        cmd = 'keentune sensitize list'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('test2'))
