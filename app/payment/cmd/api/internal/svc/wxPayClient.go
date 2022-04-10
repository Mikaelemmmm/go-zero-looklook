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
		return nil, errors.Wrapf(xerr.NewErrMsg("wechat pay fail"), " wechat pay init fail ，mchPrivateKey err : %v \n", err)
	}

	ctx := context.Background()
	// Initialize the client with the merchant's private key, etc., and make it have the ability to automatically obtain WeChat payment platform certificates at regular intervals
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(c.WxPayConf.MchId, c.WxPayConf.SerialNo, mchPrivateKey, c.WxPayConf.APIv3Key),
	}
	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrMsg("wechat pay fail"), "new wechat pay client err:%s", err)
	}

	return client, nil

}
