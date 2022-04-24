import os
import sys
import logging
import unittest

from common import sysCommand
from common import checkServerStatus

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))


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

    def test_help_FUN(self):
        cmd = 'keentune help'
        self.status, self.out, _ = sysCommand(cmd)

        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('Usage'))
        self.assertTrue(self.out.__contains__('Available Commands'))
