package config

// wechat mini config.
type KqServerConfig struct {
	Address string `json:"AppId"`  //wechat mini appId
	Secret  string `json:"Secret"` //wechat mini secret
}
