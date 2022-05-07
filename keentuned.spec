%define debug_package %{nil}
%define anolis_release 1

#
# spec file for package golang-keentuned
#

Name:           keentuned
Version:        1.1.1
Release:        %{?anolis_release}%{?dist}
Url:            https://gitee.com/anolis/keentuned
Summary:        KeenTune tuning tools
License:        MulanPSLv2
Source:         https://gitee.com/anolis/keentuned/repository/archive/%{name}-%{version}.tar.gz

BuildRoot:      %{_tmppath}/%{name}-%{version}-build
BuildRequires:  go >= 1.13

Vendor:         Alibaba

%description
KeenTune tuning tools rpm package

%prep
%autosetup -n %{name}-%{version}

%build
%make_build

%install
%make_install

%clean
[ "$RPM_BUILD_ROOT" != "/" ] && rm -rf "$RPM_BUILD_ROOT"
rm -rf $RPM_BUILD_DIR/%{name}-%{version}

%post
%systemd_post keentuned.service

%preun
%systemd_preun keentuned.service

%postun
%systemd_postun keentuned.service
CONF_DIR=%{_sysconfdir}/keentune/conf
if [ "$(ls -A $CONF_DIR)" = "" ]; then
        rm -rf $CONF_DIR
fi

%files
%defattr(0644,root,root, 0755)
%license LICENSE
%doc README.md docs/directory.md
%attr(0755, root, root) %{_bindir}/keentune
%attr(0755, root, root) %{_bindir}/keentuned
%{_sysconfdir}/keentune
%{_prefix}/lib/systemd/system/keentuned.service

%changelog
* Thu Mar 5 2022 happy_orange <songnannan@linux.alibaba.com> - 1.1.1
- add makefile
- update spec file

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