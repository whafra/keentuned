import os
import re
import sys
import logging
import time
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import deleteDependentData
from common import checkServerStatus
from common import sysCommand
from common import getTaskLogPath
from common import getTuneTaskResult
from common import getCollectTaskResult
from common import getTrainTaskResult

logger = logging.getLogger(__name__)


class TestSensitizeProfile(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_sensitize_profile testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("sensitize1")
        logger.info('the test_sensitize_profile testcase finished')

    def check_list_status(self, name):
        cmd = 'keentune profile list'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        if name:
            self.result = re.search(r'\[(.*?)\].+{}.conf'.format(name), self.out).group(1)
            self.assertTrue(self.result.__contains__('active'))
        else:
            self.assertFalse(self.out.__contains__('[active]'))

    def profile_set_data(self, name, status, msg):
        file_path = "conf/{}.conf".format(name)
        profile_path = "/var/keentune/profile"

        cmd = 'cp {} {}'.format(file_path, profile_path)
        self.status, _, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        cmd = "keentune profile set --group1 {}.conf".format(name)
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, status)
        self.assertTrue(self.out.__contains__(msg))

    def profile_rollback(self, status, msg):
        cmd = 'keentune profile rollback'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, status)
        self.assertTrue(self.out.__contains__(msg))

    def profile_list(self, name=""):
        cmd = 'keentune profile list'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.check_list_status(name)

    def test_sensitize_profile_RBT_run(self):
        cmd = 'keentune sensitize collect -i 10 --data sensitize1'
        path = getTaskLogPath(cmd)
        self.profile_set_data("profile1", 1, 'job collect sensitize1 is running')
        self.profile_rollback(1, 'job collect sensitize1 is running')
        self.profile_list()
        result = getCollectTaskResult(path)
        self.assertTrue(result)

        cmd = 'echo y | keentune sensitize train --data sensitize1 --output sensitize1 -t 10'
        path = getTaskLogPath(cmd)
        self.profile_set_data("profile1", 0, 'Set profile1.conf successfully')
        self.profile_list("profile1")
        self.profile_rollback(0, 'profile rollback successfully')
        self.profile_list()
        result = getTrainTaskResult(path)
        self.assertTrue(result)

        os.remove("/var/keentune/profile/profile1.conf")
        os.remove("/var/keentune/sensitize/sensi-sensitize1.json")