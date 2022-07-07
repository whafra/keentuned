package param

import (
	"fmt"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils/http"
)

type Service struct {
}

// Delete run param delete service
func (s *Service) Delete(flag com.DeleteFlag, reply *string) error {
	clientName := new(string)
	if com.IsBrainOffline(clientName) {
		return fmt.Errorf("brain client is offline, please get it ready")
	}

	uri := config.KeenTune.BrainIP + ":" + config.KeenTune.BrainPort + "/sensitize_delete"
	if err := http.ResponseSuccess("POST", uri, map[string]interface{}{"data": flag.Name}); err != nil {
		log.Errorf(log.ParamDel, "Delete %v failed, err:%v", flag.Name, err)
		return fmt.Errorf("Delete %v failed, err:%v", flag.Name, err)
	}

	return com.RunDelete(flag, reply)
}

