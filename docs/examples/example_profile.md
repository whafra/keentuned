# KeenTune：专家调优最佳实践
## 背景介绍
KeenTune中提供了多个专家调优方案，可以对操作系统内核参数进行一键式调优，根据运行的应用不同，我们可以选择不同的优化方案，例如CPU密集型的应用选择cpu_high_load.conf（cpu高负载）优化方案。KeenTune中内置的优化方案有：
+ cpu_high_load.conf（cpu高负载）
+ io_high_throughput.conf（IO高吞吐）
+ net_high_throuput.conf（网络高吞吐）
+ net_low_latency.conf（网络低时延）
### 题目要求
本题目需要你在我们提供的实验环境中安装并启动KeenTune，使用KeenTune选择一个专家调优方案设置到实验环境中，并在设置前后运行合适的Benchmark工具来验证优化效果。  
### 验收结果
调优前后的benchmark运行结果，证明调优方案的有效性。  

---  
## 操作指导
#### step 1. KeenTune安装和配置
安装和配置KeenTune，本题目中只需要安装keentune-target和keentuned，具体步骤参考[《KeenTune安装配置手册》](../install_cn.md)

#### step 2. 安装和运行Benchmark工具
安装对应的benchmark工具并运行
+ 验证io_high_throughput.conf（IO高吞吐）优化效果，使用fio工具，参考[《fio安装使用手册》](../benchmark-tools/fio_cn.md)
+ 验证net_high_throuput.conf（网络高吞吐）优化效果，使用wrk工具，参考[《wrk安装使用手册》](../benchmark-tools/wrk_cn.md)
+ 验证net_low_latency.conf（网络低时延）优化效果，使用wrk工具，参考[《wrk安装使用手册》](../benchmark-tools/wrk_cn.md)

#### step 3. KeenTune设置专家调优方案
使用KeenTune设置对应的专家调优方案到实验环境中，具体步骤参考[《KeenTune专家调优手册》](../profile_cn.md)

#### step 4. 验证优化结果
再次运行benchmark工具，观察优化效果  

## 子题目
本题目根据profile不同分为以下3个子题目
+ 专家调优最佳实践——IO高吞吐
+ 专家调优最佳实践——网络低时延
+ 专家调优最佳实践——网络高吞吐