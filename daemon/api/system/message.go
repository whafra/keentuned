package system

import "keentune/daemon/common/log"
type Service struct {
}

// Message run keentune msg service
func (s *Service) Message(flag string, reply *string) error {
	if flag != log.ParamTune && flag != log.SensitizeTrain && flag != log.SensitizeCollect {
		return nil
	}
	
	*reply = log.ClientLogMap[flag]
	return nil
}
