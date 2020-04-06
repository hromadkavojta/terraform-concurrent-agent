package agent

import (
	"github.com/hromadkavojta/terraform-concurrent-agent/agent"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/spf13/viper"
	"strings"
)

func main() {

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("port", "80")
	viper.SetDefault("google.cloud.project", "vojtah-sandbox")

	svc := agent.NewService()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/newconfiguration", svc.TerraformIt)

	e.Logger.Fatal(e.Start(":" + viper.GetString("port")))
}
