package shopxo

import (
	"os"
	"testing"

	"github.com/zhangpanda/goshop/config"
	"github.com/zhangpanda/goshop/internal/app"
)

func TestMain(m *testing.M) {
	app.Register(&app.Deps{Cfg: &config.Config{}})
	os.Exit(m.Run())
}
