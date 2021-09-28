package sensitize

import (
	m "keentune/daemon/modules"
	"os"
)

// Stop run sensitize stop service
func (s *Service) Stop(request string, reply *string) error {
	m.StopSig <- os.Interrupt
	return nil
}
