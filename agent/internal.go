package agent

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"log"
	"net/http"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
)

func NewService() *Service {
	return &Service{}
}

func remove(s [][]string, i int) [][]string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func readInputs(cmd *exec.Cmd) string {
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

	return tfString + tfErrString
}

func notContain(processing [][]string, planned []string) bool {
	plannedLen := len(planned) - 1
	for _, resourceStruct := range processing {
		for _, resource := range resourceStruct {
			if resource == planned[plannedLen] {
				return false
			}
		}
	}
	return true
}

//TODO this should be called on github webook ( make terraform plan and save it )
func (s *Service) TerraformPlan(c echo.Context) error {

	//Cloning infra repository to agents file system
	clone := exec.Command("git", "-C", "BP-infratest", "pull", "git@github.com:hromadkavojta/BP-infratest.git")
	err := clone.Run()
	if err != nil {
		log.Printf("Couldn't pull github repo")
		return c.JSON(http.StatusNotFound, "Couldn't fetch github repo")
	}

	//Initializes terraforming
	tfInit := exec.Command("terraform", "init", "-lock=false", "BP-infratest")
	err = tfInit.Run()
	if err != nil {
		log.Printf("Terraform couldnt run init")
		return c.JSON(http.StatusNotFound, "Couldnt init terraform")
	}

	//Creates plan with last github version
	cmd := exec.Command("terraform", "plan", "-no-color", "lock=false", "-out=version"+strconv.Itoa(s.PlansProvided)+".out", "BP-infratest")
	s.PlansProvided++

	//outputs stderr+out
	output := readInputs(cmd)
	print(output)

	//Finding all afected resources by given version of plan
	re := regexp.MustCompile(`#[ ?]([^\s]+)`)
	s.planned = append(s.planned, re.FindAllString(output, -1))

	err = cmd.Wait()
	if err != nil {
		log.Print(err)
	}

	//Compresses to json to send it over API to client application
	jsonTfString, err := json.Marshal(output)
	return c.JSON(http.StatusOK, jsonTfString)
}

//TODO this should send client info about new terraform plan ( infrastructure config)
func (s *Service) TerraformShow(c echo.Context) error {

	return c.JSON(http.StatusOK, nil)
}

//TODO this should apply terraform config from given plan
func (s *Service) TerraformApply(c echo.Context) error {

	//Binds version of plan user wants to use
	var t ApplyStruct
	err := c.Bind(&t)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	//Workaround about plan version number
	re := regexp.MustCompile(`\d+`)
	planNumArr := re.FindString(t.Plan)
	planNum, err := strconv.Atoi(planNumArr)
	if err != nil {
		panic(err)
	}

	//Right before applying new version of infrastructure, we store which
	s.processing = append(s.processing, s.planned[planNum])
	sliceLen := len(s.processing)
	sliceLen--

	fmt.Printf("%+v\n", s.planned)
	fmt.Printf("%+v\n", s.processing)

	//applying infrastructure

	if reflect.DeepEqual(s.processing, s.planned[planNum]) {
		cmd := exec.Command("terraform", "apply", "lock=false", t.Plan)

		//outputs stderr+out
		output := readInputs(cmd)
		print(output)

		//Wait for command to finish
		err = cmd.Wait()
		if err != nil {
			log.Print(err)
		}
		s.processing = remove(s.processing, sliceLen)
		fmt.Printf("%+v\n", s.processing)

	} else if notContain(s.processing, s.planned[planNum]) {

		cmd := exec.Command("terraform", "apply", "lock=false", t.Plan)

		//outputs stderr+out
		output := readInputs(cmd)
		print(output)

		//Wait for command to finish
		err = cmd.Wait()
		if err != nil {
			log.Print(err)
		}
		s.processing = remove(s.processing, sliceLen)
		fmt.Printf("%+v\n", s.processing)
	} else {
		return c.JSON(http.StatusConflict, "This plan has colisions with actual plan which is being executed right now, please wait")
	}

	return c.JSON(http.StatusOK, nil)
}
