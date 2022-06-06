# KeenTune参数敏感化
## 参数敏感化介绍
KeenTune 是一款AI算法与专家知识库双轮驱动的操作系统全栈式智能优化产品，为主流的操作系统提供轻量化、跨平台的一键式性能调优，让应用在智能定制的运行环境发挥最优性能。敏感参数识别主要辅助 KeenTune 调优，在采集数据的基础上识别出各参数对调优结果影响大小的权重值，方便我们快速识别出敏感参数，指导我们进一步优化调优参数的配置，形成调优的闭环。

## 使用手册
通过 keentune sensitize -h 命令，可查看和敏感参数识别相关的所有可用命令
```sh
>>> keentune sensitize -h
Sensitive parameter identification and explanation with AI algorithms

Usage:
  keentune sensitize [command] [flags]

Examples:
        keentune sensitize collect --data collect_test --iteration 10
        keentune sensitize delete --data collect_test
        keentune sensitize list
        keentune sensitize stop
        keentune sensitize train --data collect_test --output train_test --trials 2

Available Commands:
  collect     Collecting parameter and benchmark score as sensitivity identification data randomly
  delete      Delete the sensitivity identification data
  list        List available sensitivity identification data
  stop        Terminate a sensitivity identification job
  train       Deploy and start a sensitivity identification job

Flags:
  -h, --help   help message

Use "keentune sensitize [command] --help" for more information about a command.
```
类似于动态调优，确认好调优参数和发压配置文件后，可使用 keentune sensitize collect 命令来采集对应场景下tuning的数据，为识别敏感参数做准备。--data 参数指定采集任务名，简写为 -d ，必选参数，注意任务名尽量保持唯一性；--iteration 参数指定采集轮次，简写为 -i，可选参数，默认值为100轮，推荐值不小于10，方便后续的数据识别；--debug 参数指定为调试模式，可选参数，一般不用。
```sh
>>> keentune sensitize collect --data collect_test --iteration 10
[ok] Running Sensitize Collect Success.

        iteration: 10
        name: collect_test

        see more details by log file: "/var/log/keentune/keentuned-sensitize-collect-1652175231.log"
```
任务发起后，可以通过log查看具体任务执行情况。
```sh
>>> cat /var/log/keentune/keentuned-sensitize-collect-1652175231.log
```
数据采集完成后，可使用 keentune sensitize list 命令查看已经执行的所有任务列表，包括已完成的采集任务和动态调优任务及使用的算法等信息，以供敏感参数识别使用。
```sh
>>> keentune sensitize list
Get sensitive parameter identification results successfully, and the details displayed in the terminal.
+------------------------------------+----------------------+-----------+
|data name                           |application scenario  |algorithm  |
+------------------------------------+----------------------+-----------+
|collect_test                        |collect               |Random     |
|tune_test                           |tuning                |TPE        |
+------------------------------------+----------------------+-----------+
```
在数据采集完成后，使用 keentune sensitize train 命令来训练已采集到的数据，从而识别出各参数对调优结果影响大小的权重值，并输出到指定的文件中。此操作可以直观的看到各个调优参数对调优结果的影响大小，方便我们快速识别出敏感参数，指导我们进一步优化调优参数的配置，形成调优的闭环。--data 参数为调优任务名，简写为 -d，必选参数；--output 参数为输出文件名，简写为 -o，可选参数；--trials 参数为训练轮次，默认值为1，可选参数。
```sh
>>> keentune sensitize train --data collect_test --output train_test
[ok] Running Sensitize Train Success.

        trials: 1
        data: collect_test
        output: train_test

        see more detailsby log file:  "/var/log/keentune/keentuned-sensitize-train-1652175348.log"
```
任务发起后，可以通过log 查看具体任务执行情况。
```sh
>>> cat /var/log/keentune/keentuned-sensitize-train-1652175348.log
```
如果不再使用某个采集任务的数据了，可使用 keentune sensitize delete 命令删除对应的采集数据。
```sh
>>> keentune sensitize delete --data collect_test
[Warning] Are you sure you want to permanently delete job data 'collect_test' ?Y(yes)/N(no)y
[ok] collect_test delete successfully
```