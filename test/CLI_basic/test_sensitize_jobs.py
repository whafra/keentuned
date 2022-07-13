import os
import sys
import unittest
import logging

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import sysCommand
from common import checkServerStatus
from common import deleteDependentData
from common import runParamTune
from common import getTaskLogPath
from common import getTrainTaskResult

logger = logging.getLogger(__name__)


class TestSensitizeJobs(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain", "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        status = runParamTune("param1")
        self.assertEqual(status, 0)
        logger.info('start to run test_sensitize_jobs testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain", "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("param1")
        logger.info('the test_sensitize_jobs testcase finished')

    def test_sensitize_jobs_FUN(self):
        cmd = 'echo y | keentune sensitize train --data param1 --job param1'
        path = getTaskLogPath(cmd)
        result = getTrainTaskResult(path)
        self.assertTrue(result)

        cmd = 'keentune sensitize jobs'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('param1'))
