import os
import sys
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from test_sensitize_delete import TestSensitizeDelete
from test_sensitize_list import TestSensitizeList
from test_sensitize_train import TestSensitizeTrain
from test_sensitize_collect import TestSensitizeCollect
from test_profile_set import TestProfileSet
from test_profile_rollback import TestProfileRollback
from test_profile_list import TestProfileList
from test_profile_info import TestProfileInfo
from test_profile_generate import TestProfileGenerate
from test_profile_delete import TestProfileDelete
from test_param_delete import TestParamDelete
from test_param_rollback import TestParamRollback
from test_param_dump import TestParamDump
from test_param_list import TestParamList
from test_param_tune import TestParamTune
from test_help import TestHelp


def RunBasicCase():
    param_suite = unittest.TestSuite()
    param_suite.addTest(TestParamTune('test_param_tune'))
    param_suite.addTest(TestParamList('test_param_list'))
    param_suite.addTest(TestParamDump('test_param_dump'))
    param_suite.addTest(TestParamRollback('test_param_rollback'))
    param_suite.addTest(TestParamDelete('test_param_delete'))

    profile_suite = unittest.TestSuite()
    profile_suite.addTest(TestProfileList('test_profile_list'))
    profile_suite.addTest(TestProfileInfo('test_profile_info'))
    profile_suite.addTest(TestProfileSet('test_profile_set'))
    profile_suite.addTest(TestProfileGenerate('test_profile_generate'))
    profile_suite.addTest(TestProfileRollback('test_profile_rollback'))
    profile_suite.addTest(TestProfileDelete('test_profile_delete'))

    sensitize_suite = unittest.TestSuite()
    sensitize_suite.addTest(TestSensitizeCollect('test_sensitize_collect'))
    sensitize_suite.addTest(TestSensitizeTrain('test_sensitize_train'))
    sensitize_suite.addTest(TestSensitizeList('test_sensitize_list'))
    sensitize_suite.addTest(TestSensitizeDelete('test_sensitize_delete'))

    suite = unittest.TestSuite([param_suite, profile_suite, sensitize_suite])
    return suite


if __name__ == '__main__':
    runner = unittest.TextTestRunner(verbosity=2)
    runner.run(RunBasicCase())
