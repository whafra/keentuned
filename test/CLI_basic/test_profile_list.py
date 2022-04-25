import os
import sys
import unittest
import logging

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import sysCommand
from common import checkServerStatus
from common import deleteDependentData
from common import runParamTune
from common import runParamDump

logger = logging.getLogger(__name__)


class TestProfileList(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-target"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        status = runParamTune("param1")
        self.assertEqual(status, 0)
        status = runParamDump("param1")
        self.assertEqual(status, 0)
        logger.info('start to run test_profile_list testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-target"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("param1")
        logger.info('the test_profile_list testcase finished')

    def test_profile_list_FUN(self):
        cmd = 'keentune profile list'
        self.status, self.out, _ = sysCommand(cmd)

        file_list = ["cpu_high_load.conf",
                     "net_high_throuput.conf", "param1_group1.conf"]
        result = all([file in self.out for file in file_list])
        self.assertEqual(self.status, 0)
        self.assertTrue(result)
