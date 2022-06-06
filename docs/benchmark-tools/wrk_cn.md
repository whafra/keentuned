# wrk安装使用手册
## 工具介绍
wrk 是一款针对 http 协议的基准测试工具，它能够在单机多核 CPU 的条件下，使用系统自带的高性能 I/O 机制，如 epoll，kqueue 等，通过多线程和事件模式，对目标机器（服务端）产生大量的负载。即wrk能够开启多个连接访问接口，看接口最多每秒可以承受多少连接。

## 安装命令
使用以下命令即可成功安装wrk工具
```sh
git clone https://gitee.com/mirrors/wrk.git wrk
cd wrk
make
cp wrk /usr/bin/
```

## 参数选择和运行
命令运行格式
```sh
wrk <选项> <被测HTTP服务的URL>
```
参数说明  
```sh
-c, --connections：跟服务器建立并保持的TCP连接数量 (请求并发数量)
-d, --duration：压测时间，e.g. 2s, 2m, 2h，默认单位是s
-t, --threads：使用多少个线程进行压测 一般设置成cpu的核数的2或者4倍
-s, --script：指定Lua脚本路径
-H, --header：为每一个HTTP请求添加HTTP头，e.g. "Connection: Close"    
    --latency：在压测结束后，打印延迟统计信息
    --timeout：超时时间
-v, --version：打印正在使用的wrk的详细版本信息
```
当前工程支持的有四种发压形式：http长链接、http短链接、https长链接、https短链接，运行命令如下  
```sh
# http长链接
wrk -t 10 -c 300 -d 30 --latency http://127.0.0.1
# http短链接
wrk -H "Connection: Close" -t 10 -c 300 -d 30 --latency http://127.0.0.1
# https长链接
wrk -t 10 -c 300 -d 30 --latency https://127.0.0.1
# https短链接
wrk -H "Connection: Close" -t 10 -c 300 -d 30 --latency https://127.0.0.1
```
###### 注：-t、-c、-d 三个参数值的确定是根据不同环境有所不同，原则上是将环境压力打满时对应的值
## 性能指标分析
以  http长链接  为例来说明性能指标，使用10个线程300个连接，对本机进行30秒的压测，并要求在压测结果中输出响应延迟信息，输出结果如下所示  
```sh
>>> wrk -t 10 -c 300 -d 30 --latency http://127.0.0.1
Running 30s test @ http://127.0.0.1
  10 threads and 300 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     4.09ms    2.58ms  55.15ms   88.69%
    Req/Sec     7.79k     1.00k   16.27k    68.50%
  Latency Distribution
     50%    3.57ms
     75%    4.78ms
     90%    6.55ms
     99%   14.83ms
  2328377 requests in 30.07s, 9.09GB read
Requests/sec:  77430.64
Transfer/sec:    309.47MB
```
参数详解
```sh
  Thread Stats    Avg      Stdev      Max      +/- Stdev
  （线程统计） （平均值）（标准差）（最大值）（正负一个标准差占比）
     Latency     4.09ms    2.58ms    55.15ms    88.69%
     （响应时间）
     Req/Sec     7.79k     1.00k     16.27k     68.50%
     （每线程每秒完成请求数）
 
  Latency Distribution （延迟分布）
      50%    3.57ms （有50%的请求执行时间是在3.57ms内完成）
      75%    4.78ms
      90%    6.55ms
      99%   14.83ms  （99分位的延迟：%99的请求在14.83ms内完成）
  2328377 requests in 30.07s, 9.09GB read （30.07秒内共处理完成了2328377个请求，读取了9.09GB数据）
Requests/sec: 77430.64 （平均每秒处理完成77430.64个请求）（也就是QPS=接口每秒的查询数）
Transfer/sec: 309.47MB （平均每秒读取数据309.47MB）
```
说明：
（1）一般来说我们主要关注平均值和最大值. 标准差如果太大说明样本本身离散程度比较高. 有可能系统性能波动很大.
（2）可以测试自己编写接口的qps，即每秒可以承受的最大访问连接数。 通过增大连接数，可以找到qps开始变小的数值，此时就是接口能够支持的最优连接数。大于最优连接数会降低qps。