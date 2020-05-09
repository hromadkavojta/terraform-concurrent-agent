package agent

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
)

func NewService() *Service {
	return &Service{}
}

func readInputs(cmd *exec.Cmd) (string, string) {
	//Reading stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Print(err)
	}

	//Reading stderr
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Print(err)
	}

	//Running actual command
	err = cmd.Start()
	if err != nil {
		log.Print(err)
	}

	//Reading appending whole stdout to one variable
	var tfString string
	reader := bufio.NewReader(stdout)
	line, err := reader.ReadString('\n')
	for err == nil {
		tfString = tfString + line
		line, err = reader.ReadString('\n')
	}

	//Reading appending whole stderr to one variable
	var tfErrString string
	reader = bufio.NewReader(stderr)
	line, err = reader.ReadString('\n')
	for err == nil {
		tfErrString = tfErrString + line
		line, err = reader.ReadString('\n')
	}

	return tfString + tfErrString, tfErrString
}

//TODO this should be called on github webook ( make terraform plan and save it )
func (s *Service) TerraformPlan(c echo.Context) error {

	//Waits for apply to finish, in case it's running some configuration at that moment
	s.wg.Wait()
	//Cloning infra repository to agents file system
	clone := exec.Command("git", "-C", "infrastructure", "pull", "git@github.com:hromadkavojta/BP-infratest.git")
	err := clone.Run()
	if err != nil {
		log.Printf("Couldn't pull github repo")
		return c.JSON(http.StatusNotFound, "Couldn't fetch github repo")
	}

	//Initializes terraforming
	tfInit := exec.Command("terraform", "init", "infrastructure")
	err = tfInit.Run()
	if err != nil {
		log.Printf("Terraform couldnt run init")
		return c.JSON(http.StatusNotFound, "Couldnt init terraform")
	}

	//Creates plan with last github version
	cmd := exec.Command("terraform", "plan", "-no-color", "-out=version"+strconv.Itoa(s.PlansProvided)+".out", "infrastructure")
	s.PlansProvided++

	//outputs stderr+out
	output, errString := readInputs(cmd)
	if errString != "" {
		fmt.Printf("%+v", errString)
	}

	err = cmd.Wait()
	if err != nil {
		log.Print(err)
	}

	f, err := os.Create("planned_infra")
	if err != nil {
		panic(err)
	}
	_, err = f.Write([]byte(output))
	if err != nil {
		panic(err)
	}

	//Compresses to json to send it over API to client application
	return c.JSON(http.StatusOK, "planning succesfully finished")
}

//TODO this should send client info about new terraform plan ( infrastructure config)
func (s *Service) TerraformShow(c echo.Context) error {

	dat, err := ioutil.ReadFile("planned_infra")
	if err != nil {
		panic(err)
	}

	jsonTfString, err := json.Marshal(dat)
	return c.JSON(http.StatusOK, jsonTfString)

}

//TODO this should apply terraform config from given plan
func (s *Service) TerraformApply(c echo.Context) error {

	//Binds version of plan user wants to use
	s.wg.Add(1)
	var t ApplyStruct
	err := c.Bind(&t)
	if err != nil {
		s.wg.Done()
		return c.JSON(http.StatusBadRequest, err)
	}

	//applying infrastructure
	cmd := exec.Command("terraform", "apply", t.Plan)

	//outputs stderr+out
	output, errString := readInputs(cmd)
	if errString != "" {
		return c.JSON(http.StatusAccepted, "There occured error during applying infrastrucure, please check this log and fix your problems")
	}

	print(output)

	//Wait for command to finish
	err = cmd.Wait()
	if err != nil {
		log.Print(err)
	}

	s.wg.Done()
	return c.JSON(http.StatusOK, "Plan successfuly applied on your cloud environment, you can see the changelog on your github")
}
