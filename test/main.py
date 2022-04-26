import os
import sys
import unittest

from common import deleteDependentData
from CLI_basic.main import RunBasicCase
from CLI_reliability.main import RunReliabilityCase

os.chdir(os.path.abspath(os.path.join(os.getcwd(), "test")))


def RunAllCase():
    basic_suite = RunBasicCase()
    reliability_suite = RunReliabilityCase()
    suite = unittest.TestSuite([basic_suite, reliability_suite])
    return suite


if __name__ == '__main__':
    print("--------------- start to run test cases ---------------")
    deleteDependentData("param1")
    deleteDependentData("sensitize1")
    runner = unittest.TextTestRunner(verbosity=2)
    runner.run(RunAllCase())
    print("--------------- run test cases end ---------------")
