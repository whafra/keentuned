%define debug_package %{nil}
%define anolis_release 1

#
# spec file for package golang-keentuned
#

Name:           keentuned
Version:        1.1.2
Release:        %{?anolis_release}%{?dist}
Url:            https://gitee.com/anolis/keentuned
Summary:        KeenTune tuning tools
License:        MulanPSLv2
Source:         https://gitee.com/anolis/keentuned/repository/archive/%{name}-%{version}.tar.gz

BuildRoot:      %{_tmppath}/%{name}-%{version}-build
BuildRequires:  go >= 1.13
BuildRequires:	systemd

Requires(post): systemd
Requires(preun): systemd
Requires(postun): systemd

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
* Wed Dec 15 2021 Runzhe Wang <15501019889@126.com> - 1.1.2
- remove unsupported profile in anolis23
- remove useless requires in .service file

* Thu Mar 5 2022 happy_orange <songnannan@linux.alibaba.com> - 1.1.1
- add makefile
- update spec file

* Wed Dec 15 2021 Runzhe Wang <15501019889@126.com> - 1.0.0
- add tpce tpch benchmark files
- fix bug: can not running in alinux2 and centos7
- change modify codeup address to gitee
- manage keentuned with systemctl
- fix: show brain error in the keentuned log
- fix: profile set supports absolute and relative paths
- fix: show exact job abort log after the stop command
- add nginx_conf parameter config file
- use '%license' macro
- update license to MulanPSLv2
- Init Keentuned.