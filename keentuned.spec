%define debug_package %{nil}
%define anolis_release 4

Name:           keentuned
Version:        1.1.3
Release:        %{?anolis_release}%{?dist}
Url:            https://gitee.com/anolis/keentuned
Summary:        KeenTune tuning tools
Vendor:         Alibaba
License:        MulanPSLv2
Source:         https://gitee.com/anolis/keentuned/repository/archive/%{name}-%{version}.tar.gz

BuildRequires:  go >= 1.13
BuildRequires:	systemd

Requires(post): systemd
Requires(preun): systemd
Requires(postun): systemd

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
if [ -f "%{_prefix}/lib/systemd/system/keentuned.service" ]; then
    systemctl enable keentuned.service || :
    systemctl start keentuned.service || :
fi

%preun
%systemd_preun keentuned.service

%postun
%systemd_postun_with_restart keentuned.service

%files
%defattr(0644,root,root, 0755)
%license LICENSE
%doc README.md docs/*
%attr(0755, root, root) %{_bindir}/keentune
%attr(0755, root, root) %{_bindir}/keentuned
%dir %{_sysconfdir}/keentune
%dir %{_sysconfdir}/keentune/conf
%{_sysconfdir}/keentune
%{_prefix}/lib/systemd/system/keentuned.service

%changelog
* Tue May 24 2022 Runzhe Wang <runzhe.wrz@alibaba-inc.com> - 1.1.3-4
- Fix getting real IP failure during system initialization.

* Mon May 23 2022 Runzhe Wang <runzhe.wrz@alibaba-inc.com - 1.1.3-3
- modify servie type to idle

* Fri May 20 2022 happy_orange <songnannan@linux.alibaba.com> - 1.1.3-2
- rebuild

* Thu May 19 2022 happy_orange <songnannan@linux.alibaba.com> - 1.1.3-1
- update spec

* Thu May 19 2022 Runzhe Wang <runzhe.wrz@alibaba-inc.com - 1.1.3
- fix bug in service exit

* Mon May 09 2022 Runzhe Wang <runzhe.wrz@alibaba-inc.com - 1.1.2
- remove unsupported profile in anolis23
- remove useless requires in .service file

* Thu May 05 2022 happy_orange <songnannan@linux.alibaba.com> - 1.1.1
- add makefile
- update spec file

* Wed Dec 15 2021 Runzhe Wang <runzhe.wrz@alibaba-inc.com - 1.0.0
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