# KeenTune：智能参数调优
## 题目介绍
智能调优在系统优化领域有着重要的应用，KeenTune也提供了参数智能调优的功能，并内置了多种参数优化算法，可以对操作系统内核参数和应用参数进行智能化调优，与专家调优方案不同，参数智能调优需要耗费的时间更长。KeenTune中实现的智能优化算法包括：
+ Bayesian Optimization，贝叶斯优化
+ Tree-structured Parzen Estimator，TPE算法，贝叶斯优化算法的改进形式
+ Hyperparameter Optimization using RBFbasedsurrogate and DYCORS，HORD算法，高维空间表现较好

KeenTune中已经实现的参数优化域包括：
+ 操作系统内核参数
+ Nginx应用参数
+ Mysql应用参数

### 题目要求
本题目中，我们希望你在提供的实验环境中安装并启动KeenTune，使用KeenTune选择一个智能优化算法对内核参数或者Nginx应用参数进行调优，并在调优前后运行合适的benchmark工具验证优化效果。

### 验收结果
调优前后的benchmark运行结果，证明调优方案的有效性。

## 操作指导
#### step 1. KeenTune安装和配置
安装和配置KeenTune，本题目中需要在VM1上安装keentune-brain，keentune-target和keentuned，VM2上安装keentune-bench，并确保VM1和VM2网络连通。具体步骤参考[《KeenTune安装配置手册》](../install_cn.md)

#### step 2. 安装Nginx服务端
在VM1上安装启动Nginx服务。具体步骤参考[《Nginx安装配置》](../application/nginx.md)

#### step 3. 安装和运行wrk benchmark工具
在VM2上安装部署wrk工具，并以VM1作为目标运行wrk命令。具体安装和运行步骤参考[《wrk安装使用手册》](../benchmark-tools/wrk_cn.md)

#### step 4. 使用KeenTune进行参数智能调优
使用KeenTune基于wrk性能指标对内核参数或者Nginx应用参数进行参数智能调优。具体步骤参考[《KeenTune智能调优》](../tuning_cn.md)

#### step 5. 验证优化结果
再次运行benchmark工具，观察优化效果