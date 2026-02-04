Name:           svcm
Version:        0.1.0
Release:        1%{?dist}
Summary:        Lightweight systemd service manager for Wayland

License:        MIT
URL:            https://github.com/your/svcm
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang
Requires:       systemd

%description
svcm is a unified service manager with CLI, TUI (k9s-style), GUI (Tray), and MCP interfaces.
It allows managing both user and system services (via --privileged).

%prep
%setup -q

%build
go build -o svcm ./src/cmd/svcm

%install
install -Dpm 0755 svcm %{buildroot}%{_bindir}/svcm

%files
%{_bindir}/svcm
%doc README.md

%changelog
* Wed Feb 04 2026 Arya <arya@example.com> - 0.1.0-1
- Initial package
