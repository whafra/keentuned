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
from common import runParamTune
from common import deleteDependentData
from common import getSysBackupData
from common import checkBackupData
from common import checkProfileData
from common import deleteTmpFiles

logger = logging.getLogger(__name__)


class TestParamProfileRollback(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        self.msg = "All Targets No Need to Rollback"
        getSysBackupData()
        logger.info('start to run test_param_profile_rollback testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("param1")
        deleteDependentData("param2")
        deleteTmpFiles(["param1", "param2", "profile1", "profile2"])
        logger.info('the test_param_profile_rollback testcase finished')

    def rollback_data(self, name, msg=None):
        cmd = 'keentune {} rollback'.format(name)
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        if msg:
            self.assertTrue(self.out.__contains__(msg))
        else:
            self.assertTrue(self.out.__contains__('{} rollback successfully'.format(name)))

    def param_tune_test(self, name, backup=False):
        status = runParamTune(name)
        self.assertEqual(status, 0)
        result = checkBackupData()
        self.assertEqual(result, 1)

        if backup:
            cmd = 'echo y | keentune param dump -j {} -o {}.conf'.format(name, name)
            self.status, self.out, _ = sysCommand(cmd)
            self.assertEqual(self.status, 0)
            self.assertTrue(self.out.__contains__('dump successfully'))

    def profile_set_data(self, name):
        file_path = "conf/{}.conf".format(name)
        profile_path = "/var/keentune/profile"
        cmd = 'cp {} {}'.format(file_path, profile_path)
        self.status, _, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)

        cmd = "keentune profile set --group1 {}.conf".format(name)
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('Succeeded'))

        self.status = checkProfileData(name)
        self.assertEqual(self.status, 0)
        self.result = checkBackupData()
        self.assertEqual(self.result, 1)

    def test_param_tune_RBT_rollback(self):
        self.param_tune_test("param1")
        self.param_tune_test("param2")
        self.rollback_data("param")
        result = checkBackupData()
        self.assertEqual(result, 0)
        self.rollback_data("param", self.msg)
        result = checkBackupData()
        self.assertEqual(result, 0)

    def test_profile_set_RBT_rollback(self):
        self.profile_set_data("profile1")
        self.profile_set_data("profile2")
        self.rollback_data("profile")
        result = checkBackupData()
        self.assertEqual(result, 0)
        self.rollback_data("profile", self.msg)
        result = checkBackupData()
        self.assertEqual(result, 0)

    def test_param_profile_RBT_rollback_01(self):
        self.param_tune_test("param1")
        self.profile_set_data("profile1")

        self.rollback_data("profile")
        self.status = checkBackupData()
        self.assertEqual(self.status, 0)
        self.rollback_data("param", self.msg)
        self.status = checkBackupData()
        self.assertEqual(self.status, 0)
        
    def test_param_profile_RBT_rollback_02(self):
        self.profile_set_data("profile1")
        self.param_tune_test("param1")

        self.rollback_data("param")
        self.status = checkBackupData()
        self.assertEqual(self.status, 0)
        self.rollback_data("profile", self.msg)
        self.status = checkBackupData()
        self.assertEqual(self.status, 0)
