package sensitize

import (
	com "keentune/daemon/api/common"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils/http"
	m "keentune/daemon/modules"
	"fmt"
	"os"
)

// Delete run sensitize delete service
func (s *Service) Delete(flag com.DeleteFlag, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.SensitizeDel]
		log.ClearCliLog(log.SensitizeDel)
	}()
	
	path := fmt.Sprintf("%s/sensi-%s.json", m.GetSensitizePath(), flag.Name)
	if file.IsPathExist(path) && !flag.Force {
		log.Errorf(log.SensitizeDel, "file %v exists, but given up to delete it", flag.Name)	
		return nil				
	}

	uri := config.KeenTune.BrainIP + ":" + config.KeenTune.BrainPort + "/sensitize_delete"
	if err := http.ResponseSuccess("POST", uri, map[string]interface{}{"data": flag.Name}); err != nil {
		log.Errorf(log.SensitizeDel, "Delete [%v] failed, err:%v", flag.Name, err)
		return err
	}

	os.RemoveAll(path)	
	log.Infof(log.SensitizeDel, "[ok] %v delete successfully.", flag.Name)
	return nil
}

