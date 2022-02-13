package config

// 微信小程序配置
type WxMiniConf struct {
	AppId  string `json:"AppId"`  //微信小程序appId（非公众号）
	Secret string `json:"Secret"` //微信小程序secret（非公众号）
}
