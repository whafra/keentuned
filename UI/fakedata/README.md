fake data: mv all files in this directory to /var/   

### Directory Tree  
/var/keentune
	tuning_jobs.csv
    sensitize_jobs.csv
    
    tuning_workspace/
    	[task name]/
            bench.json
            group.conf
            knobs.json
            points.csv
            score.csv          
            
        [task name 2]/
            
    sensitize_workspace/
    	[task name]/
        	result.csv
            knobs.json
    
    files/
    profile/

/var/log/keentune
	xxxxx.log
    xxxxx.log


### tuning_jobs.csv
name：任务创建的时候写入
starttime：任务创建的时候写入
endtime：任务结束的时候更新
algorithm：任务创建的时候写入
iteration：任务创建的时候写入
status：running（问题：是否区分benchmark runing，algorithm runing），finish（正常结束），error（报错），abort（手动终止）
parameter config path: /var/keentune/[task_name]/target_group.conf
benchmark配置文件路径：任务创建的时候写入，（问题：多bench怎么设计）
日志文件路径：任务创建写入
*工作路径：workspace
*best文件：workspace+best.json（多个group怎么展示）
*数据文件路径: workspace+/data（brain传回）


### target_group.conf
```conf
[target-group-1]
target = 127.0.0.1, 127.0.0.2
parameter = /var/keentune/parameter/sysctl.json

[target-group-2]
target = 127.0.0.3
parameter = /var/keentune/parameter/sysctl.json, /var/keentune/parameter/nginx.json
```

### sensitize_jobs.csv
name：创建时写入
trials：创建时写入
status：running，finish（正常结束），error（报错），abort（手动终止）
epoch：创建时写入
log path：创建写入
result path：结束时更新
*工作路径：workspace

