import os
import sys
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from MT_restful.test_keentuned_apply_result import TestKeentunedApplyResult
from MT_restful.test_keentuned_benchmark_result import TestKeentunedBenchmarkResult
from MT_restful.test_keentuned_sensitize_result import TestKeentunedSensitizeResult


def RunModelCase():
    suite = unittest.TestSuite()
    suite.addTest(TestKeentunedApplyResult('test_keentuned_server_FUN_apply_result'))
    suite.addTest(TestKeentunedBenchmarkResult('test_keentuned_server_FUN_benchmark_result'))
    suite.addTest(TestKeentunedSensitizeResult('test_keentuned_server_FUN_sensitize_result'))
    return suite


if __name__ == '__main__':
    runner = unittest.TextTestRunner(verbosity=2)
    runner.run(RunModelCase())
