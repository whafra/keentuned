import os
import sys
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))
os.chdir(os.path.abspath(os.path.join(os.getcwd(), "..")))

from CLI_reliability.test_multi_target import TestMultiTarget


def RunReliabilityCase():
    multi_target = unittest.TestSuite()
    multi_target.addTests(unittest.TestLoader().loadTestsFromTestCase(TestMultiTarget))

    suite = unittest.TestSuite([multi_target])
    return suite


if __name__ == '__main__':
    runner = unittest.TextTestRunner(verbosity=2)
    runner.run(RunReliabilityCase())