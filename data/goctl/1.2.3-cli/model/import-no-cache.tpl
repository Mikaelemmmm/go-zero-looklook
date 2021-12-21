import (
	"database/sql"
	"fmt"
	"strings"
	{{if .time}}"time"{{end}}

	"looklook/common/globalkey"

	"github.com/tal-tech/go-zero/core/stores/builder"
	"github.com/tal-tech/go-zero/core/stores/sqlc"
	"github.com/tal-tech/go-zero/core/stores/sqlx"
	"github.com/tal-tech/go-zero/core/stringx"
)
