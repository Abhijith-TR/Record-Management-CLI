package main

import (
	"fmt"
	"os"
	// "log"
	// "os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "First CLI"
	app.Usage = "Testing"

	myFlags := []cli.Flag{
		&cli.StringFlag{
			Name: "name",
			Value: "Stranger",
		},
	}

	app.Commands = []*cli.Command{
		{
			Name: "print",
			Usage: "Prints hello world",
			Flags: myFlags,
			Action: func(c *cli.Context) error {
				fmt.Println("Hello", c.String("name"))
				return nil
			},
		},
	}

	app.Run(os.Args)
}