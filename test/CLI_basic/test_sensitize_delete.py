import os
import sys
import unittest
import logging

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import sysCommand
from common import checkServerStatus
from common import deleteDependentData
from common import runSensitizeCollect

logger = logging.getLogger(__name__)

class TestSensitizeDelete(unittest.TestCase):
    def setUp(self) -> None:
        self.algorithm = "Random"
        server_list = ["keentuned", "keentune-brain", "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        status = runSensitizeCollect("sensitize1")
        self.assertEqual(status, 0)
        logger.info('start to run test_sensitize_delete testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain", "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("sensitize1")
        logger.info('the test_sensitize_delete testcase finished')

    def test_sensitize_delete_FUN(self):
        cmd = 'echo y | keentune sensitize delete --data sensitize1'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('delete successfully'))

        cmd = 'keentune sensitize list'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertFalse(self.out.__contains__('sensitize1'))

        self.path = "/var/keentune/data/tuning_data/collect/'sensitize1[{}]'".format(self.algorithm)
        res = os.path.exists(self.path)
        self.assertFalse(res)
