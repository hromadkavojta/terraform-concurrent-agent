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
)

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

	err = cmd.Wait()
	if err != nil {
		log.Print(err)
	}

	return tfString + tfErrString, tfErrString
}

//TODO this should be called on github webook ( make terraform plan and save it )
func (s *Service) TerraformPlan(c echo.Context) error {

	//Waits for apply to finish, in case it's running some configuration at that moment
	s.wg.Wait()
	//Cloning infra repository to agents file system
	gitpull := exec.Command("git", "-C", s.repo, "pull", s.url)
	err := gitpull.Run()
	if err != nil {
		fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
		return c.JSON(http.StatusNotFound, "Couldn't fetch github repo")
	}

	//Initializes terraforming
	tfInit := exec.Command("terraform", "init", s.repo)
	err = tfInit.Run()
	if err != nil {
		fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
		return c.JSON(http.StatusNotFound, "Couldnt init terraform")
	}

	//Creates plan with last github version
	cmd := exec.Command("terraform", "plan", "-no-color", "-out=plan.out", s.repo)
	s.PlansProvided++

	//outputs stderr+out
	output, errString := readInputs(cmd)
	if errString != "" {
		fmt.Printf("%+v", errString)
	}

	f, err := os.Create(s.repo + "/planned_infra")
	if err != nil {
		panic(err)
	}
	_, err = f.Write([]byte(output))
	if err != nil {
		panic(err)
	}

	//Compresses to json to send it over API to client application
	return c.JSON(http.StatusOK, "planning succesfully finished, you can take a look with ./tfagent show")
}

//TODO this should send client info about new terraform plan ( infrastructure config)
func (s *Service) TerraformShow(c echo.Context) error {

	if _, err := os.Stat(s.repo + "/planned_infra"); err != nil {
		return c.JSON(http.StatusForbidden, "At this moment doesnt exist any plan to apply!")
	}
	dat, err := ioutil.ReadFile(s.repo + "/planned_infra")
	if err != nil {
		panic(err)
	}

	fmt.Printf(string(dat))
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
	if _, err = os.Stat("plan.out"); err != nil {
		s.wg.Done()
		return c.JSON(http.StatusForbidden, "Something had to go wrong, there is no plan to apply")
	}

	cmd := exec.Command("terraform", "apply", "-no-color", "plan.out")
	//outputs stderr+out
	output, errString := readInputs(cmd)
	fmt.Printf(errString)
	if errString != "" {
		s.wg.Done()
		return c.JSON(http.StatusAccepted, "There occured error during applying infrastrucure, please check this log and fix your problems"+errString)
	}
	print(output)

	err = os.Remove("plan.out")
	if err != nil {
		s.wg.Done()
		c.JSON(http.StatusInternalServerError, err)
	}

	gitCmd1 := exec.Command("git", "-C", s.repo, "add", "planned_infra")
	output, errString = readInputs(gitCmd1)
	print(output)

	gitCmd2 := exec.Command("git", "-C", s.repo, "commit", "-m'Last infra changelog'")
	output, errString = readInputs(gitCmd2)
	print(output)

	gitCmd3 := exec.Command("git", "-C", s.repo, "pull", s.url)
	output, errString = readInputs(gitCmd3)
	print(output)

	gitCmd4 := exec.Command("git", "-C", s.repo, "push", "origin", "master")
	output, errString = readInputs(gitCmd4)
	print(output)

	s.wg.Done()
	return c.JSON(http.StatusOK, output)
}
