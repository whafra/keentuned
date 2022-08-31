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


class TestParamDump(unittest.TestCase):
    @classmethod
    def setUpClass(self) -> None:
        logger.info("TestParamDump begin...")
        status = runParamTune("param1")
        assert status == 0

    @classmethod
    def tearDownClass(self) -> None:
        deleteDependentData("param1")
        logger.info("TestParamDump end...") 

    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain", "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        
        self.tune_name = "param1"
        self.path = "/var/keentune/profile/{}.conf".format(self.tune_name)
        logger.info('start to run test_param_dump testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain", "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('the test_param_dump testcase finished')

    def test_param_dump_RBT_lose_job_param(self):
        cmd = 'keentune param dump'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__('Incomplete or Unmatched command'))

    def test_param_dump_RBT_lose_job_value(self):
        cmd = 'echo y | keentune param dump -j'
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__("flag needs an argument: 'j' in -j"))

    def test_param_dump_RBT_job_value_empty(self):
        cmd = "echo y | keentune param dump -j ' '"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__('Incomplete or Unmatched command'))

    def test_param_dump_RBT_job_value_null(self):
        cmd = "echo y | keentune param dump -j ''"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__('Incomplete or Unmatched command'))
