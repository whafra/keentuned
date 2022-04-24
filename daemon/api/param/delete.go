package param

import (
	com "keentune/daemon/api/common"
)

type Service struct {
}

// Delete run param delete service
func (s *Service) Delete(flag com.DeleteFlag, reply *string) error {
	return com.RunDelete(flag, reply)
}
