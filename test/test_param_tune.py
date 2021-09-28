import os
import sys
import logging
import time
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import deleteDependentData
from common import checkServerStatus
from common import sysCommand

logger = logging.getLogger(__name__)


class TestParamTune(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_param_tune testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("test1")
        logger.info('the test_param_tune testcase finished')

    def test_param_tune(self):
        cmd = 'keentune param tune --param_conf parameter/param_100.json -i 1 --bench_conf benchmark/wrk/bench_wrk_nginx_long.json --name test1'
        status, _, _ = sysCommand(cmd)
        self.assertEqual(status, 0)

        while True:
            cmd = "keentune msg --name 'param tune'"
            self.status, self.out, _ = sysCommand(cmd)
            if '[BEST] Tuning improvment' in self.out:
                break
            time.sleep(8)

        word_list = ["Step1", "Step2", "Step3", "Step4",
                     "Step5", "Step6", "[BEST] Tuning improvment"]
        result = all([word in self.out for word in word_list])

        self.assertEqual(self.status, 0)
        self.assertTrue(result)
