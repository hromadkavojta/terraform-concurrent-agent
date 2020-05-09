package main

import (
	"github.com/hromadkavojta/terraform-concurrent-agent/agent"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/spf13/viper"
	"log"
	"os/exec"
	"strings"
)

func main() {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetDefault("port", "8080")
	viper.SetDefault("google.cloud.project", "vojtah-sandbox")
	viper.SetDefault("SOURCE.OWNER", "hromadkavojta")
	viper.SetDefault("SOURCE.REPOT", "BP-infratest")
	viper.SetDefault("COMMIT.BRANCH", "master")
	viper.SetDefault("BASE.BRANCH", "master")
	viper.SetDefault("ACCESS.TOKEN", "ba178473c29d33eff42d314f5cd47ff1c84977c9")
	viper.SetDefault("GIT.URL", "git@github.com:hromadkavojta/BP-infratest.git")

	//ctx := context.Background()
	//ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "ba178473c29d33eff42d314f5cd47ff1c84977c9"})
	//tc := oauth2.NewClient(ctx, ts)

	svc := agent.NewService()

	//r, err := git.PlainClone(".", false, &git.CloneOptions{
	//	Auth: &http.BasicAuth{
	//		Username: "randomstring",
	//		Password: viper.GetString("ACCESS_TOKEN"),
	//	},
	//	URL: viper.GetString("GIT.URL"),
	//
	//})
	//if err != nil {
	//	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	//	os.Exit(1)
	//}
	//
	//svc.Repo = r

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
