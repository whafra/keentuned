import unittest
import requests
import json
from common import keentuned_ip, keentuned_port

class TestKeentunedSensitizeResult(unittest.TestCase):
    def setUp(self) -> None:
        self.proxies={"http": None, "https": None}
        url = "http://{}:{}/status".format(keentuned_ip, keentuned_port)
        re = requests.get(url, proxies=self.proxies)
        if re.status_code != 200:
            print("ERROR: Can't reach KeenTuned.")
            exit()

    def tearDown(self) -> None:
        pass

    def test_keentuned_server_FUN_sensitize_result(self):
        url = "http://{}:{}/{}".format(keentuned_ip, keentuned_port, "sensitize_result")
        data_base = {
                    "suc": True,
                    "result": [
                        {
                            "domain": "sysctl",
                            "dtype": "string",
                            "name": "net.ipv4.tcp_moderate_rcvbuf",
                            "options": [
                                "0",
                                "1"
                            ],
                            "weight": 0.04892173984048826
                        },
                        {
                            "domain": "sysctl",
                            "dtype": "string",
                            "name": "kernel.sched_autogroup_enabled",
                            "options": [
                                "0",
                                "1"
                            ],
                            "weight": 0.04300587239183498
                        },
                        {
                            "domain": "sysctl",
                            "dtype": "string",
                            "name": "net.ipv4.conf.all.promote_secondaries",
                            "options": [
                                "0",
                                "1"
                            ],
                            "weight": 0.02991800689640738
                        }
                    ],
                    "msg": ""
                }

        headers = {"Content-Type": "application/json"}
        
        result = requests.post(url, data=json.dumps(data_base), headers=headers, proxies=self.proxies)
        self.assertEqual(result.status_code, 200)
        self.assertEqual(result.text, '{"suc": true, "msg": ""}')
