#!/usr/bin/python3
#-*- coding: utf-8 -*-
import re
import sys
import subprocess
import logging
import time
logger = logging.getLogger(__name__)

"""
iperf benchmark
e.g.
'iperf3 -c 127.0.0.1 -u -t 10 -P 1 -b 1M  -M 100 -w 10240 -l 8192 -Z'
"""

# const
TIME = 15
PARALLEL = 10
COMMAND = "-t {} -i 3".format(TIME)
DEFAULT = "-P 10 -w 10240 -l 131072"


class Benchmark():
    def __init__(self, url, default=DEFAULT, time=TIME, command=COMMAND, parallel=PARALLEL):
        """Init benchmark

        Args:
            url (string): url
            connections (int, optional): Connections to keep open. Defaults to 300.
            threads (int, optional):  Number of threads to use. Defaults to 10.
            duration (int, optional): Duration of test. Defaults to 30.
        """
        self.url = url
        self.time = time
        self.parallel = parallel
        self.command = ' '.join((command, default))
        self.out = ""
        self.error = ""
    
    def __transfMeasurement(self, value, measurement):
        if measurement in ['B', 'b']:
            return value
        elif measurement in ['KB', 'Kb']:
            return value * 10 ** 3
        elif measurement in ['MB', 'Mb']:
            return value * 10 ** 6
        elif measurement in ['GB', 'Gb']:
            return value * 10 ** 9
        
        else:
            logger.warning("Unknown measurement: %s", measurement)
            return value

    def run(self):
        """Run benchmark and parse output
        
        Return True and score list if running benchmark successfully, otherwise return False and empty list.
        """
        cmd = 'iperf3 -c {} {}'.format(self.url, self.command)
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

            label = r"\s+\d" if int(self.parallel) == 1 else "SUM"
            sender_transfer = re.compile(r'\[{}\]\s+0\.00-{}\..+sec\s+([\d.]+)\s+([A-Z]+)ytes.+sender'.format(label, self.time))
            sender_bandwidth = re.compile(r'\[{}\]\s+0\.00-{}\..+Bytes\s+([\d.]+)\s+([a-zA-Z]+)its/sec.+sender'.format(label, self.time))
            receiver_transfer = re.compile(r'\[{}\]\s+0\.00-{}\..+sec\s+([\d.]+)\s+([A-Z]+)ytes.+receiver'.format(label, self.time))
            receiver_bandwidth = re.compile(r'\[{}\]\s+0\.00-{}\..+Bytes\s+([\d.]+)\s+([a-zA-Z]+)its/sec.+receiver'.format(label, self.time))
            
            if not re.search(sender_transfer, self.out) \
                or not re.search(sender_bandwidth, self.out) \
                or not re.search(receiver_transfer, self.out) \
                or not re.search(receiver_bandwidth, self.out):
                logger.error("can not parse output: %s", self.out)
                return False, []
            
            _transfer = float(re.search(sender_transfer, self.out).group(1))
            transfer_measurement = re.search(sender_transfer, self.out).group(2)
            sender_transfer = self.__transfMeasurement(_transfer, transfer_measurement)

            _bandwidth = float(re.search(sender_bandwidth, self.out).group(1))
            bandwidth_measurement = re.search(sender_bandwidth, self.out).group(2)
            sender_bandwidth = self.__transfMeasurement(_bandwidth, bandwidth_measurement)

            _transfer = float(re.search(receiver_transfer, self.out).group(1))
            transfer_measurement = re.search(receiver_transfer, self.out).group(2)
            receiver_transfer = self.__transfMeasurement(_transfer, transfer_measurement)

            _bandwidth = float(re.search(receiver_bandwidth, self.out).group(1))
            bandwidth_measurement = re.search(receiver_bandwidth, self.out).group(2)
            receiver_bandwidth = self.__transfMeasurement(_bandwidth, bandwidth_measurement)
            
            result = {
                "Interval": self.time,
                "Sender_Transfer": sender_transfer,
                "Receiver_Transfer": receiver_transfer,
                "Sender_Bandwidth": sender_bandwidth,
                "Receiver_Bandwidth": receiver_bandwidth,
            }
            
            result_str = ", ".join(["{}={}".format(k,v) for k,v in result.items()])
            time.sleep(60)
            print(result_str)
            return True, result_str

        else:
            logger.error(self.error)
            return False, []


if __name__ == "__main__":
    if sys.argv.__len__() <= 1:
        print("'Target ip' is wanted: python3 iperf.py [Target ip]")
        exit(1)
    bench = Benchmark(sys.argv[1])
    suc, res = bench.run()
