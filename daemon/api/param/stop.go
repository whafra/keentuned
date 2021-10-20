package param

import (
	m "keentune/daemon/modules"
	com "keentune/daemon/api/common"
	"os"
)

// Stop run param stop service
func (s *Service) Stop(request string, reply *string) error {
	if com.IsTuning {
		m.StopSig <- os.Interrupt
	}

	return nil
}
