package main

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
	viper.SetDefault("port", "8080")
	viper.SetDefault("google.cloud.project", "vojtah-sandbox")

	svc := agent.NewService()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/terraformplan", svc.TerraformPlan)
	e.GET("/terraformshow", svc.TerraformShow)
	e.GET("/terraformapply", svc.TerraformApply)

	e.Logger.Fatal(e.Start(":" + viper.GetString("port")))
}
