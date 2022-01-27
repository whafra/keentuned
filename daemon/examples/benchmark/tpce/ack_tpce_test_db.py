#!/usr/bin/python3
#-*- coding: utf-8 -*-
import re
import sys
import subprocess
import logging
logger = logging.getLogger(__name__)
"""
TPCE Pressure test DataBase
e.g.
'tpce-run-workload-pg -a pg -n pg -c 5000 -t 5000 -d 120 -u 30 -f 500 -w 300 -e 30000 -g 30010 -z 1 -p 0 -v 30 -i /opt/tpc-ebenchmark/egen -o /opt/tpc-ebenchmark/tmp/results'
"""
class Benchmark():
    def __transfMeasurement(self,value,measurement):
        if measurement == '':
            return value

        # measurement of
        elif measurement == 'h':
            return value * 60 * 60 * 10 ** 6
        elif measurement == 'minute(s)':
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
        cmd = 'tpce-run-workload-pg -a pg -n pg -c 5000 -t 5000 -d 30 -u 30 -f 500 -w 300 -e 30000 -g 30010 -z 1 -p 0 -v 30 -i /opt/tpc-ebenchmark/egen -o /opt/tpc-ebenchmark/tmp/result2218'
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
            pattern_tpse = re.compile(r'tpse=([\d.]+)')
            pattern_ramp_up = re.compile(r'ramp_up=([\d.]+)')
            pattern_trans_total = re.compile(r'trans_total=([\d.]+)')

            if not re.search(pattern_tpse,self.out) \
                or not re.search(pattern_ramp_up,self.out) \
                or not re.search(pattern_trans_total,self.out):
                logger.error("can not parse output: {}".format(self.out))
                return False, []

            tpse = float(re.search(pattern_tpse,self.out).group(1))
            ramp_up = float(re.search(pattern_ramp_up,self.out).group(1))
            trans_total = float(re.search(pattern_trans_total,self.out).group(1))
            result = {
                "Tpse": tpse,
                "Ramp_up": ramp_up,
                "Trans_total": trans_total,
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

