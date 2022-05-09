import os
import sys
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))
os.chdir(os.path.abspath(os.path.join(os.getcwd(), "..")))

from CLI_reliability.test_param_tune import TestParamTune
from CLI_reliability.test_param_dump import TestParamDump
from CLI_reliability.test_param_delete import TestParamDelete
from CLI_reliability.test_profile_info import TestProfileInfo
from CLI_reliability.test_profile_set import TestProfileSet
from CLI_reliability.test_profile_delete import TestProfileDelete
from CLI_reliability.test_profile_generate import TestProfileGenerate
from CLI_reliability.test_sensitize_collect import TestSensitizeCollect
from CLI_reliability.test_multi_scenes import TestMultiScenes
from CLI_reliability.test_multi_target import TestMultiTarget


def RunReliabilityCase():
    param_suite = unittest.TestSuite()
    param_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestParamTune))
    param_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestParamDump))
    param_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestParamDelete))

    profile_suite = unittest.TestSuite()
    profile_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestProfileInfo))
    profile_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestProfileSet))
    profile_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestProfileDelete))
    profile_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestProfileGenerate))

    sensitize_suite = unittest.TestSuite()
    sensitize_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestSensitizeCollect))

    multi_scenes = unittest.TestSuite()
    multi_scenes.addTests(unittest.TestLoader().loadTestsFromTestCase(TestMultiScenes))

    multi_target = unittest.TestSuite()
    multi_target.addTests(unittest.TestLoader().loadTestsFromTestCase(TestMultiTarget))

    suite = unittest.TestSuite([param_suite, profile_suite, sensitize_suite, multi_scenes, multi_target])
    return suite


if __name__ == '__main__':
    runner = unittest.TextTestRunner(verbosity=2)
    runner.run(RunReliabilityCase())