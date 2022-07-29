import os
import sys
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from UI_base.test_experts_tuning_normal import TestKeenTuneUiNormal


def RunBasicCase():
    suite = unittest.TestSuite()
    suite.addTest(TestKeenTuneUiNormal("test_copyfile"))
    suite.addTest(TestKeenTuneUiNormal("test_creatfile"))
    suite.addTest(TestKeenTuneUiNormal("test_checkfile"))
    suite.addTest(TestKeenTuneUiNormal("test_editor"))
    suite.addTest(TestKeenTuneUiNormal("test_set_group"))
    suite.addTest(TestKeenTuneUiNormal("test_restore"))
    suite.addTest(TestKeenTuneUiNormal("test_deletefile"))
    suite.addTest(TestKeenTuneUiNormal("test_language_switch"))
    suite.addTest(TestKeenTuneUiNormal("test_refresh"))
    suite.addTest(TestKeenTuneUiNormal("test_set_list"))
    return suite


if __name__ == '__main__':
    if sys.argv.__len__() <= 1:
        print("'web_ip' is wanted: python3 main.py 127.0.0.1")
        exit(1)
    TestKeenTuneUiNormal.web_ip = sys.argv[1]
    runner = unittest.TextTestRunner(verbosity=2)
    runner.run(RunBasicCase())
