{{.head}}

package {{.filePackage}}

import (
	"context"

	{{.package}}

	"github.com/tal-tech/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	{{.alias}}

	{{.serviceName}} interface {
		{{.interface}}
	}

	default{{.serviceName}} struct {
		cli zrpc.Client
	}
)

func New{{.serviceName}}(cli zrpc.Client) {{.serviceName}} {
	return &default{{.serviceName}}{
		cli: cli,
	}
}

{{.functions}}
