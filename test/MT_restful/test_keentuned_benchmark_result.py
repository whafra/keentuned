import unittest
import requests
import json
from common import keentuned_ip, keentuned_port

class TestKeentunedBenchmarkResult(unittest.TestCase):
    def setUp(self) -> None:
        self.proxies={"http": None, "https": None}
        url = "http://{}:{}/status".format(keentuned_ip, keentuned_port)
        re = requests.get(url, proxies=self.proxies)
        if re.status_code != 200:
            print("ERROR: Can't reach KeenTuned.")
            exit()

    def tearDown(self) -> None:
        pass

    def test_keentuned_server_FUN_benchmark_result(self):
        url = "http://{}:{}/{}".format(keentuned_ip, keentuned_port, "benchmark_result")
        data_base = {
                    "suc": True,
                    "result": {
                        "Latency_90": {
                            "value": 63280.0
                        },
                        "Latency_99": {
                            "value": 135010.0
                        },
                        "Requests_sec": {
                            "value": 22819.67
                        },
                        "Transfer_sec": {
                            "value": 92410000.0
                        }
                    },
                    "msg": ""
                }

        headers = {"Content-Type": "application/json"}
        
        result = requests.post(url, data=json.dumps(data_base), headers=headers, proxies=self.proxies)
        self.assertEqual(result.status_code, 200)
        self.assertEqual(result.text, '{"suc": true, "msg": ""}')
