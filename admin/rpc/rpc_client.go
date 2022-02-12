package rpc

import (
	"fmt"
	"github.com/zeromicro/go-zero/zrpc"
	"looklook/admin/global"
	"looklook/app/banner/cmd/rpc/bannersrv"

	"sync"
)

var client *Client
var once sync.Once

type Client struct {
	BannerRpc bannersrv.BannerSrv
}

func InitClient() {

	fmt.Println(global.GVA_CONFIG.RpcConf.BannerRpcConf)
	once.Do(func() {
		client = &Client{
			BannerRpc: bannersrv.NewBannerSrv(zrpc.MustNewClient(global.GVA_CONFIG.RpcConf.BannerRpcConf)),
		}


		fmt.Printf("client : %+v \n",client)
	})
}

func GetClient() *Client {

	fmt.Printf("client : %+v \n",client)
	return client
}
