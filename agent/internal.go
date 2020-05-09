package agent

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/labstack/echo"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

func checkError(err error) {
	if err == nil {
		return
	}
	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
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

	err = cmd.Wait()
	if err != nil {
		log.Print(err)
	}

	return tfString + tfErrString, tfErrString
}

func (s *Service) TerraformPlan(c echo.Context) error {

	//Waits for apply to finish, in case it's running some configuration at that moment
	s.wg.Wait()
	//Cloning infra repository to agents file system

	var err error
	r, err := git.PlainOpen("BP-infratest")
	checkError(err)

	w, err := r.Worktree()

	err = w.Pull(&git.PullOptions{
		RemoteName: "origin",
		Auth: &githttp.BasicAuth{
			Username: "notNeeded",
			Password: viper.GetString("ACCESS_TOKEN"),
		},
	})

	checkError(err)

	//Initializes terraforming
	tfInit := exec.Command("terraform", "init", s.repo)
	output, errString := readInputs(tfInit)
	println(errString)

	//Creates plan with last github version
	cmd := exec.Command("terraform", "plan", "-no-color", "-out=plan.out", s.repo)

	//outputs stderr+out
	output, errString = readInputs(cmd)
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

	r, err := git.PlainOpen("BP-infratest")
	checkError(err)

	w, err := r.Worktree()
	checkError(err)

	_, err = w.Add("planned_infra")
	checkError(err)

	err = w.Pull(&git.PullOptions{
		RemoteName: "origin",
		Auth: &githttp.BasicAuth{
			Username: "notNeeded",
			Password: viper.GetString("ACCESS_TOKEN"),
		},
	})
	checkError(err)

	_, err = w.Commit("last Infrastructure changes", &git.CommitOptions{
		Author: &object.Signature{
			Name:  s.committer,
			Email: s.committerEmail,
			When:  time.Now(),
		},
	})

	checkError(err)

	err = r.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth: &githttp.BasicAuth{
			Username: "notNeeded",
			Password: viper.GetString("ACCESS_TOKEN"),
		},
	})
	checkError(err)

	s.wg.Done()
	return c.JSON(http.StatusOK, output)
}
