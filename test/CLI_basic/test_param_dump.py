import os
import sys
import logging
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import runParamTune
from common import deleteDependentData
from common import checkServerStatus
from common import sysCommand

logger = logging.getLogger(__name__)


class TestParamDump(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        status = runParamTune("param1")
        self.assertEqual(status, 0)
        logger.info('start to run test_param_dump testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("param1")
        logger.info('the test_param_dump testcase finished')

    def test_param_dump_FUN(self):
        cmd = 'echo y | keentune param dump -j param1'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('dump successfully'))

        cmd = 'keentune profile list'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__("param1_group1.conf"))

        self.path = "/var/keentune/profile/param1_group1.conf"
        res = os.path.exists(self.path)
        self.assertTrue(res)
