# KeenTune：Nginx敏感参数识别
## 题目介绍
操作系统内核参数和应用参数庞大的数量对参数智能调优带来了极大的挑战，使用敏感参数识别算法对参数进行筛选可以大大降低参数的规模，提高那参数智能调优的效率和准确性。KeenTune中实现的敏感参数识别算法有：
+ lasso, 稀疏化线性敏感参数识别算法
+ univariate, 基于互信息(mutual information)的单一敏感参数识别算法
+ shap，基于博弈论的非线性敏感参数识别算法
### 题目要求
本题目中，我们希望你在提供的实验环境中安装并启动KeenTune，使用KeenTune任意选择一个或者多个敏感参数优化算法对Nginx调优中使用的操作系统内核参数进行敏感性估计和参数筛选，我们会提供算法运行所必须的数据。  
### 验收结果
Nginx优化中操作系统内核参数的敏感性排序。

## 操作指导
#### step 1. KeenTune安装和配置
安装和配置KeenTune，本题目中需要在实验环境上安装keentune-brain和keentuned，具体步骤参考[《KeenTune安装配置手册》](../install_cn.md)

#### step 2. 将预置数据放到对应目录下
下载附件中的预置数据，解压保存到目录/var/keentune/data/tuning_data下
+ [http long](../data/demo-http-long.tar)
+ [https long](../data/demo-https-long.tar)
+ [http short](../data/demo-http-short.tar)
+ [https short](../data/demo-https-short.tar)

#### step 3. 使用KeenTune运行敏感参数识别算法
使用KeenTune选择敏感参数识别算法并运行，指定预置数据作为输入，通过输出结果筛选TOP10的内核参数。具体步骤参考[《KeenTune敏感参数识别》](../sensitization_cn.md)

## 子题目
本题目根据使用的数据不同分为以下4个子题目
+ Nginx敏感参数识别——http long
+ Nginx敏感参数识别——https long
+ Nginx敏感参数识别——http short
+ Nginx敏感参数识别——https short