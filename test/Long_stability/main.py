import os
import sys
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import deleteDependentData
from Long_stability.test_long_stability import TestLongStability


def RunLongStableCase():
    suite = unittest.TestSuite()
    suite.addTest(TestLongStability("test_long_stability"))
    return suite


if __name__ == '__main__':
    print("--------------- start to run test cases ---------------")
    deleteDependentData("param1")
    deleteDependentData("sensitize1")
    runner = unittest.TextTestRunner(verbosity=2)
    runner.run(RunLongStableCase())
    print("--------------- run test cases end ---------------")
