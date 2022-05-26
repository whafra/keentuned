import unittest
import requests
import json
from common import keentuned_ip, keentuned_port

class TestKeentunedApplyResult(unittest.TestCase):
    def setUp(self) -> None:
        self.proxies={"http": None, "https": None}
        url = "http://{}:{}/status".format(keentuned_ip, keentuned_port)
        re = requests.get(url, proxies=self.proxies)
        if re.status_code != 200:
            print("ERROR: Can't reach KeenTuned.")
            exit()

    def tearDown(self) -> None:
        pass

    def test_keentuned_server_FUN_apply_result(self):
        url = "http://{}:{}/{}".format(keentuned_ip, keentuned_port, "apply_result")
        data_base = {
                    "suc": True,
                    "data": {
                        "sysctl": {
                            "fs.aio-max-nr": {
                                "value": 819200,
                                "dtype": "int",
                                "suc": True,
                                "msg": ""
                            },
                            "fs.file-max": {
                                "value": 890880,
                                "dtype": "int",
                                "suc": True,
                                "msg": ""
                            }
                        }
                    },
                    "msg": ""
                }

        headers = {"Content-Type": "application/json"}
        
        result = requests.post(url, data=json.dumps(data_base), headers=headers, proxies=self.proxies)
        self.assertEqual(result.status_code, 200)
        self.assertEqual(result.text, '{"suc": true, "msg": ""}')
