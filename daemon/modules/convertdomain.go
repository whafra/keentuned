package modules

var specialDomain = map[string]string{
	"scheduler.amd":   "scheduler",
	"scheduler_amd":   "scheduler",
	"sysctl.thunderx": "sysctl",
	"sysctl_thunderx": "sysctl",
	"vm.thunderx":     "vm",
	"vm_thunderx":     "vm",
}

func convertDomain(domain string) string {
	matchedDomain, find := specialDomain[domain]
	if find {
		return matchedDomain
	}

	return domain
}

