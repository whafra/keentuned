package sensitize

import (
	m "keentune/daemon/modules"
	com "keentune/daemon/api/common"
	"os"
)

// Stop run sensitize stop service
func (s *Service) Stop(request string, reply *string) error {
	if com.IsSensitizing {
		m.StopSig <- os.Interrupt
	}

	return nil
}
