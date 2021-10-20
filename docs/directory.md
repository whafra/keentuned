# keentuned 目录结构介绍
- cli/
  cli 主函数入口
  - api.go
    keentune rpc 接口调用部分
  - common.go
    keentune cli 部分公共函数
  - main.go
    keentune 所有命令主函数
  - param.go
    keentune param 系列命令初始化
  - profile.go
    keentune profile 系列命令初始化
  - sensitize.go
    keentune sensitize 系列命令初始化
- daemon
    KeenTune 守护进程实现
   - daemon.go
     KeenTune 守护进程入口
   - api/
      所有的守护进程服务实现
     - common/
      KeenTune 守护进程的公共服务，包括心跳检测，开放的http 服务
     - param/
      parameter 命令rpc服务的集合
      - delete.go
       param delete 服务入口，删除指定的历史调优数据（仅限删除用户生成的配置文件）
      - dump.go
       param dump 服务入口，将指定的最优配置json文件转换为conf配置文件
      - list.go
       param list 服务入口, 展示可用的调优配置，其中历史执行完成的调优结果以调优时指定的名称（调优时产生的json文件按照此目录管理）展示，经验配置以json文件名展示
      - rollback.go
       param rollback 服务入口，执行调优后的环境回滚操作
      - stop.go
       param stop 服务入口，停止正在执行的敏感参数任务
      - tune.go
       param tune 服务入口，下发指定参数动态调优（核心命令）
     - profile/
      profile 命令rpc服务的集合
      - delete.go
       profile delete 服务入口，删除指定的conf文件（仅限删除用户生成的配置文件）
      - generate.go
       profile generate 服务入口，将指定的conf文件转换为json配置文件
      - info.go
       profile info 服务入口，查看配置文件信息
      - list.go
       profile list 服务入口, 展示可用的静态调优配置列表
      - rollback.go
       profile rollback 服务入口，执行静态调优后的环境回滚操作
      - set.go
       profile set 服务入口，下发静态调优      
     - sensitize/
      sensitize rpc服务的集合
      - collect.go 
        sensitize collect 服务入口，执行敏感参数采集任务
      - delete.go
       sensitize delete 服务入口，删除指定的敏感参数数据
      - list.go
       sensitize list 服务入口, 展示可用的敏感参数训练数据列表
      - stop.go
       sensitize stop 服务入口，停止正在执行的敏感参数任务
      - train.go
        sensitize train 服务入口，执行敏感参数训练任务
     - system/
      keentuned 守护进程系统操作，包括耗时服务的执行状况查看，耗时服务的停止
      - message.go
        keentune msg --name task_name 耗时服务的执行状况查看, task_name为枚举值: "param tune", "sensitize collect", "sensitize train".
  - common/
    KeenTune golang 代码的公共库函数
  - examples/
    KeenTune 经验配置，安装后默认存放在/etc/keentune/下
     - benchmark/wrk
       wrk 压测经验配置和脚本
     - parameter/
       经验调优参数集合
     - profile/
       经验参数静态调优配置文件
  - modules/
    本目录实现rpc服务的业务逻辑处理    
- go.mod
  golang的依赖包管理文件
- go.sum
- keentuned_install.sh
  KeenTune golang 源码安装部署脚本，安装前可先修改第一部分的参数，也可直接执行默认配置参数
- keentuned.conf
  KeenTune 的守护服务进程的配置文件，可参考使用手册按需修改