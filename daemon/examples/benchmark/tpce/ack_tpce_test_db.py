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
        cmd = 'tpce-run-workload-pg -a pg -n pg -c 5000 -t 5000 -d 120 -u 30 -f 500 -w 300 -e 30000 -g 30010 -z 1 -p 0 -v 30 -i /home/lilinjie/tpc-ebenchmark/egen -o /home/lilinjie/tpc-ebenchmark/tmp/results'
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
            
            pattern_tpse = re.compile(r'tpse=\s+([\d.]+)(\w+)')
            pattern_ramp_up = re.compile(r'ramp_up=\s+([\d.]+)(\w+)')
           # pattern_measurement = re.compile(r'measurement=\s+([\d.]+)(\w+)')
            pattern_trans_total = re.compile(r'trans_total=\s+([\d.]+)(\w+)')
            
            if not re.search(pattern_tpse,self.out) \
                or not re.search(pattern_ramp_up,self.out) \
                or not re.search(pattern_trans_total,self.out):
                logger.error("can not parse output: {}".format(self.out))
                return False, []
            
            tpse = float(re.search(pattern_tpse,self.out).group(1))

            _ramp_up = float(re.search(pattern_ramp_up,self.out).group(1))
            ramp_up_measurement = re.search(pattern_ramp_up,self.out).group(2)
            ramp_up = self.__transfMeasurement(_ramp_up, ramp_up_measurement)
            
           # _measurement = float(re.search(pattern_measurement,self.out).group(1))
           # measurement_measurement = re.search(pattern_measurement,self.out).group(2)
           # measurement = self.__transfMeasurement(_measurement, measurement_measurement)
            
            trans_total = float(re.search(pattern_trans_total,self.out).group(1))

            result = {
                "Tpse": tpse,
                "Ramp_up": ramp_up,
               # "Measurement": measurement,
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
