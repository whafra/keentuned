import os
import sys
import unittest

from UI_base.main import RunBasicCase
from UI_base.main import TestKeenTuneUiNormal
from UI_base.main import TestKeenTuneUiSmartNormal
from UI_base.main import TestKeenTuneUiSensitiveNormal
from UI_reliablity.main import RunReliabilityCase
from UI_reliablity.main import TestKeenTuneUiAbnormal
from UI_reliablity.main import TestKeenTuneUiSmartAbnormal
from UI_reliablity.main import TestKeenTuneUiSensitiveAbnormal


def RunAllCase():
    basic_suite = RunBasicCase()
    reliability_suite = RunReliabilityCase()
    suite = unittest.TestSuite([basic_suite, reliability_suite])
    return suite


if __name__ == '__main__':
    if sys.argv.__len__() <= 1:
        print("'web_ip' is wanted: python3 main.py 127.0.0.1")
        exit(1)
    TestKeenTuneUiNormal.web_ip = sys.argv[1]
    TestKeenTuneUiAbnormal.web_ip = sys.argv[1]
    TestKeenTuneUiSmartNormal.web_ip = sys.argv[1]
    TestKeenTuneUiSmartAbnormal.web_ip = sys.argv[1]
    TestKeenTuneUiSensitiveNormal.web_ip = sys.argv[1]
    TestKeenTuneUiSensitiveAbnormal.web_ip = sys.argv[1]
    print("--------------- start to run test cases ---------------")
    runner = unittest.TextTestRunner(verbosity=2)
    runner.run(RunAllCase())
    print("--------------- run test cases end ---------------")
