#!/usr/bin/python3
#-*- coding: utf-8 -*-
import re
import sys
import subprocess
import logging
logger = logging.getLogger(__name__)


class Benchmark():
    def __transfMeasurement(self,value,measurement):
        if measurement == '':
            return value

        # measurement of TPCH_time/Total_time
        elif measurement == 'h':
            return value * 60 * 60 * 10 ** 6
        elif measurement == 'm':
            return value * 60 * 10 ** 6
        elif measurement == "s":
            return value * 10 ** 6
        elif measurement == "ms":
            return value * 10 ** 3
        elif measurement == 'us':
            return value

        else:
            logger.warning("Unknown measurement: {}".format(measurement))
            return value

    def run(self):
        """Run benchmark and parse output

        Return True and score list if running benchmark successfully, otherwise return False and empty list.
        """
        cmd = 'cd /root/tpch/; sh tpch.sh'
        logger.info(cmd)
        result = subprocess.run(
                    cmd,
                    shell=True,
                    close_fds=True,
                    stderr=subprocess.PIPE,
                    stdout=subprocess.PIPE
                )
        self.out = result.stdout.decode('UTF-8','strict')
        self.error = result.stderr.decode('UTF-8','strict')
        if result.returncode == 0:
            logger.info(self.out)

            pattern_TPCH_time = re.compile(r'TPCH test time:\s+([\d.]+)(\w+)')
            pattern_Total_time = re.compile(r'total time:\s+([\d.]+)(\w+)')

            if not re.search(pattern_TPCH_time,self.out) \
                or not re.search(pattern_Total_time,self.out):
                logger.error("can not parse output: {}".format(self.out))
                return False, []

            _TPCH_time = float(re.search(pattern_TPCH_time,self.out).group(1))

            _Total_time = float(re.search(pattern_Total_time,self.out).group(1))

            result = {
                "TPCH_Time": _TPCH_time,
                "Total_Time": _Total_time,
            }

            result_str = ", ".join(["{} = {}".format(k,v) for k,v in result.items()])
            print(result_str)
            return True, result_str
        else:
            logger.error(self.error)
            return False, []
if __name__ == "__main__":
    bench = Benchmark()
    suc, result = bench.run()
