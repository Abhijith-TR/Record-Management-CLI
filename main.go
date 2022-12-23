package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"net/http"
	"bufio"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

func main() {
	env, err := godotenv.Read(".env")
	if err != nil {
		fmt.Println("Could not load .env")
		return
	}

	app := cli.NewApp()
	app.Name = "irms"
	app.Version = "1.0.0.0"
	app.Authors = []*cli.Author{
		{
			Name: "Abhijith TR",
			Email: "2020csb1062@iitrpr.ac.in",
		},
	}
	app.Usage = "Use IRMS Server through CLI"
	app.UsageText = "IRMS - CLI"
	app.EnableBashCompletion = true

	app.Commands = []*cli.Command{
		{
			Name: "login",
			Aliases: []string{"l"},
			Usage: "Used to login",
			UsageText: "Other text",
			Action: func(ctx *cli.Context) error {
				if ctx.NArg() == 0 {
					return fmt.Errorf("invalid username")
				}
				password, err := readPassword()
				if err != nil {
					fmt.Println(err.Error())
					return nil
				}
				postBody, err := json.Marshal(map[string]string {
					"email": ctx.Args().Get(0),
					"password": password,
				})
				if err != nil {
					return err
				}
				responseBody := bytes.NewBuffer(postBody)
				resp, err := http.Post(env["WEBSITE"]+"/authorize/admin", "application/json", responseBody)
				if err != nil {
					return err
				}
				defer resp.Body.Close()
				var data map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&data)
				if err != nil {
					return err
				}
				msg, ok := data["msg"].(string)
				if ok {
					return fmt.Errorf(msg)
				}
				env["TOKEN"] = data["token"].(string)
				godotenv.Write(env, ".env")
				fmt.Printf("Authenticated!")
				return nil
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err.Error())
	}
}

func readPassword() (string, error) {
	fmt.Print("Password: ")
	var password string
	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	password = strings.TrimSpace(password)
	if len(password) == 0 {
		return "", fmt.Errorf("empty password")
	}
	return password, err
}