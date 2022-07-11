# Configuration of KeenTune
---  
## KeenTuned
```conf
[keentuned]
KEENTUNED_HOME = /etc/keentune
HEARTBEAT_TIME = 30
PORT           = 9871

[brain]
BRAIN_IP = localhost
BRAIN_PORT = 9872
ALGORITHM = tpe

[target-group-1]
TARGET_IP = localhost
TARGET_PORT = 9873
PARAMETER = sysctl.json

[bench-group-1]
BENCH_SRC_IP = localhost
BENCH_DEST_IP = localhost
BENCH_SRC_PORT = 9874
BENCH_DEST_PORT = 9875
BASELINE_BENCH_ROUND = 5
TUNING_BENCH_ROUND = 3
RECHECK_BENCH_ROUND = 4
BENCH_CONFIG = bench_wrk_nginx_long.json

[dump]
DUMP_BASELINE_CONFIGURATION = false
DUMP_TUNING_CONFIGURATION = false
DUMP_BEST_CONFIGURATION = true
DUMP_HOME = /var/keentune

[sensitize]
ALGORITHM = random
BENCH_ROUND = 3

[log]
LOGFILE_LEVEL  = DEBUG
LOGFILE_NAME   = keentuned.log
LOGFILE_INTERVAL = 2
LOGFILE_BACKUP_COUNT = 14

[version]
VERSION_NUM = 1.1.0
```
## KeenTune-brain
## KeenTune-target
## KeenTune-bench