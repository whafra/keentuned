import os
import sys
import logging
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import sysCommand
from common import checkServerStatus

logger = logging.getLogger(__name__)


class TestVersion(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_version testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('the test_version testcase finished')

    def test_version_FUN(self):
        cmd = 'keentune -v'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("keentune version", self.out)
