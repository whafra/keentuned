# iperf3安装使用手册
## 工具介绍
iperf3是一款带宽测试工具,它支持调节各种参数,比如通信协议,数据包个数,发送持续时间,测试完会报告网络带宽,丢包率和其他参数。  
需要说明的是，iperf3是进行四层网络测试的工具，因此在准备环境的时候，需要是两台能够联网的机器，一台作为server、一台作为client，从而实现对TPC/UDP的网络带宽的测试。当然，单机也是可以正常安装和使用iperf3的（ip为127.0.0.1），只是这样测试出来的结果意义不大。  
## 安装命令
方法一：使用yum安装
```sh
yum install -y iperf3
```
方法二：使用源码安装
```sh
git clone https://github.com/esnet/iperf.git
cd iperf3
./configure 
make
make install
```
## 参数选择和运行
### 在server侧启动
```sh
iperf3 -s -p 8888
```
### 在client侧执行测试命令
简单进行网络带宽测试的示例：
```sh
iperf3 -c 192.168.1.123 -p 8888 -t 10
```
指定网络包为10M，进行UDP测试的示例：
```sh
iperf3 -c 192.168.1.123 -p 8888 -4 -f K -n 10M -b 10M --get-server-output(-u)
```
iperf3的参数很多，详细可以查看其help，在第二个例子中使用到的参数说明如下：
```sh
- c 指定client端
- p 指定端口(要和服务器端一致)
- B 绑定客户端的ip地址
- 4 指定ipv4
- f 格式化带宽数输出
- n 指定传输的字节数
- b 使用带宽数量
- u 指定udp协议
--get-server-output 获取来自服务器端的结果
```
## 性能指标分析
iperf的测试结果中，如果是TCP的话主要关注测试出的网络带宽，如果是UDP的话除了带宽也要关注丢包率。  
附录中给出的是TCP的测试结果，主要关注的是作为发送端的带宽。  
```sh
[ ID] Interval Transfer Bitrate Retr
[ 5] 0.00-10.00 sec 41.2 GBytes 35.4 Gbits/sec 10 sender
[ 5] 0.00-10.04 sec 0.00 GBytes 0.00 Gbits/sec receiver
```
附：iperf3测试结果示例
```sh
[root@iZbp11sdj1sc8o3r17rnwgZ ~]# iperf3 -c 192.168.1.123 -p 8888 -t 10
Connecting to host 192.168.1.123, port 8888
[ 5] local 192.168.1.121 port 44818 connected to 192.168.1.123 port 8888
[ ID] Interval Transfer Bitrate Retr Cwnd
[ 5] 0.00-1.00 sec 4.30 GBytes 36.9 Gbits/sec 1 3.06 MBytes
[ 5] 1.00-2.00 sec 3.61 GBytes 31.0 Gbits/sec 2 3.31 MBytes
[ 5] 2.00-3.00 sec 4.27 GBytes 36.7 Gbits/sec 1 3.31 MBytes
[ 5] 3.00-4.00 sec 3.56 GBytes 30.5 Gbits/sec 1 3.31 MBytes
[ 5] 4.00-5.00 sec 4.43 GBytes 38.1 Gbits/sec 0 3.31 MBytes
[ 5] 5.00-6.00 sec 3.67 GBytes 31.5 Gbits/sec 0 3.12 MBytes
[ 5] 6.00-7.00 sec 4.45 GBytes 38.2 Gbits/sec 0 3.12 MBytes
[ 5] 7.00-8.00 sec 4.44 GBytes 38.2 Gbits/sec 0 3.12 MBytes
[ 5] 8.00-9.00 sec 4.37 GBytes 37.5 Gbits/sec 0 3.12 MBytes
[ 5] 9.00-10.00 sec 4.09 GBytes 35.1 Gbits/sec 5 3.12 MBytes
- - - - - - - - - - - - - - - - - - - - - - - - -
[ ID] Interval Transfer Bitrate Retr
[ 5] 0.00-10.00 sec 41.2 GBytes 35.4 Gbits/sec 10 sender
[ 5] 0.00-10.04 sec 0.00 GBytes 0.00 Gbits/sec receiver
iperf Done.
```