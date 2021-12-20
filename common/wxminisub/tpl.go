package wxminisub

//订单支付成功
const OrderPaySuccessTemplateID = "QIJPmfxaNqYzSjOlXGk1T6Xfw94JwbSPuOd3u_hi3WE"

type OrderPaySuccessDataParam struct {
	Sn            string
	GoodsName     string
	PayTotal      string
	LiveStateDate string
	LiveEndDate   string
}

func OrderPaySuccessData(params OrderPaySuccessDataParam) map[string]string {
	return map[string]string{
		"character_string6": params.Sn,
		"thing1":            params.GoodsName,
		"amount2":           params.PayTotal,
		"time4":             params.LiveStateDate,
		"time5":             params.LiveEndDate,
	}
}

//支付成功入驻通知
const OrderPaySuccessLiveKnowTemplateID = "kmm-maRr6v_9eMxEPpj-5clJ2YW_EFpd8-ngyYk63e4"

type OrderPaySuccessLiveKnowDataParam struct {
	LiveStartDate string
	LiveEndDate   string
	TradeCode     string
	Remark        string
}

func OrderPaySuccessLiveKnowData(params OrderPaySuccessLiveKnowDataParam) map[string]string {
	return map[string]string{
		"date2":             params.LiveStartDate,
		"date3":             params.LiveEndDate,
		"character_string4": params.TradeCode,
		"thing1":            params.Remark,
	}
}
