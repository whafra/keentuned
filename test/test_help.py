import os
import sys
import logging
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import checkServerStatus
from common import sysCommand

logger = logging.getLogger(__name__)


class TestHelp(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_help testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('the test_help testcase finished')

    def test_help(self):
        cmd = 'keentune help'
        self.status, self.out, _ = sysCommand(cmd)

        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('Usage'))
        self.assertTrue(self.out.__contains__('Available Commands'))
