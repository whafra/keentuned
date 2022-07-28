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
from common import runParamDump


class TestProfileGenerate(unittest.TestCase):
    @classmethod
    def setUpClass(self) -> None:
        logger.info("TestProfileGenerate begin...")
        status = runParamTune("param1")
        assert status == 0
        status = runParamDump("param1")
        assert status == 0

    @classmethod
    def tearDownClass(self) -> None:
        deleteDependentData("param1")
        logger.info("TestProfileGenerate end...") 

    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-target"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_profile_generate testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-target"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('the test_profile_generate testcase finished')

    def test_profile_generate_RBT_lose_params(self):
        cmd = 'echo y | keentune profile generate'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__('Incomplete or Unmatched command'))

    def test_profile_generate_RBT_lose_name_param(self):
        cmd = 'echo y | keentune profile generate -o param1'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__('Incomplete or Unmatched command'))

    def test_profile_generate_RBT_lose_name_value(self):
        cmd = 'echo y | keentune profile generate -n -o param1'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__('profile.Generate failed'))

    def test_profile_generate_RBT_name_value_empty(self):
        cmd = "echo y | keentune profile generate -n ' ' -o param1"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__('Incomplete or Unmatched command'))

    def test_profile_generate_RBT_name_value_null(self):
        cmd = "echo y | keentune profile generate -n '' -o param1"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__('Incomplete or Unmatched command'))

    def test_profile_generate_RBT_lose_output_param(self):
        cmd = 'echo y | keentune profile generate -n param1_group1.conf'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('generate successfully'))

    def test_profile_generate_RBT_lose_output_value(self):
        cmd = 'echo y | keentune profile generate -n param1_group1.conf -o'
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__("flag needs an argument: 'o' in -o"))

    def test_profile_generate_RBT_output_value_empty(self):
        cmd = "echo y | keentune profile generate -n param1_group1.conf -o ' '"
        self.status,self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('generate successfully'))

    def test_profile_generate_RBT_output_value_null(self):
        cmd = "echo y | keentune profile generate -n param1_group1.conf -o ''"
        self.status,self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('generate successfully'))
