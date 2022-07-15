# How to sensitize knob
---  
## Why need sensitizing knob
过多的优化参数会导致Auto-tuning算法的效率下降，KeenTune提供了一系列的敏感参数识别算法来，针对某个特定的benchmark输出，计算和筛选哪些参数是更为有效的参数，从而为下一次调优提供参考，通过降低参数数量来提升Auto-tuning算法的效率。

## How to run knob sensitizing tasks  
敏感参数识别需要由参数配置集合(X)及其benchmark得分(Y)组成的数据集{(X,Y),}作为算法的输入，KeenTune的[Auto-tuning](./How%20to%20tuning%20knobs_cn.md)运行过程中会自动保存中间数据来提供给敏感参数识别算法。我们可以通过`keentune param jobs`命令来查看所有tuning任务并选择其中**已经完成**的任务数据作为敏感参数的输入。
```s
keentune param jobs
name	algorithm	iteration	status	start_time	end_time
test	random	100	finish	2022-07-11 10:59:33	2022-07-11 11:00:01
```

具体命令是`keentune sensitize train --data [tuning job name] --job [sensi job name]  --trials [trials]`
```s
keentune sensitize train --data test --job test --trials 5
[ok] Running Sensitize Train Success.

	trials: 5
	data: test
	job: test

	see more detailsby log file:  "/var/log/keentune/keentuned-sensitize-train-test5.log"
```

`keentune sensitize jobs`命令可以看到所有正在运行的敏感参数识别任务，在同一时间只能运行一个敏感参数任务，如果要运行下一个敏感参数识别，需要使用命令`keentune sensitize stop`来终止前一个任务或等待其结束
```s
$keentune sensitize jobs
name	algorithm	trials	status	start_time	end_time
test1	shap	2	finish	2022-07-11 11:37:52	2022-07-11 11:39:35
test2	shap	2	finish	2022-07-11 11:40:10	2022-07-11 11:41:46
test	shap	5	running	2022-07-11 14:17:39	-
```