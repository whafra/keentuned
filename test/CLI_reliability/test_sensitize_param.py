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
from common import getSysBackupData
from common import checkBackupData
from common import getTaskLogPath
from common import getTuneTaskResult
from common import getCollectTaskResult
from common import getTrainTaskResult

logger = logging.getLogger(__name__)


class TestSensitizeParam(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_sensitize_param testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("param1")
        logger.info('the test_sensitize_param testcase finished')

    def run_task(self, cmd, status, msg):
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, status)
        self.assertTrue(self.out.__contains__(msg))

    def test_sensitize_param_RBT_run(self):
        tune_cmd = 'keentune param tune -i 15 --job param1'
        train_cmd = 'echo y | keentune sensitize train --data param1 --job param1 -t 5'
        path = getTaskLogPath(tune_cmd)
        time.sleep(10)
        self.run_task(train_cmd, 1, "job 'param1' is running, can not training sensibility model")
        result = getTuneTaskResult(path)
        self.assertTrue(result)

        path = getTaskLogPath(train_cmd)
        time.sleep(3)
        tune_cmd = 'keentune param tune -i 10 --job param2'
        self.run_task(tune_cmd, 1, "Another sensitizing job param1 is running")
        result = getTrainTaskResult(path)
        self.assertTrue(result)
        deleteDependentData("param1")
