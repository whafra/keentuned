# fio安装使用手册
## fio介绍
FIO 工具是一款用于测试硬件存储性能的辅助工具，兼具灵活性、可靠性从而从众多性能测试工具中脱颖而出。磁盘的 I/O 是衡量硬件性能的最重要的指标之一，而 FIO 工具通过模拟 I/O负载对存储介质进行压力测试，并将存储介质的 I/O 数据直观的呈现出来。  
根据实际业务的场景，一般将 I/O 的表现分为四种场景，随机读、随机写、顺序读、顺序写。FIO 工具允许指定具体的应用模式，配合多线程对磁盘进行不同深度的测试。  
FIO 工具已经集成在 AnolisOS  8.2/8.4的yum仓库中，可以直接获取安装。也可以通过 github 的地址 https://github.com/axboe/fio.git 进行访问。  

## 安装命令
方法一：在Anolis 8.2/8.4使用yum安装  
```sh
$ yum install -y fio
```
方法二：从github获取最新源码编译安装
```sh
git clone https://github.com/axboe/fio.git
yum -y install gcc
cd fio
./configure
make
make install
```

## 参数选择和运行
### 运行命令示例
#### 顺序写：向/dev/sda分区存储上以2M块文件大小顺序写1100GB文件
```sh
fio -output=/tmp/100S100W -name=100S100W -filename=/dev/sda -ioengine=libaio -direct=1 -blocksize=2M -size=1100GB -rw=write -iodepth=8 -numjobs=1
```
#### 随机写: 向/dev/sda分区存储上以2M块文件大小随机写1100GB文件
```sh
fio -output=/tmp/100R100W -name=100R100W -filename=/dev/sdb:/dev/sdc:/dev/sdd -ioengine=libaio -direct=1 -blocksize=2M -size=3356GB -rw=randwrite -iodepth=8 -numjobs=1
```
#### 顺序读
```sh
fio -output=/tmp/100S100W -name=100S100W -filename=/dev/sda -ioengine=libaio -direct=1 -blocksize=2M –runtime=1800 -rw=read -iodepth=8 -numjobs=1
```
#### 随机读
```sh
fio -output=/tmp/100S100Wsdbsdcsdd -name=100S100W -write_bw_log=bw_log -write_lat_log=lat_log -filename=/dev/sdb:/dev/sdc:/dev/sdd -ioengine=libaio -direct=1 -blocksize=2M -runtime=1800 -rw=randread -iodepth=32 -numjobs=1
```
#### 混合随机读写：70%随机读，30%随机写，以2M块文件大小向/dev/sdb:/dev/sdc:/dev/sdd三个分区存储上随机读写300s时间
```sh
fio -output=/tmp/100S100W -name=100S100W -filename=/dev/sdb:/dev/sdc:/dev/sdd -ioengine=libaio -direct=1 -blocksize=2M -runtime=300 -rw=randrw -rwmixread=70 -rwmixwrite=30 -iodepth=32 -numjobs=1
```
### 参数信息
命令中使用参数具体使用方式参考如下：
```sh
filename=/dev/emcpowerb　支持文件系统或者裸设备，-filename=/dev/sda2或-filename=/dev/sdb
direct=1                 测试过程绕过机器自带的buffer，使测试结果更真实
rw=randwread             测试随机读的I/O
rw=randwrite             测试随机写的I/O
rw=randrw                测试随机混合写和读的I/O
rw=read                  测试顺序读的I/O
rw=write                 测试顺序写的I/O
rw=rw                    测试顺序混合写和读的I/O
bs=4k                    单次io的块文件大小为4k
bsrange=512-2048         同上，提定数据块的大小范围
size=5g                  本次的测试文件大小为5g，以每次4k的io进行测试
numjobs=30               本次的测试线程为30
runtime=1000             测试时间为1000秒，如果不写则一直将5g文件分4k每次写完为止
ioengine=psync           io引擎使用pync方式，如果要使用libaio引擎，需要yum install libaio-devel包
rwmixwrite=30            在混合读写的模式下，写占30%
group_reporting          关于显示结果的，汇总每个进程的信息
此外
lockmem=1g               只使用1g内存进行测试
zero_buffers             用0初始化系统buffer
nrfiles=8                每个进程生成文件的数量
```
## 性能指标分析
下图给出了FIO进行混合随机读写的测试结果作为示例，来说明需要关注的测试结果，也就是磁盘读写速度和时延。在给出的具体测试结果里，需要关注的部分是
```sh
……
read : io=28976KB, bw=2854.6KB/s, iops=178 , runt= 10151msec
    clat (usec): min=49 , max=525390 , avg=35563.60, stdev=69691.20
     lat (usec): min=49 , max=525390 , avg=35563.72, stdev=69691.20
……
  write: io=29616KB, bw=2917.6KB/s, iops=182 , runt= 10151msec
    clat (usec): min=64 , max=2030.4K, avg=19768.52, stdev=155468.56
     lat (usec): min=64 , max=2030.4K, avg=19768.86, stdev=155468.56
……
```
+ bw：磁盘的吞吐量，这个是顺序读写考察的重点，类似于下载速度。
+ iops：磁盘的每秒读写次数，这个是随机读写考察的重点
+ io总的输入输出量 
+ runt：总运行时间
+ lat (msec)：延迟(毫秒)

