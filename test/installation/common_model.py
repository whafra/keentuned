import time
import unittest
import logging
import subprocess

logger = logging.getLogger(__name__)


def sysCommand(cmd):
    result = subprocess.run(
                    cmd, 
                    shell=True,
                    close_fds=True,
                    stderr=subprocess.PIPE,
                    stdout=subprocess.PIPE
                )
    if result.returncode != 0:
        err_log=result.stderr.decode('UTF-8','strict')
        logger.info(err_log)
        return False, err_log
    else:
        return True, result.stdout.decode('UTF-8','strict').strip()

def clearKeentuneEnv():
    cmd = "rm -rf /usr/local/lib/python3.6/site-packages/keentune*;\
           rm -rf /usr/local/lib/python3.6/site-packages/brain;\
           rm -rf /usr/local/lib/python3.6/site-packages/target;\
           rm -rf /usr/local/lib/python3.6/site-packages/bench;\
           rm -f /usr/local/bin/keentune*;\
           yum remove -y keentuned keentune-brain keentune-bench keentune-target;\
           rpm -evh keentuned keentune-brain keentune-bench keentune-target;\
           ps -ef|grep -E 'keentuned|keentune-brain|keentune-target|keentune-bench'|grep -v grep|awk '{print $2}'| xargs -I {} kill -9 {}"
    sysCommand(cmd)


def getOsType():
    os_type = ""
    debian_type = ["Ubuntu", "Debian"]
    red_hat_type = ["Anolis", "CentOS", "Alibaba Cloud Linux"]

    result = sysCommand("cat /etc/os-release | grep ^NAME |awk -F= '{print$2}'")[1]
    for type_name in red_hat_type:
        if type_name in result:
            os_type = "redhat"
            break
    for type_name in debian_type:
        if type_name in result:
            os_type = "debian"
            break
    return os_type


def installPackage():
    package = [("tornado", "6.1"), ("numpy", "1.19.5"), ("POAP", "0.1.26"), ("bokeh", "2.3.2"), 
           ("hyperopt", "0.2.5"), ("scikit-learn", "0.24.2"), ("pySOT", "0.3.3"), ("paramiko", "2.7.2")]

    assert sysCommand("pip3 install pbr")[0] == True
    assert sysCommand("pip3 install pynginxconfig")[0] == True
    for item in package:
        status = sysCommand("pip3 install {}=={}".format(item[0], item[1]))[0]
        assert status == True


def redHatPrepare(os_arch, flag):
    if not sysCommand("go version")[0]: 
        sysCommand("yum -y install go")
    if not sysCommand("pip3 -V")[0]: 
        if not sysCommand("ln -s /usr/local/bin/pip3.[6-9] /usr/bin/pip3")[0]:
            sysCommand("ln -s /usr/bin/python3 /usr/bin/python")
    
    if os_arch == "aarch64":
        if not sysCommand("make --version")[0]: 
            sysCommand("yum -y install make")
        sysCommand("yum install -y gcc-c++")
        if sysCommand("pip3 -V | awk '{print$2}'")[1] != "21.2.4": 
            sysCommand("pip3 install --upgrade pip")
    if os_arch == "x86_64":
        if int(sysCommand("pip3 -V | awk '{print$2}' | awk -F'.' '{print$1}'")[1]) < 20:
            sysCommand("pip3 install --upgrade pip")

    installPackage()

    if flag == "rpm":
        if not sysCommand("rpmbuild --version")[0]: 
            sysCommand("yum -y install rpmdevtools")
        sysCommand("rpmdev-setuptree")


def debianPrepare(os_arch, platform_name, flag):
    rpm_list = ["BUILD", "BUILDROOT", "RPMS", "SOURCES", "SPECS", "SRPMS"]
    sysCommand("sudo apt update")
    if not sysCommand("pip3 -V")[0]: 
        sysCommand("apt install python3-pip -y")
    if int(sysCommand("pip3 -V | awk '{print$2}' | awk -F'.' '{print$1}'")[1]) < 20:
        sysCommand("pip3 install --upgrade pip")
    
    res1 = sysCommand("go version | awk '{print$3}' | awk -F'.' '{print$2}'")[0]
    res2 = int(sysCommand("go version | awk '{print$3}' | awk -F'.' '{print$2}'")[1])
    if not res1 or res2 < 13:
        if int(sysCommand("apt search golang 2>/dev/null | grep -i \"golang-go/\" | awk '{print$2}' | awk -F':' '{print $2}' | awk -F'~' '{print $1}' | awk -F'.' '{print$2}'")[1]) > 13:
            sysCommand("apt install golang -y")
        else:
            status = sysCommand("wget https://storage.googleapis.com/golang/go1.15.11.linux-{0}.tar.gz;\
                            tar -C /usr/local -xzf go1.15.11.linux-{0}.tar.gz;\
                            rm -f go1.15.11.linux-{0}.tar.gz".format(platform_name))
            assert status == True
            sysCommand("/bin/cp -f /usr/local/go/bin/go /usr/bin/go")
            sysCommand("go env -w GOROOT=\"/usr/local/go\"")

    if flag == "rpm":
        if not sysCommand("rpm --version")[0]:
            self.assertTrue(sysCommand("apt install -y rpm")[0])
        sysCommand("mkdir -p /root/rpmbuild")
        for rpm_dir in rpm_list:
            sysCommand("mkdir -p /root/rpmbuild/{}".format(rpm_dir))
    
    if os_arch == "aarch64":
        if not sysCommand("g++ --version")[0]: 
            sysCommand("apt install g++ -y")
    
    installPackage()


def serverPrepare(flag):
    os_arch = sysCommand("arch")[1]
    os_type = getOsType()
    if os_type == "redhat":
        redHatPrepare(os_arch, flag)
    elif os_type == "debian":
        platform_name="arm64" if os_arch=="aarch64" else "amd64"
        debianPrepare(os_arch, platform_name, flag)
    else:
        msg = "this system is not currently supported!"
        logger.error("prepare server failed, error is: {}".format(msg))
        return False
    return True
