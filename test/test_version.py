import os
import sys
import logging
import unittest

from common import sysCommand
from common import checkServerStatus

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))


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

    def test_version(self):
        cmd = 'keentune version'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("keentune version", self.out)