附：混合随机读写的测试结果
```sh
[root@localhost dev]#  fio -filename=/dev/sda1 -direct=1 -iodepth 1 -thread -rw=randrw -ioengine=psync -bs=16k -size=500M -numjobs=10 -runtime=10
 -group_reporting -name=mytest mytest: (g=0): rw=randrw, bs=16K-16K/16K-16K, ioengine=psync, iodepth=1
...
mytest: (g=0): rw=randrw, bs=16K-16K/16K-16K, ioengine=psync, iodepth=1
fio 2.0.7
Starting 10 threads
Jobs: 10 (f=10): [mmmmmmmmmm] [100.0% done] [1651K/1831K /s] [100 /111  iops] [eta 00m:00s]
mytest: (groupid=0, jobs=10): err= 0: pid=4075
  read : io=28976KB, bw=2854.6KB/s, iops=178 , runt= 10151msec
    clat (usec): min=49 , max=525390 , avg=35563.60, stdev=69691.20
     lat (usec): min=49 , max=525390 , avg=35563.72, stdev=69691.20
    clat percentiles (usec):
     |  1.00th=[   65],  5.00th=[   70], 10.00th=[   92], 20.00th=[  116],
     | 30.00th=[  137], 40.00th=[  151], 50.00th=[  175], 60.00th=[  286],
     | 70.00th=[14144], 80.00th=[69120], 90.00th=[138240], 95.00th=[197632],
     | 99.00th=[280576], 99.50th=[301056], 99.90th=[452608], 99.95th=[528384],
     | 99.99th=[528384]
    bw (KB/s)  : min=   16, max= 1440, per=12.25%, avg=349.55, stdev=236.44
  write: io=29616KB, bw=2917.6KB/s, iops=182 , runt= 10151msec
    clat (usec): min=64 , max=2030.4K, avg=19768.52, stdev=155468.56
     lat (usec): min=64 , max=2030.4K, avg=19768.86, stdev=155468.56
    clat percentiles (usec):
     |  1.00th=[   70],  5.00th=[   83], 10.00th=[   93], 20.00th=[  115],
     | 30.00th=[  131], 40.00th=[  141], 50.00th=[  151], 60.00th=[  161],
     | 70.00th=[  177], 80.00th=[  209], 90.00th=[  310], 95.00th=[  532],
     | 99.00th=[700416], 99.50th=[1056768], 99.90th=[2007040], 99.95th=[2023424],
     | 99.99th=[2023424]
    bw (KB/s)  : min=   21, max= 1392, per=11.98%, avg=349.49, stdev=253.50
    lat (usec) : 50=0.03%, 100=13.11%, 250=58.60%, 500=7.76%, 750=1.34%
    lat (usec) : 1000=0.38%
    lat (msec) : 2=0.87%, 4=0.74%, 10=0.38%, 20=1.37%, 50=3.63%
    lat (msec) : 100=2.57%, 250=6.80%, 500=1.64%, 750=0.33%, 1000=0.14%
    lat (msec) : 2000=0.27%, >=2000=0.05%
  cpu          : usr=0.04%, sys=1.52%, ctx=45844, majf=0, minf=798
  IO depths    : 1=100.0%, 2=0.0%, 4=0.0%, 8=0.0%, 16=0.0%, 32=0.0%, >=64=0.0%
     submit    : 0=0.0%, 4=100.0%, 8=0.0%, 16=0.0%, 32=0.0%, 64=0.0%, >=64=0.0%
     complete  : 0=0.0%, 4=100.0%, 8=0.0%, 16=0.0%, 32=0.0%, 64=0.0%, >=64=0.0%
     issued    : total=r=1811/w=1851/d=0, short=r=0/w=0/d=0
 
Run status group 0 (all jobs):
   READ: io=28976KB, aggrb=2854KB/s, minb=2854KB/s, maxb=2854KB/s, mint=10151msec, maxt=10151msec
  WRITE: io=29616KB, aggrb=2917KB/s, minb=2917KB/s, maxb=2917KB/s, mint=10151msec, maxt=10151msec
 
Disk stats (read/write):
  sda: ios=1818/1846, merge=0/5, ticks=63776/33966, in_queue=104258, util=99.84%
```