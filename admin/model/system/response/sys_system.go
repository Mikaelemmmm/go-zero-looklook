package response

import "looklook/admin/config"

type SysConfigResponse struct {
	Config config.Server `json:"config"`
}
