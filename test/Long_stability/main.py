import os
import sys
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import deleteDependentData
from Long_stability.test_long_stability import TestLongStability


def RunLongStableCase():
    suite = unittest.TestSuite()
    suite.addTest(TestLongStability("test_long_stability_RBT"))
    return suite


if __name__ == '__main__':
    if sys.argv.__len__() <= 1:
        print("'time limit' is wanted, unit is hour: python3 main.py 24")
        exit(1)
    TestLongStability.time_limit = int(sys.argv[1])
    print("--------------- start to run long stable test cases ---------------")
    deleteDependentData("param1")
    deleteDependentData("sensitize1")
    runner = unittest.TextTestRunner(verbosity=2)
    runner.run(RunLongStableCase())
    print("--------------- run long stable test cases end ---------------")
