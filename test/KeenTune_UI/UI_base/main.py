import unittest
from test_experts_tuning_normal import TestKeenTune_UI_normal


def RunKeenTuneCase():
    suite = unittest.TestSuite()
    suite.addTest(TestKeenTune_UI_normal("test_copyfile"))
    suite.addTest(TestKeenTune_UI_normal("test_creatfile"))
    suite.addTest(TestKeenTune_UI_normal("test_checkfile"))
    suite.addTest(TestKeenTune_UI_normal("test_editor"))
    suite.addTest(TestKeenTune_UI_normal("test_set_group"))
    suite.addTest(TestKeenTune_UI_normal("test_restore"))
    suite.addTest(TestKeenTune_UI_normal("test_deletefile"))
    suite.addTest(TestKeenTune_UI_normal("test_language_switch"))
    suite.addTest(TestKeenTune_UI_normal("test_refresh"))
    suite.addTest(TestKeenTune_UI_normal("test_set_list"))
    return suite


if __name__ == '__main__':
    runner = unittest.TextTestRunner(verbosity=2)
    runner.run(RunKeenTuneCase())
