package system

import (
	"fmt"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
)

// Init ...
func (s *Service) Init(flag string, reply *string) error {
	result, err := com.Initialize()
	if err != nil {
		*reply = fmt.Sprintf("%v %v", utils.ColorString("yellow", "[Warning]"), result)
		log.Warnln("", result)
		return nil
	}

	*reply = fmt.Sprintf("%v KeenTune Init success\n", utils.ColorString("green", "[OK]"))
	log.Info("", "KeenTune Init success")
	return nil
}

