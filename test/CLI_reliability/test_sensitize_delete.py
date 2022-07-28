import os
import sys
import logging
import subprocess
import unittest

logger = logging.getLogger(__name__)
sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import sysCommand
from common import checkServerStatus
from common import deleteDependentData
from common import runParamTune

class TestSensitizeDelete(unittest.TestCase):
    @classmethod
    def setUpClass(self) -> None:
        logger.info("TestSensitizeDelete begin...")
        status = runParamTune("param1")
        assert status == 0

    @classmethod
    def tearDownClass(self) -> None:
        deleteDependentData("param1")
        logger.info("TestSensitizeDelete end...") 

    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-target"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_sensitize_delete testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-target"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('the test_sensitize_delete testcase finished')

    def test_sensitize_delete_RBT_lose_data_param(self):
        cmd = 'keentune sensitize delete'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__('Incomplete or Unmatched command'))

    def test_sensitize_delete_RBT_lose_data_value(self):
        cmd = 'keentune sensitize delete --job'
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__('flag needs an argument: --job'))

    def test_sensitize_delete_RBT_data_value_null(self):
        cmd = "keentune sensitize delete --job ''"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__('Incomplete or Unmatched command'))

    def test_sensitize_delete_RBT_data_value_empty(self):
        cmd = "keentune sensitize delete --job ' '"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__('Incomplete or Unmatched command'))
