import os
import sys
import unittest
import logging

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import runParamDump
from common import runParamTune
from common import deleteDependentData
from common import checkServerStatus
from common import sysCommand

logger = logging.getLogger(__name__)


class TestProfileGenerate(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-target"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        status = runParamTune("param1")
        self.assertEqual(status, 0)
        status = runParamDump("param1")
        self.assertEqual(status, 0)
        logger.info('start to run test_profile_generate testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-target"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("param1")
        logger.info('the test_profile_generate testcase finished')

    def test_profile_generate_FUN(self):
        cmd = 'echo y | keentune profile generate -n param1_group1.conf -o param1_group1'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('generate successfully'))

        cmd = 'keentune param jobs'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__("param1"))

        path = "/var/keentune/parameter/generate/param1_group1.json"
        res = os.path.exists(path)
        self.assertTrue(res)
