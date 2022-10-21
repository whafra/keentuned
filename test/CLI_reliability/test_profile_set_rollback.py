import os
import re
import sys
import logging
import time
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import checkServerStatus
from common import sysCommand
from common import getSysBackupData
from common import checkBackupData
from common import checkProfileData

logger = logging.getLogger(__name__)


class TestProfileSetRollback(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_profile_set_rollback testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)

        logger.info('the test_profile_set_rollback testcase finished')

    def rollback_profile_status(self):
        cmd = 'keentune profile list'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertFalse(self.out.__contains__('[active]'))

    def check_profile_status(self, name):
        cmd = 'keentune profile list'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.result = re.search(r'\[(.*?)\].+{}.conf'.format(name), self.out).group(1)
        self.assertTrue(self.result.__contains__('active'))

    def set_conf_data(self, name):
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
        self.check_profile_status(name)

    def test_profile_rollback_RBT_set(self):
        getSysBackupData()
        self.set_conf_data("profile1")
        self.set_conf_data("profile2")
        
        cmd = 'keentune profile rollback'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('profile rollback successfully'))
        res = checkBackupData()
        self.assertEqual(res, 0)
        self.rollback_profile_status()

        os.remove("/var/keentune/profile/profile1.conf")
        os.remove("/var/keentune/profile/profile2.conf")