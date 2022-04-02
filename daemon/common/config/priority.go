package config

// PriorityList append this white List when new parameter.json add to examples/parameter
const PRILevel = 2
var PriorityList = map[string]int{
	"sysctl":     1,
	"cri_sysctl": 1,
	"nginx": 0,
}
