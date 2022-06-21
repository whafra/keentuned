%define debug_package %{nil}
%define anolis_release 1

Name:           keentuned
Version:        1.2.1
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
%setup -n %{name}-%{version}

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
%doc README.md docs/directory.md
%attr(0755, root, root) %{_bindir}/keentune
%attr(0755, root, root) %{_bindir}/keentuned
%dir %{_sysconfdir}/keentune
%dir %{_sysconfdir}/keentune/conf
%{_sysconfdir}/keentune
%{_prefix}/lib/systemd/system/keentuned.service

%changelog
* Mon Jun 20 2022 Runzhe Wang <runzhe.wrz@alibaba-inc.com> - 1.2.1-1
- update docs

* Thu May 05 2022 happy_orange <songnannan@linux.alibaba.com> - 1.2.0-2
- add makefile
- update spec file

* Mon Apr 04 2022 Runzhe Wang <runzhe.wrz@alibaba-inc.com> - 1.2.0
- Add capabilities of target-group and bench-group
- Fix some issues
- Add 'keentune version' command

* Thu Mar 03 2022 Runzhe Wang <runzhe.wrz@alibaba-inc.com> - 1.1.0
- remove parameter fs.nr_open
- Add support for GP (in iTuned) in sensitizing algorithms
- Add support for lasso in sensitizing algorithms
- refactor tornado module: replace await by threadpool
- lazy load domain in keentune-target
- fix other bugs
- Add baseline reading before init brain
- Supporting multiple param json for tuning
- Fix rollback failure
- Clean empty dir when uninstall
- Verify input arguments of command 'param tune'
- Supporting of multiple target tuning
- Fix bug which cause keentune hanging after command 'param stop'
- Add verification of conflicting commands such as 'param dump', 'param delete' when a tuning job is runing.
- Remove version limitation of tornado
- Refactor sysctl domain to improve stability of parameter setting
- Fix some user experience issues

* Wed Dec 15 2021 Runzhe Wang <runzhe.wrz@alibaba-inc.com> - 1.0.0
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