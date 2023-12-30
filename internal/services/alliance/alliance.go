package alliance

import "github.com/mccune1224/betrayal/pkg/data"

type AllianceHandler struct {
	m data.Models
}

func InitAllianceHandler(models data.Models) *AllianceHandler {
	ah := &AllianceHandler{
		m: models,
	}
	return ah
}
