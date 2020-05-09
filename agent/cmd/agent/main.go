package main

import (
	"github.com/hromadkavojta/terraform-concurrent-agent/agent"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/spf13/viper"
	"log"
	"os/exec"
)

func main() {
	viper.AutomaticEnv()
	viper.SetDefault("port", "8080")
	viper.SetDefault("google_cloud_project", "vojtah-sandbox")
	viper.SetDefault("SOURCE_OWNER", "hromadkavojta")
	viper.SetDefault("SOURCE_REPO", "BP-infratest")
	viper.SetDefault("COMMIT_BRANCH", "master")
	viper.SetDefault("BASE_BRANCH", "master")
	viper.SetDefault("ACCESS_TOKEN", "")
	viper.SetDefault("GIT_URL", "git@github.com:hromadkavojta/BP-infratest.git")

	serviceVariables := agent.ServiceVariables{
		Repo: viper.GetString("SOURCE_REPO"),
		Url:  viper.GetString("GIT_URL"),
	}

	svc := agent.NewService(serviceVariables)

	clone := exec.Command("git", "clone", "git@github.com:hromadkavojta/BP-infratest.git")
	err := clone.Run()
	if err != nil {
		log.Printf("Couldn't fetch github repo %v", err)
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/terraformplan", svc.TerraformPlan)
	e.GET("/terraformshow", svc.TerraformShow)
	e.POST("/terraformapply", svc.TerraformApply)

	e.Logger.Fatal(e.Start(":" + viper.GetString("port")))
}
