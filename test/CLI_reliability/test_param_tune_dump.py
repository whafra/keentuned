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


class TestParamTuneDump(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        status = runParamTune("param1")
        self.assertEqual(status, 0)
        logger.info('start to run test_param_tune_dump testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("param1")
        deleteDependentData("param2")
        logger.info('the test_param_tune_dump testcase finished')

    def check_file(self, name):
        path = "/var/keentune/profile/{}_group1.conf".format(name)
        res = os.path.exists(path)
        self.assertTrue(res)

    def dump_job_data(self, name):
        cmd = 'echo y | keentune param dump -j {}'.format(name)
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('dump successfully'))
        self.check_file(name)

    def test_param_tune_RBT_dump(self):
        self.dump_job_data("param1")
        self.dump_job_data("param1")
        cmd = 'echo n | keentune param dump -j param1'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__('outputFile exist and you have given up to overwrite it'))
        self.check_file("param1")

        cmd = 'keentune param tune -i 10 --job param2'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        log_path = re.search(r'\s+"(.*?)"', self.out).group(1)
        time.sleep(2)
        self.status, self.out, _ = sysCommand('echo y | keentune param dump -j param2')
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__('job param2 status is running'))
        self.dump_job_data("param1")

        result = getTuneTaskResult(log_path)
        self.assertTrue(result)
        self.dump_job_data("param2")
