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


class TestProfileInfo(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-target"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        status = runParamTune("param1")
        self.assertEqual(status, 0)
        status = runParamDump("param1")
        self.assertEqual(status, 0)
        logger.info('start to run test_profile_info testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-target"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("param1")
        logger.info('the test_profile_info testcase finished')

    def test_profile_info_FUN(self):
        cmd = 'keentune profile info --name param1_group1.conf'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('[sysctl]'))

        cmd = 'keentune profile info --name cpu_high_load.conf'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('[sysctl]'))
