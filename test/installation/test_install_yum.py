import os
import sys
import time
import unittest
import logging

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import sysCommand
from common import runParamTune
from common import checkServerStatus
from common import deleteDependentData
from installation.common_model import getOsType
from installation.common_model import serverPrepare
from installation.common_model import installPackage
from installation.common_model import clearKeentuneEnv

logger = logging.getLogger(__name__)


class TestInstallYum(unittest.TestCase):
    @classmethod
    def setUpClass(self) -> None:
        clearKeentuneEnv()
        logger.info("start to run test_install_yum testcase")

    @classmethod
    def tearDownClass(self) -> None:
        clearKeentuneEnv()
        logger.info("the test_install_yum testcase finished")

    def setUp(self) -> None:
        os_type = getOsType()
        self.assertEqual(os_type, "redhat")
        self.keentune = ["keentuned", "keentune-target", "keentune-brain", "keentune-bench"]

    def tearDown(self) -> None:
        logger.info("the test_install_yum testcase finished")

    def install_server(self):
        for server in self.keentune:
            cmd = "yum install -y {0};systemctl restart {0}".format(server)
            self.assertEqual(sysCommand(cmd)[0], 0)
            time.sleep(3)

    def test_install_yum_FUN(self):
        self.status = sysCommand("pip3 install --upgrade pip")[0]
        self.assertEqual(self.status, 0)
        installPackage()

        cmd = "grep -o -i keentune /etc/yum.repos.d/epel.repo"
        status = sysCommand(cmd)[0]
        if status != 0:
            baseurl = "https://mirrors.openanolis.cn/anolis/8.6/Plus/$basearch/os"
            gpgkey = "https://mirrors.openanolis.cn/anolis/RPM-GPG-KEY-ANOLIS"
            cmd = "echo -e '\n[keentune]\nname=keentune-os\nbaseurl={}\ngpgkey={}\nenabled=1\ngpgcheck=0' >> /etc/yum.repos.d/epel.repo;\
                   yum clean all;yum makecache".format(baseurl, gpgkey)
            self.assertEqual(sysCommand(cmd)[0], 0)
        
        self.install_server()

        self.status = checkServerStatus(self.keentune)
        self.assertEqual(self.status, 0)
        self.status = runParamTune("param1")
        self.assertEqual(self.status, 0)
        deleteDependentData("param1")
    
    def test_instatll_yum_RBT_start(self):
        for server in self.keentune:
            cmd = "systemctl restart {0};systemctl status {0}".format(server)
            self.status, self.output, _ = sysCommand(cmd)
            self.assertEqual(self.status, 0)
            self.assertIn("Active: active (running)", self.output)

    def test_instatll_yum_RBT_stop(self):
        for server in self.keentune:
            cmd = "systemctl stop {0};systemctl status {0}".format(server)
            self.status, self.output, _ = sysCommand(cmd)
            self.assertEqual(self.status, 3)
            self.assertIn("Active: inactive (dead)", self.output)
