import os
import sys
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from CLI_basic.test_help import TestHelp
from test_version import TestVersion
from CLI_basic.test_param_tune import TestParamTune
from CLI_basic.test_param_list import TestParamList
from CLI_basic.test_param_dump import TestParamDump
from CLI_basic.test_param_rollback import TestParamRollback
from CLI_basic.test_param_delete import TestParamDelete
from CLI_basic.test_profile_delete import TestProfileDelete
from CLI_basic.test_profile_generate import TestProfileGenerate
from CLI_basic.test_profile_info import TestProfileInfo
from CLI_basic.test_profile_list import TestProfileList
from CLI_basic.test_profile_rollback import TestProfileRollback
from CLI_basic.test_profile_set import TestProfileSet
from CLI_basic.test_sensitize_collect import TestSensitizeCollect
from test_sensitize_train import TestSensitizeTrain
from CLI_basic.test_sensitize_list import TestSensitizeList
from CLI_basic.test_sensitize_delete import TestSensitizeDelete
from test_multi_target import TestMultiTarget
from common import deleteDependentData


def RunBasicCase():
    param_suite = unittest.TestSuite()
    param_suite.addTest(TestHelp('test_help_FUN'))
    param_suite.addTest(TestVersion('test_version'))
    param_suite.addTest(TestParamTune('test_param_tune_FUN'))
    param_suite.addTest(TestParamList('test_param_list_FUN'))
    param_suite.addTest(TestParamDump('test_param_dump_FUN'))
    param_suite.addTest(TestParamRollback('test_param_rollback_FUN'))
    param_suite.addTest(TestParamDelete('test_param_delete_FUN'))

    profile_suite = unittest.TestSuite()
    profile_suite.addTest(TestProfileList('test_profile_list_FUN'))
    profile_suite.addTest(TestProfileInfo('test_profile_info_FUN'))
    profile_suite.addTest(TestProfileSet('test_profile_set_FUN'))
    profile_suite.addTest(TestProfileGenerate('test_profile_generate_FUN'))
    profile_suite.addTest(TestProfileRollback('test_profile_rollback_FUN'))
    profile_suite.addTest(TestProfileDelete('test_profile_delete_FUN'))

    sensitize_suite = unittest.TestSuite()
    sensitize_suite.addTest(TestSensitizeCollect('test_sensitize_collect_FUN'))
    sensitize_suite.addTest(TestSensitizeTrain('test_sensitize_train'))
    sensitize_suite.addTest(TestSensitizeList('test_sensitize_list_FUN'))
    sensitize_suite.addTest(TestSensitizeDelete('test_sensitize_delete_FUN'))

    target_suite = unittest.TestSuite()
    target_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestMultiTarget))

    suite = unittest.TestSuite([param_suite, profile_suite, sensitize_suite, target_suite])
    return suite


if __name__ == '__main__':
    deleteDependentData("test1")
    deleteDependentData("test01")
    runner = unittest.TextTestRunner(verbosity=2)
    runner.run(RunBasicCase())
