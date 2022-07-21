import os
import re
import sys
import logging
import time
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import sysCommand
from common import checkServerStatus
from common import getTuneTaskResult
from common import getTaskLogPath
from common import deleteDependentData

logger = logging.getLogger(__name__)


class TestParamTuneStop(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        self.cmd = 'keentune param tune -i 10 --job param1'
        logger.info('start to run test_param_tune_stop testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("param1")
        logger.info('the test_param_tune_stop testcase finished')

    def param_tune_stop(self):
        cmd = 'keentune param stop'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('Abort parameter optimization job'))
        time.sleep(5)

    def test_param_tune_RBT_stop(self):
        self.status, self.out, _  = sysCommand(self.cmd)
        self.assertEqual(self.status, 0)
        time.sleep(7)
        self.param_tune_stop()
        deleteDependentData("param1")
        
        path = getTaskLogPath(self.cmd)
        result = getTuneTaskResult(path)
        self.assertTrue(result)

    def test_param_tune_RBT_stop_log(self):
        path = getTaskLogPath(self.cmd)
        self.param_tune_stop()

        with open(path, 'r') as f:
            log_data = f.read()

        self.assertTrue(log_data.__contains__("parameter optimization job abort!"))
