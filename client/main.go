package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sort"
)

var Url string

func main() {

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
				resp, err := http.Get(Url + "terraformshow")
				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)

				var responseString string
				err = json.Unmarshal(body, &responseString)
				if err != nil {
					panic(err)
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

				interrupt := make(chan os.Signal, 1)
				signal.Notify(interrupt, os.Interrupt)

				u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/terraformapply"}
				log.Printf("connecting to server")

				conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
				if err != nil {
					conn.Close()
					log.Fatal("dial:", err)
				}
				defer conn.Close()

				for {
					_, message, err := conn.ReadMessage()
					if err != nil {
						log.Println("read:", err)
					}
					if bytes.Compare(message, []byte("\n\r")) == 0 {
						break
					}
					log.Printf("%s", message)
				}
				return nil
			},
		},
		{
			Name:    "plan",
			Aliases: []string{"p"},
			Usage:   "Creates new terraform plan fom master repo on github",
			Action: func(c *cli.Context) error {
				fmt.Printf("Creating infrastructure plan\n Please wait a moment...\n")
				resp, err := http.Get(Url + "terraformplan")
				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)

				var responseString string
				err = json.Unmarshal(body, &responseString)
				if err != nil {
					panic(err)
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
