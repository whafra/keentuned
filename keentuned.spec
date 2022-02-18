%define debug_package %{nil}
%define anolis_release 6

#
# spec file for package golang-keentuned
#

Name:           keentuned
Version:        1.0.0
Release:        %{?anolis_release}%{?dist}
Url:            https://gitee.com/anolis/keentuned
Summary:        KeenTune tuning tools
License:        MulanPSLv2
Source:         %{name}-%{version}.tar.gz

BuildRoot:      %{_tmppath}/%{name}-%{version}-build
BuildRequires:  go >= 1.13

Vendor:         Alibaba

%description
KeenTune tuning tools rpm package

%prep
%setup -n %{name}-%{version}

%build
go env -w CGO_ENABLED=0
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
cd daemon
go build -ldflags=-linkmode=external -o keentuned
mv -f keentuned ../
cd ../cli
go build -ldflags=-linkmode=external -o keentune
mv -f keentune ../
cd ../

%install
rm -rf $RPM_BUILD_ROOT
mkdir -p ${RPM_BUILD_ROOT}/usr/bin/
mkdir -p ${RPM_BUILD_ROOT}/etc/keentune/
mkdir -p ${RPM_BUILD_ROOT}/etc/keentune/conf/
mkdir -p ${RPM_BUILD_ROOT}/usr/lib/systemd/system/

cp -f keentune ${RPM_BUILD_ROOT}/usr/bin/keentune
cp -f keentuned ${RPM_BUILD_ROOT}/usr/bin/keentuned
cp -rf daemon/examples/. ${RPM_BUILD_ROOT}/etc/keentune
cp -f ./keentuned.conf ${RPM_BUILD_ROOT}/etc/keentune/conf/
cp -f ./keentuned.service ${RPM_BUILD_ROOT}/usr/lib/systemd/system/

%clean
[ "$RPM_BUILD_ROOT" != "/" ] && rm -rf "$RPM_BUILD_ROOT"
rm -rf $RPM_BUILD_DIR/%{name}-%{version}

%post
systemctl daemon-reload

%postun
CONF_DIR=%{_sysconfdir}/keentune/conf
if [ "$(ls -A $CONF_DIR)" = "" ]; then
        rm -rf $CONF_DIR
fi

%files
%defattr(0444,root,root, 0555)
%attr(0555, root, root) /usr/bin/keentune
%attr(0555, root, root) /usr/bin/keentuned
%{_bindir}/%{name}
%{_bindir}/keentune
%{_sysconfdir}/keentune
%license LICENSE
%{_prefix}/lib/systemd/system/keentuned.service

%changelog
* Tue Dec 21 2021 Lilinjie <lilinjie@uniontech.com> - 1.0.0-6
- add tpce tpch benchmark files

* Wed Dec 15 2021 Runzhe Wang <15501019889@126.com> - 1.0.0-5
- fix bug: can not running in alinux2 and centos7
- change modify codeup address to gitee

* Fri Dec 03 2021 wenchao <yuxiongkong159@gmail.com> - 1.0.0-4
- manage keentuned with systemctl
- fix: show brain error in the keentuned log
- fix: profile set supports absolute and relative paths
- fix: show exact job abort log after the stop command

* Wed Nov 24 2021 runzhe.wrz <15501019889@126.com> - 1.0.0-3
- add nginx_conf parameter config file

* Wed Nov 10 2021 wenchao <yuxiongkong159@gmail.com> - 1.0.0-2
- use '%license' macro
- update license to MulanPSLv2

* Sun Aug  1 2021 wenchao <yuxiongkong159@gmail.com> - 1.0.0-1
- Init Keentuned.
