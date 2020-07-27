package agent

import (
	"bufio"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
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
	s.wg.Add(1)
	defer s.wg.Done()
	//Cloning infra repository to agents file system

	var err error

	gitPull := exec.Command("git", "-C", s.repo, "pull", s.url)
	output, _ := readInputs(gitPull)
	println(output)

	//Initializes terraforming
	tfInit := exec.Command("terraform", "init", s.repo)
	output, errString := readInputs(tfInit)
	if errString != "" {
		fmt.Printf("%+v", errString)
		f, err := os.Create(s.repo + "/infrastructure_changes")
		if err != nil {
			panic(err)
		}
		_, err = f.Write([]byte(errString))
		if err != nil {
			panic(err)
		}
		return c.JSON(http.StatusOK, errString)
	}

	//Creates plan with last github version
	cmd := exec.Command("terraform", "plan", "-no-color", "-out=plan.out", s.repo)

	//outputs stderr+out
	output, errString = readInputs(cmd)
	if errString != "" {
		fmt.Printf("%+v", errString)
		f, err := os.Create(s.repo + "/infrastructure_changes")
		if err != nil {
			panic(err)
		}
		_, err = f.Write([]byte(errString))
		if err != nil {
			panic(err)
		}
		return c.JSON(http.StatusOK, errString)
	}

	f, err := os.Create(s.repo + "/infrastructure_changes")
	if err != nil {
		panic(err)
	}
	trimOutput := strings.Split(output, "------------------------------------------------------------------------")
	_, err = f.Write([]byte(trimOutput[1]))
	if err != nil {
		panic(err)
	}

	//Compresses to json to send it over API to client application
	return c.JSON(http.StatusOK, "planning succesfully finished, you can take a look on planned changes with ./tfagent show\n")
}

func (s *Service) TerraformShow(c echo.Context) error {

	s.wg.Wait()

	if _, err := os.Stat("plan.out"); err != nil {
		return c.JSON(http.StatusForbidden, "At this moment doesnt exist any plan to apply!\n")
	}

	if _, err := os.Stat(s.repo + "/infrastructure_changes"); err != nil {
		return c.JSON(http.StatusForbidden, "At this moment doesnt exist any plan to apply!\n")
	}

	dat, err := ioutil.ReadFile(s.repo + "/infrastructure_changes")
	if err != nil {
		panic(err)
	}

	return c.JSON(http.StatusOK, string(dat))

}

func (s *Service) TerraformApply(c echo.Context) error {
	//Binds version of plan user wants to use
	s.wg.Add(1)
	defer s.wg.Done()

	var (
		upgrader = websocket.Upgrader{}
	)

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	//applying infrastructure
	if _, err := os.Stat("plan.out"); err != nil {
		err = ws.WriteMessage(websocket.TextMessage, []byte("Something had to go wrong, there is no plan to apply\n Please check if you have plan before you apply!\n"))

		err = ws.WriteMessage(websocket.TextMessage, []byte("\n\r"))
		if err != nil {
			c.Logger().Error(err)
		}
	}

	cmd := exec.Command("terraform", "apply", "-no-color", "plan.out")
	//outputs stderr+out

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(stdout)
	line, err := reader.ReadString('\n')
	for err == nil {
		line, err = reader.ReadString('\n')
		err := ws.WriteMessage(websocket.TextMessage, []byte(line))
		if err != nil {
			c.Logger().Error(err)
		}
	}

	err = os.Remove("plan.out")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	gitPush := exec.Command("git", "-C", s.repo, "pull", s.url)
	output, _ := readInputs(gitPush)
	println(output)

	r, err := git.PlainOpen(s.repo)
	checkError(err)

	w, err := r.Worktree()
	checkError(err)

	_, err = w.Add("infrastructure_changes")
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
		RemoteName: "push",
		Auth: &githttp.BasicAuth{
			Username: "notNeeded",
			Password: viper.GetString("ACCESS_TOKEN"),
		},
	})
	checkError(err)

	err = ws.WriteMessage(websocket.TextMessage, []byte("Successfully finished logging on git\n"))

	err = ws.WriteMessage(websocket.TextMessage, []byte("\n\r"))
	if err != nil {
		c.Logger().Error(err)
	}

	return nil

}
