package main

import (
	"fmt"
	git "github.com/go-git/go-git/v5"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hromadkavojta/terraform-concurrent-agent/agent"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/spf13/viper"
	"os"
)

func main() {
	viper.AutomaticEnv()
	viper.SetDefault("port", "8080")
	viper.SetDefault("google_cloud_project", "vojtah-sandbox")
	viper.SetDefault("COMMITTER", "hromadkavojta")
	viper.SetDefault("COMMITTER_EMAIL", "hromadkavojta@gmail.com")
	viper.SetDefault("SOURCE_REPO", "BP-infratest")
	viper.SetDefault("COMMIT_BRANCH", "master")
	viper.SetDefault("BASE_BRANCH", "master")
	viper.SetDefault("ACCESS_TOKEN", "f542581834df185f66eff400afa289636fd10920")
	viper.SetDefault("GIT_URL", "https://github.com/hromadkavojta/BP-infratest")

	_, err := git.PlainClone(viper.GetString("SOURCE_REPO"), false, &git.CloneOptions{
		Auth: &githttp.BasicAuth{
			Username: "notNeeded",
			Password: viper.GetString("ACCESS_TOKEN"),
		},
		URL:      viper.GetString("GIT_URL"),
		Progress: os.Stdout,
	})
	if err != nil {
		fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	}

	serviceVariables := agent.ServiceVariables{
		Repo:           viper.GetString("SOURCE_REPO"),
		Url:            viper.GetString("GIT_URL"),
		AccessToken:    viper.GetString("ACCESS_TOKEN"),
		Committer:      viper.GetString("COMMITTER"),
		CommitterEmail: viper.GetString("COMMITTER_EMAIL"),
	}

	svc := agent.NewService(serviceVariables)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/terraformplan", svc.TerraformPlan)
	e.GET("/terraformshow", svc.TerraformShow)
	e.POST("/terraformapply", svc.TerraformApply)

	e.Logger.Fatal(e.Start(":" + viper.GetString("port")))
}
