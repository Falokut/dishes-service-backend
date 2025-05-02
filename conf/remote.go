package conf

import (
	"reflect"

	"github.com/Falokut/go-kit/db"
	"github.com/Falokut/go-kit/log"
	"github.com/Falokut/go-kit/remote/schema"
	"github.com/Falokut/go-kit/tg_botx"
)

// nolint: gochecknoinits
func init() {
	schema.CustomGenerators.Register("logLevel", func(field reflect.StructField, t *schema.Schema) {
		t.Type = "string"
		t.Enum = []any{"debug", "info", "warn", "error", "fatal"}
	})
}

type Remote struct {
	LogLevel log.Level `schemaGen:"logLevel" schema:"Уровень логирования"`
	App      App
	Bot      tg_botx.Config
	Db       db.Config
	Images   Images
	Payment  Payment
	Auth     Auth
}

type App struct {
	AdminSecret string `schema:"secret"`
}

type Images struct {
	BaseImagePath   string
	BaseServicePath string
}

type Payment struct {
	ExpirationDelayMinutes int `validate:"required,gte=1"`
}

type Auth struct {
	Access                      JwtToken `schema:"secret"`
	Refresh                     JwtToken `schema:"secret"`
	TelegramExpireDurationHours int      `validate:"required,min=24"`
}

type JwtToken struct {
	TtlHours int    `validate:"required,min=24"`
	Secret   string `validate:"required,min=10"`
}
