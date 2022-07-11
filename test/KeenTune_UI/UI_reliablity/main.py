import unittest
from test_experts_tuning_abnormal import TestKeenTune_UI_abnormal


def RunKeenTuneCase():
    suite = unittest.TestSuite()
    suite.addTest(TestKeenTune_UI_abnormal("test_group_empty"))
    suite.addTest(TestKeenTune_UI_abnormal("test_copyfile_name_exit"))
    suite.addTest(TestKeenTune_UI_abnormal("test_creatfile_name_exit"))
    suite.addTest(TestKeenTune_UI_abnormal("test_editorfile_exit_name"))
    suite.addTest(TestKeenTune_UI_abnormal("test_copyfile_name_empty"))
    suite.addTest(TestKeenTune_UI_abnormal("test_copyfile_context_empty"))
    suite.addTest(TestKeenTune_UI_abnormal("test_copyfile_context_error"))
    suite.addTest(TestKeenTune_UI_abnormal("test_creatfile_name_empty"))
    suite.addTest(TestKeenTune_UI_abnormal("test_creatfile_content_empty"))
    suite.addTest(TestKeenTune_UI_abnormal("test_creatfile_content_error"))
    suite.addTest(TestKeenTune_UI_abnormal("test_editorfile_delete_name"))
    suite.addTest(TestKeenTune_UI_abnormal("test_editorfile_delete_content"))
    suite.addTest(TestKeenTune_UI_abnormal("test_editorfile_error_content"))
    return suite


if __name__ == '__main__':
    runner = unittest.TextTestRunner(verbosity=2)
    runner.run(RunKeenTuneCase())



