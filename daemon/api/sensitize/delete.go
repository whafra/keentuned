package sensitize

import (
	"fmt"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils/http"
)

type Service struct {
}

// Delete run sensitize delete service
func (s *Service) Delete(flag com.DeleteFlag, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.SensitizeDel]
		log.ClearCliLog(log.SensitizeDel)
	}()

	uri := config.KeenTune.BrainIP + ":" + config.KeenTune.BrainPort + "/sensitize_delete"
	if err := http.ResponseSuccess("POST", uri, map[string]interface{}{"data": flag.Name}); err != nil {
		log.Errorf(log.SensitizeDel, "Delete %v failed, err:%v", flag.Name, err)
		return fmt.Errorf("Delete %v failed, err:%v", flag.Name, err)
	}

	return com.RunTrainDelete(flag, reply)
}

