package agent

import (
	"bufio"
	"encoding/json"
	"github.com/labstack/echo"
	"go/types"
	"log"
	"net/http"
	"os/exec"
)

type Service types.Struct

func NewService() *Service {
	return &Service{}
}

//TODO this should be called on github webook ( make terraform plan and save it )
func (s *Service) TerraformPlan(c echo.Context) error {

	//Cloning infra repo
	clone := exec.Command("git", "clone", "git@github.com:hromadkavojta/BP-infratest.git")
	err := clone.Run()
	if err != nil {
		log.Printf("Couldn't fetch github repo")
		return c.JSON(http.StatusNotFound, "Couldn't fetch github repo")
	}

	tfInit := exec.Command("terraform", "init", "BP-infratest")
	err = tfInit.Run()
	if err != nil {
		log.Printf("Terraform couldnt run init")
		return c.JSON(http.StatusNotFound, "Couldnt init terraform")
	}

	cmd := exec.Command("terraform", "plan", "-no-color", "-out=plan.out", "BP-infratest")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Print(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Print(err)
	}

	err = cmd.Start()
	if err != nil {
		log.Print(err)
		log.Println("2")
	}

	var tf_string string
	reader := bufio.NewReader(stdout)
	line, err := reader.ReadString('\n')
	for err == nil {
		tf_string = tf_string + line
		line, err = reader.ReadString('\n')
	}

	var tf_err_string string
	reader = bufio.NewReader(stderr)
	line, err = reader.ReadString('\n')
	for err == nil {
		tf_err_string = tf_err_string + line
		line, err = reader.ReadString('\n')
	}

	print(tf_string)
	print(tf_err_string)

	err = cmd.Wait()
	if err != nil {
		log.Print(err)
		log.Println("4")
	}

	//REMOVING CLONED DIRECTORY
	rm_dir := exec.Command("rm", "-r", "BP-infratest")
	err = rm_dir.Run()
	if err != nil {
		log.Printf("Couldnt remove directory %v", err)
		return c.JSON(http.StatusForbidden, nil)
	}

	jsonTfString, err := json.Marshal(tf_string)
	return c.JSON(http.StatusOK, jsonTfString)
}

//TODO this should send client info about new terraform plan ( infrastructure config)
func (s *Service) TerraformShow(c echo.Context) error {

	return c.JSON(http.StatusOK, nil)
}

//TODO this should apply terraform config from given plan
func (s *Service) TerraformApply(c echo.Context) error {

	return c.JSON(http.StatusOK, nil)
}
