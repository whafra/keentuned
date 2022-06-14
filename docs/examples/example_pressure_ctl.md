# KeenTune：智能控压初识
Benchmark压力不足也会对参数智能调优的工作带来困扰，使benchmark成为性能瓶颈导致调优失效。KeenTune提供了对benchmark参数的智能调整能力，用来解决benchmark压力参数选择困难的问题。目前KeenTune已经对以下几个benchmark工具进行了智能控压的适配：
+ [iperf3](../../daemon/examples/benchmark/iperf/iperf.py)
+ [sysbench](../../daemon/examples/benchmark/sysbench/sysbench_mysql_read_write.py)
+ [wrk](../../daemon/examples/benchmark/wrk/ack_nginx_http_long_base.py)

KeenTune中实现的智能控压优化算法包括：
+ Tree-structured Parzen Estimator，TPE算法，贝叶斯优化算法的改进形式
+ Hyperparameter Optimization using RBFbasedsurrogate and DYCORS，HORD算法，高维空间表现较好

### 题目要求
本题目中，我们希望你在提供的实验环境中安装并启动KeenTune，安装iperf和sysbench并使用KeenTune对benchmark工具的压力参数进行调整，获得合适的benchmark压力参数。
### 验收结果
智能控压调整之后的benchmark工具压力参数

## 操作指导
#### step 1. KeenTune安装和配置
安装和配置KeenTune，本题目中需要在VM1(Client)上安装keentune-brain，keentune-target和keentuned，VM2(Server)上安装keentune-bench，并确保VM1和VM2网络连通。具体步骤参考[《KeenTune安装配置手册》](../install_cn.md)  

#### step 2. 安装benchmark工具
根据题目要求安装iperf工具或者sysbench工具，分别参考[《iperf3安装使用手册》](../benchmark-tools/iperf3_cn.md)[《sysbench安装使用手册》](../benchmark-tools/sysbench_cn.md)  

#### step 3. 使用KeenTune进行benchmark压力调整
根据题目要求选择KeenTune中的TPE算法和HORD算法对benchmark的压力参数进行调整。具体步骤参考[《KeenTune智能控压》](../pressure_control_cn.md)

## 子题目
本题目根据算法和benchmark不同分为以下六个子题目
+ 智能控压初识——iperf（TPE）
+ 智能控压初识——iperf（HORD）
+ 智能控压初识——sysbench（TPE）
+ 智能控压初识——sysbench（HORD）