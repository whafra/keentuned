import os
import sys
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from UI_reliablity.test_experts_tuning_abnormal import TestKeenTuneUiAbnormal


def RunReliabilityCase():
    suite = unittest.TestSuite()
    suite.addTest(TestKeenTuneUiAbnormal("test_group_empty"))
    suite.addTest(TestKeenTuneUiAbnormal("test_copyfile_name_exsit"))
    suite.addTest(TestKeenTuneUiAbnormal("test_creatfile_name_exsit"))
    suite.addTest(TestKeenTuneUiAbnormal("test_editorfile_exsit_name"))
    suite.addTest(TestKeenTuneUiAbnormal("test_copyfile_name_empty"))
    suite.addTest(TestKeenTuneUiAbnormal("test_copyfile_context_empty"))
    suite.addTest(TestKeenTuneUiAbnormal("test_copyfile_context_error"))
    suite.addTest(TestKeenTuneUiAbnormal("test_creatfile_name_empty"))
    suite.addTest(TestKeenTuneUiAbnormal("test_creatfile_content_empty"))
    suite.addTest(TestKeenTuneUiAbnormal("test_creatfile_content_error"))
    suite.addTest(TestKeenTuneUiAbnormal("test_editorfile_delete_name"))
    suite.addTest(TestKeenTuneUiAbnormal("test_editorfile_delete_content"))
    suite.addTest(TestKeenTuneUiAbnormal("test_editorfile_error_content"))
    return suite


if __name__ == '__main__':
    if sys.argv.__len__() <= 1:
        print("'web_ip' is wanted: python3 main.py 127.0.0.1")
        exit(1)
    TestKeenTuneUiAbnormal.web_ip = sys.argv[1]
    runner = unittest.TextTestRunner(verbosity=2)
    runner.run(RunReliabilityCase())
