package profile

import (
	com "keentune/daemon/api/common"
)

type Service struct {
}

// Delete run profile delete service
func (s *Service) Delete(flag com.DeleteFlag, reply *string) error {
	return com.RunDelete(flag, reply)
}
