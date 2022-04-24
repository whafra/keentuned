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


class TestProfileDelete(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-target"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        status = runParamTune("param1")
        self.assertEqual(status, 0)
        status = runParamDump("param1")
        self.assertEqual(status, 0)
        logger.info('start to run test_profile_delete testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-target"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("param1")
        logger.info('the test_profile_delete testcase finished')

    def test_profile_delete_FUN(self):
        cmd = 'echo y | keentune profile delete --name param1_group1.conf'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('delete successfully'))

        path = "/var/keentune/profile/param1_group1.conf"
        res = os.path.exists(path)
        self.assertFalse(res)

        cmd = 'echo y | keentune profile delete --name cpu_high_load.conf'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__('not supported to delete'))
