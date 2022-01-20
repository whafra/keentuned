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
        status = runSensitizeCollect("test01")
        self.assertEqual(status, 0)
        logger.info('start to run test_sensitize_delete testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain", "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("test01")
        logger.info('the test_sensitize_delete testcase finished')

    def test_sensitize_delete(self):
        cmd = 'echo y | keentune sensitize delete --data test01'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('delete successfully'))

        cmd = 'keentune sensitize list'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertFalse(self.out.__contains__('test01'))

        self.path = "/var/keentune/data/tuning_data/collect/'test01[{}]'".format(self.algorithm)
        res = os.path.exists(self.path)
        self.assertFalse(res)
