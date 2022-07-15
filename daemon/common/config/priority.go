package config

// PriorityList append this white List when new parameter.json add to examples/parameter
const PRILevel = 2
const NginxDomain = "nginx"

var PriorityList = map[string]int{
	"iperf":      1,
	"sysctl":     1,
	"cri_sysctl": 1,
	"nginx":      0,
}
