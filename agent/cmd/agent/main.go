package main

import (
	"fmt"
	"github.com/hromadkavojta/terraform-concurrent-agent/agent"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/spf13/viper"
	"os/exec"
)

func main() {
	viper.AutomaticEnv()
	viper.SetDefault("port", "8080")
	viper.SetDefault("COMMITTER", "hromadkavojta")
	viper.SetDefault("COMMITTER_EMAIL", "hromadkavojta@gmail.com")
	viper.SetDefault("SOURCE_REPO", "BP-infratest")
	viper.SetDefault("ACCESS_TOKEN", "")
	viper.SetDefault("GIT_URL_HTTPS", "https://github.com/hromadkavojta/BP-infratest.git")
	viper.SetDefault("GIT_URL_SSH", "git@github.com:hromadkavojta/BP-infratest.git")

	serviceVariables := agent.ServiceVariables{
		Repo:           viper.GetString("SOURCE_REPO"),
		Url:            viper.GetString("GIT_URL_SSH"),
		AccessToken:    viper.GetString("ACCESS_TOKEN"),
		Committer:      viper.GetString("COMMITTER"),
		CommitterEmail: viper.GetString("COMMITTER_EMAIL"),
	}

	svc := agent.NewService(serviceVariables)

	clone := exec.Command("git", "clone", viper.GetString("GIT_URL_SSH"))
	err := clone.Run()
	if err != nil {
		fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error which is here: %s", err))
	}

	remoteAdd := exec.Command("git", "-C", viper.GetString("SOURCE_REPO"), "remote", "add", "push", viper.GetString("GIT_URL_HTTPS"))
	err = remoteAdd.Run()
	if err != nil {
		fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error which is here: %s", err))
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/terraformplan", svc.TerraformPlan)
	e.GET("/terraformshow", svc.TerraformShow)
	e.GET("/terraformapply", svc.TerraformApply)

	e.Logger.Fatal(e.Start(":" + viper.GetString("port")))
}
