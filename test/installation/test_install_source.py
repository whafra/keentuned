import os
import sys
import time
import unittest
import logging

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import runParamTune
from common import checkServerStatus
from common import deleteDependentData
from installation.common_model import sysCommand
from installation.common_model import serverPrepare
from installation.common_model import clearKeentuneEnv

logger = logging.getLogger(__name__)


class TestInstallSource(unittest.TestCase):
    def setUp(self) -> None:
        clearKeentuneEnv()
        status = serverPrepare("source")
        self.assertTrue(status)
        logger.info("start to run test_install_source testcase")
        self.keentune = ["keentuned", "keentune-target", "keentune-brain", "keentune-bench"]

    def tearDown(self) -> None:
        clearKeentuneEnv()
        logger.info("the test_sensitize_train testcase finished")

    def test_install_source_FUN(self):
        self.assertTrue(sysCommand("go env -w GO111MODULE=on;go env -w GOPROXY=https://goproxy.cn,direct")[0])
        self.assertTrue(sysCommand("cd ../../keentuned/;sh ./keentuned_install.sh")[0])
        self.assertTrue(sysCommand("cd ../../keentune-target/;python3 setup.py install")[0])
        self.assertTrue(sysCommand("cd ../../keentune-bench/;python3 setup.py install")[0])
        self.assertTrue(sysCommand("cd ../../keentune-brain/;python3 setup.py install")[0])

        for app_name in self.keentune:
            self.assertTrue(sysCommand("nohup {} > /dev/null 2>&1 &".format(app_name))[0])
            time.sleep(5)

        status = checkServerStatus(self.keentune)
        self.assertEqual(status, 0)
        status = runParamTune("param1")
        self.assertEqual(status, 0)
        deleteDependentData("param1")
