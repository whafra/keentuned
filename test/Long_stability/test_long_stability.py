import os
import re
import sys
import logging
import time
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import sysCommand
from common import runParamTune
from common import getTaskLogPath
from common import checkServerStatus
from common import getTrainTaskResult
from common import deleteDependentData
from common import runSensitizeCollect

logger = logging.getLogger(__name__)


class TestLongStability(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain", "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_long_stability testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain", "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('the test_long_stability testcase finished')

    def profile_list(self, state):
        cmd = 'keentune profile list'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.result = re.search(r'\[(.*?)\].+param1_group1.conf', self.out).group(1)
        self.assertIn(state, self.result)

    def profile_set(self):
        cmd = 'echo y | keentune param dump -j param1'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('dump successfully'))

        cmd = 'keentune profile set --group1 param1_group1.conf'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('Set param1_group1.conf successfully'))

        self.profile_list("active")

    def profile_rollback(self):
        self.profile_list("active")
        cmd = 'keentune profile rollback'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('profile rollback successfully'))
        self.profile_list("available")

    def test_long_stability (self):
        while True:
            result = runParamTune("param1", iteration=500)
            self.assertEqual(result, 0)
            time.sleep(5)

            result = runSensitizeCollect("sensitize1", iteration=500)
            self.assertEqual(result, 0)
            time.sleep(5)

            cmd = "echo y | keentune sensitize train --data sensitize1 --output sensitize1"
            path = getTaskLogPath(cmd)
            result = getTrainTaskResult(path)
            self.assertTrue(result)
            time.sleep(2)

            for i in range(100):
                self.profile_set()
                time.sleep(2)
                self.profile_rollback()
                time.sleep(2)

            deleteDependentData("param1")
            deleteDependentData("sensitize1")
            cmd = 'echo y | keentune profile delete --name param1_group1.conf'
            self.status, self.out, _ = sysCommand(cmd)
            self.assertEqual(self.status, 0)
            self.assertTrue(self.out.__contains__('delete successfully'))

            logger.info("current round testcase finished")
