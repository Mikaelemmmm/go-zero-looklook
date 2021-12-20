package svc

import (
	"context"
	"looklook/app/payment/cmd/api/internal/config"
	"looklook/common/xerr"

	"github.com/pkg/errors"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

func NewWxPayClientV3(c config.Config) (*core.Client, error) {

	mchPrivateKey, err := utils.LoadPrivateKey(c.WxPayConf.PrivateKey)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrMsg("微信支付失败"), " 微信支付初始化 wx pay client 失败 ，mchPrivateKey err : %v \n", err)
	}

	ctx := context.Background()
	// 使用商户私钥等初始化 client，并使它具有自动定时获取微信支付平台证书的能力
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(c.WxPayConf.MchId, c.WxPayConf.SerialNo, mchPrivateKey, c.WxPayConf.APIv3Key),
	}
	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrMsg("微信支付失败"), "new wechat pay client err:%s", err)
	}

	return client, nil

}
