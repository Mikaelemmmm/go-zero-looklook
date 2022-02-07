package {{.PkgName}}

import (
	"net/http"

	"looklook/common/result"
	{{.ImportPackages}}
	{{if .After1_1_10}}{{if .HasRequest}}"github.com/zeromicro/go-zero/rest/httpx"{{end}}{{end}}
)

func {{.HandlerName}}(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		{{if .HasRequest}}var req types.{{.RequestType}}
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		{{end}}l := {{.LogicName}}.New{{.LogicType}}(r.Context(), ctx)
		{{if .HasResp}}resp, {{end}}err := l.{{.Call}}({{if .HasRequest}}req{{end}})
		result.HttpResult(r, w, {{if .HasResp}}resp{{else}}nil{{end}}, err)
	}
}
