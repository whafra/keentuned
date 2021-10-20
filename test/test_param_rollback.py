import os
import sys
import logging
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import checkServerStatus
from common import sysCommand

logger = logging.getLogger(__name__)


class TestParamRollback(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_param_rollback testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('the test_param_rollback testcase finished')

    def test_param_rollback(self):
        cmd = 'keentune param rollback'
        self.status, self.out, _ = sysCommand(cmd)

        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('param rollback successfully'))
