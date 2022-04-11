import os
import re
import sys
import time
import unittest
import logging

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import sysCommand
from common import checkServerStatus
from common import deleteDependentData

logger = logging.getLogger(__name__)

class TestSensitizeCollect(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain", "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_sensitize_collect testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain", "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("test01")
        logger.info('the test_sensitize_collect testcase finished')

    def test_sensitize_collect(self):
        cmd = 'keentune sensitize collect -i 10 --data test01'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)

        path = re.search(r'\s+"(.*?)"', self.out).group(1)
        time.sleep(3)
        while True:
            with open(path, 'r') as f:
                res_data = f.read()
            if 'Sensitization collection finished' in res_data:
                break
            time.sleep(8)

        word_list = ["Step1", "Step2", "Step3", "Step4", "Sensitization collection finished"]
        result = all([word in res_data for word in word_list])
        self.assertTrue(result)

        cmd = 'keentune sensitize list'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('test01'))
