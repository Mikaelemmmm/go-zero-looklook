package config

// 微信支付配置
type WxPayConf struct {
	MchId      string `json:"MchId"`      //微信商户id
	SerialNo   string `json:"SerialNo"`   //商户证书的证书序列号
	APIv3Key   string `json:"APIv3Key"`   //apiV3Key，商户平台获取
	PrivateKey string `json:"PrivateKey"` //privateKey：私钥 apiclient_key.pem 读取后的内容
	NotifyUrl  string `json:"NotifyUrl"`  //支付通知回调服务端地址
}
