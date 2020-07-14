package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
)

var Url string
var Hash string

func main() {

	viper.AutomaticEnv()

	app := &cli.App{
		Name:  "Terraform agent CLI",
		Usage: "Control client for terraform agent in your cloud",
	}

	app.Commands = []cli.Command{
		{
			Name:    "show",
			Aliases: []string{"s"},
			Usage:   "Shows terraform plan for next infastructure changes",
			Action: func(c *cli.Context) error {
				var bearer = "Bearer " + Hash

				req, err := http.NewRequest("GET", Url+"terraformshow", nil)

				req.Header.Add("Authorization", bearer)

				client := &http.Client{}

				resp, err := client.Do(req)

				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)

				var responseString string
				err = json.Unmarshal(body, &responseString)
				if err != nil {
					fmt.Printf("Bad Authorization")
					return nil
				}
				fmt.Printf("%+v\n", responseString)
				return nil
			},
		},
		{
			Name:    "apply",
			Aliases: []string{"a"},
			Usage:   "Applies last terraform plans",
			Action: func(c *cli.Context) error {
				fmt.Printf("Applying infrastructure plan\n This might take a while...\n")

				var bearer = "Bearer " + Hash
				req, err := http.NewRequest("GET", Url+"terraformshow", nil)
				req.Header.Add("Authorization", bearer)

				u := url.URL{Scheme: "ws", Host: strings.Trim(Url, "https://"), Path: "terraformapply"}
				fmt.Printf("connecting to server\n")

				conn, _, err := websocket.DefaultDialer.Dial(u.String(), req.Header)
				if err != nil {
					fmt.Printf("Not authorized!\n")
					return nil
				}
				defer conn.Close()

				for {
					_, message, err := conn.ReadMessage()
					if bytes.Compare(message, []byte("\n\r")) == 0 {
						break
					}
					if err != nil {
						log.Println("read:", err)
					}
					fmt.Printf("%s", message)
				}
				return nil
			},
		},
		{
			Name:    "plan",
			Aliases: []string{"p"},
			Usage:   "Creates new terraform plan fom master repo on github",
			Action: func(c *cli.Context) error {

				var bearer = "Bearer " + Hash

				req, err := http.NewRequest("GET", Url+"terraformplan", nil)

				req.Header.Add("Authorization", bearer)

				client := &http.Client{}

				fmt.Printf("Creating infrastructure plan\n Please wait a moment...\n")
				resp, err := client.Do(req)

				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)

				var responseString string
				err = json.Unmarshal(body, &responseString)
				if err != nil {
					fmt.Printf("Bad Authorization")
					return nil
				}
				fmt.Printf(responseString)
				return nil
			},
		},
	}

	app.Action = func(c *cli.Context) error {
		println(c.NArg())
		return nil
	}

	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
