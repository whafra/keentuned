import os
import re
import sys
import logging
import time
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import deleteDependentData
from common import checkServerStatus
from common import runParamTune
from common import sysCommand
from common import getTuneTaskResult

logger = logging.getLogger(__name__)


class TestParamTuneDelete(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        status = runParamTune("param1")
        self.assertEqual(status, 0)
        logger.info('start to run test_param_tune_delete testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        deleteDependentData("param1")
        self.assertEqual(status, 0)

        logger.info('the test_param_tune_delete testcase finished')

    def delete_job_data(self, path):
        cmd = "echo y | keentune param delete --job param1"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('delete successfully'))
        res = os.path.exists(path)
        self.assertFalse(res)

    def test_param_tune_RBT_delete(self):
        cmd = 'keentune param tune  -i 10 --job param1'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__("--job the specified name 'param1' already exists"))

        job_path = "/var/keentune/tuning_workspace/param1"
        res = os.path.exists(job_path)
        self.assertTrue(res)
        self.delete_job_data(job_path)

        cmd = 'keentune param tune -i 10 --job param1'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        log_path = re.search(r'\s+"(.*?)"', self.out).group(1)
        time.sleep(2)
        self.status, self.out, _ = sysCommand('echo y | keentune param delete --job param1')
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('Auto-tuning job param1 is running'))

        result = getTuneTaskResult(log_path)
        self.assertTrue(result)
        self.delete_job_data(job_path)
