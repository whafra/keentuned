import os
import sys
import logging
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import sysCommand
from common import checkServerStatus

logger = logging.getLogger(__name__)


class TestInit(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_init testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        os.remove(self.path)
        logger.info('the test_init testcase finished')

    def test_init_FUN(self):
        cmd = 'keentune init'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("KeenTune Init success", self.out)

        self.path = "/etc/keentune/conf/init.yaml"
        res = os.path.exists(self.path)
        self.assertTrue(res)
