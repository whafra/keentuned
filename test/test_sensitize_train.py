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
from common import runSensitizeCollect

logger = logging.getLogger(__name__)


class TestSensitizeTrain(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain", "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        status = runSensitizeCollect("test01")
        self.assertEqual(status, 0)
        logger.info('start to run test_sensitize_train testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain", "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("test01")
        logger.info('the test_sensitize_train testcase finished')

    def test_sensitize_train(self):
        cmd = 'echo y | keentune sensitize train --data test01 --output test01'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)

        path = re.search(r'\s+"(.*?)"', self.out).group(1)
        time.sleep(3)
        while True:
            with open(path, 'r') as f:
                res_data = f.read()
            if '"sensitize train" finish' in res_data:
                break
            time.sleep(8)

        word_list = ["Step1", "Step2", "Step3", "Step4", '"sensitize train" finish']
        result = all([word in res_data for word in word_list])
        self.assertTrue(result)

        self.path = "/var/keentune/sensitize/sensi-test01.json"
        res = os.path.exists(self.path)
        self.assertTrue(res)
