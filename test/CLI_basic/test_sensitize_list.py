import os
import sys
import unittest
import logging

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import sysCommand
from common import checkServerStatus
from common import deleteDependentData
from common import runParamTune
from common import runSensitizeCollect

logger = logging.getLogger(__name__)


class TestSensitizeList(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain", "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        status = runParamTune("param1")
        self.assertEqual(status, 0)
        status = runSensitizeCollect("sensitize1")
        self.assertEqual(status, 0)
        logger.info('start to run test_sensitize_list testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain", "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("param1")
        deleteDependentData("sensitize1")
        logger.info('the test_sensitize_list testcase finished')

    def test_sensitize_list_FUN(self):
        cmd = 'keentune sensitize list'
        self.status, self.out, _ = sysCommand(cmd)

        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('identification results successfully'))
        self.assertTrue(self.out.__contains__('param1'))
        self.assertTrue(self.out.__contains__('sensitize1'))