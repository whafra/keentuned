import os
import sys
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from UI_reliablity.test_experts_tuning_abnormal import TestKeenTuneUiAbnormal
from UI_reliablity.test_smart_tuning_abnormal import TestKeenTuneUiSmartAbnormal
from UI_reliablity.test_sensitive_tuning_abnormal import TestKeenTuneUiSensitiveAbnormal


def RunReliabilityCase():
    profile_suite = unittest.TestSuite()
    profile_suite.addTest(TestKeenTuneUiAbnormal("test_group_empty"))
    profile_suite.addTest(TestKeenTuneUiAbnormal("test_copyfile_name_exsit"))
    profile_suite.addTest(TestKeenTuneUiAbnormal("test_creatfile_name_exsit"))
    profile_suite.addTest(TestKeenTuneUiAbnormal("test_editorfile_exsit_name"))
    profile_suite.addTest(TestKeenTuneUiAbnormal("test_copyfile_name_empty"))
    profile_suite.addTest(TestKeenTuneUiAbnormal("test_copyfile_context_empty"))
    profile_suite.addTest(TestKeenTuneUiAbnormal("test_copyfile_context_error"))
    profile_suite.addTest(TestKeenTuneUiAbnormal("test_creatfile_name_empty"))
    profile_suite.addTest(TestKeenTuneUiAbnormal("test_creatfile_content_empty"))
    profile_suite.addTest(TestKeenTuneUiAbnormal("test_creatfile_content_error"))
    profile_suite.addTest(TestKeenTuneUiAbnormal("test_editorfile_delete_name"))
    profile_suite.addTest(TestKeenTuneUiAbnormal("test_editorfile_delete_content"))
    profile_suite.addTest(TestKeenTuneUiAbnormal("test_editorfile_error_content"))

    param_suite = unittest.TestSuite()
    param_suite.addTest(TestKeenTuneUiSmartAbnormal("test_required_para_empty"))
    param_suite.addTest(TestKeenTuneUiSmartAbnormal("test_create_name_exsit"))
    param_suite.addTest(TestKeenTuneUiSmartAbnormal("test_Abnormal_input"))
    param_suite.addTest(TestKeenTuneUiSmartAbnormal("test_rerun_name_exsit"))
    param_suite.addTest(TestKeenTuneUiSmartAbnormal("test_rerun_Abnormal_input"))

    sensi_suite = unittest.TestSuite()
    sensi_suite.addTest(TestKeenTuneUiSensitiveAbnormal("test_required_para_empty"))
    sensi_suite.addTest(TestKeenTuneUiSensitiveAbnormal("test_create_name_exsit"))
    sensi_suite.addTest(TestKeenTuneUiSensitiveAbnormal("test_Abnormal_input"))
    sensi_suite.addTest(TestKeenTuneUiSensitiveAbnormal("test_rerun_required_para_empty"))
    sensi_suite.addTest(TestKeenTuneUiSensitiveAbnormal("test_rerun_name_exsit"))
    sensi_suite.addTest(TestKeenTuneUiSensitiveAbnormal("test_rerun_Abnormal_input"))

    suite = unittest.TestSuite([profile_suite, param_suite, sensi_suite])
    return suite


if __name__ == '__main__':
    if sys.argv.__len__() <= 1:
        print("'web_ip' is wanted: python3 main.py 127.0.0.1")
        exit(1)
    TestKeenTuneUiAbnormal.web_ip = sys.argv[1]
    TestKeenTuneUiSmartAbnormal.web_ip = sys.argv[1]
    TestKeenTuneUiSensitiveAbnormal.web_ip = sys.argv[1]
    runner = unittest.TextTestRunner(verbosity=2)
    runner.run(RunReliabilityCase())
