package conf_test

import (
	"dishes-service-backend/conf"
	"github.com/Falokut/go-kit/test/rct"
	"testing"
)

func TestDefaultRemoteConfig(t *testing.T) {
	t.Parallel()
	rct.Test(t, "default_remote_config.json", conf.Remote{})
}
