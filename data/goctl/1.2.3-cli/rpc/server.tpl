{{.head}}

package server

import (
	{{if .notStream}}"context"{{end}}

	{{.imports}}
)

type {{.server}}Server struct {
	svcCtx *svc.ServiceContext
}

func New{{.server}}Server(svcCtx *svc.ServiceContext) *{{.server}}Server {
	return &{{.server}}Server{
		svcCtx: svcCtx,
	}
}

{{.funcs}}
