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
	"github.com/xuri/excelize/v2"
)

// Functionality to be added
// 1. Convert add students and add records into sub actions
// 2. Add students
// 3. Allow user to specify which columns contain what values

var env map[string]string

func main() {
	var err error
	env, err = godotenv.Read(".env")
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

	insertFlags := []cli.Flag{
		&cli.StringFlag{
			Name: "semester",
			Value: "",
		},
		&cli.StringFlag{
			Name: "subjectcode",
			Value: "",
		},
		&cli.BoolFlag{
			Name: "createsubject",
			Value: false,
		},
	}

	registerFlags := []cli.Flag{
		&cli.StringFlag{
			Name: "degree",
			Value: "",
			Required: true,
		},
	}

	app.Commands = []*cli.Command{
		{
			Name: "login",
			Aliases: []string{"l"},
			Usage: "Used to login",
			UsageText: "Other text",
			Action: handleLogin,
		},
		{
			Name: "insert",
			Aliases: []string{"i"},
			Usage: "Inserting records or subjects", 
			Subcommands: []*cli.Command{
				{
					Name: "rec",
					Aliases: []string{"bi"},
					Usage: "Insert Records from .xlsx file",
					Flags: insertFlags,
					Action: handlebInsert,
				},
				{
					Name: "sub",
					Aliases: []string{"si"},
					Usage: "Insert subject into database",
					Action: handlesInsert,
				},
			},
		},
		{
			Name: "register",
			Aliases: []string{"r"},
			Usage: "Register users from .xlsx file",
			Flags: registerFlags,
			Action: registerUsers,
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

func handleLogin(ctx *cli.Context) error {
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
	res, err := http.Post(env["WEBSITE"]+"/authorize/admin", "application/json", responseBody)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	var data map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return err
	}
	msg, ok := data["msg"].(string)
	if ok {
		return fmt.Errorf(strings.ToLower(msg))
	}
	env["TOKEN"] = data["token"].(string)
	godotenv.Write(env, ".env")
	fmt.Printf("Authenticated. ")
	fmt.Println("The login expires after 10 hours!")
	return nil
}

func handlebInsert(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return fmt.Errorf("please provide file name")
	}
	fileName := string(ctx.Args().Get(0))
	sheetName := string(ctx.Args().Get(1))
	f, err := excelize.OpenFile(fileName)
	if err != nil {
		return fmt.Errorf("could not open .xlsx file")
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println("File not closed")
		}
	}()
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("could not read .xlsx file")
	}
	for idx, row := range rows {
		postBody, err := json.Marshal(map[string]string {
			"entryNumber": row[0],
			"grade": row[1],
			"subjectCode": func() string {
				if len(ctx.String("subjectcode")) == 0 {
					return row[2]
				} else {
					return ctx.String("subjectcode")
				}
			}(),
			"semester": func() string {
				if len(ctx.String("semester")) == 0 {
					return row[3]
				} else {
					return ctx.String("semester")
				}
			}(),
		})
		if err != nil {
			f.SetCellValue(sheetName, "E"+fmt.Sprint(idx+1), "Failed")
			return fmt.Errorf("could not build json body")
		}
		responseBody := bytes.NewBuffer(postBody)
		client := &http.Client{}
		req, err := http.NewRequest("POST", env["WEBSITE"]+"/admin/records/single", responseBody)
		if err != nil {
			f.SetCellValue(sheetName, "E"+fmt.Sprint(idx+1), "Failed")
			return fmt.Errorf("request formation failed")
		}
		req.Header.Add("Authorization", "Bearer " + env["TOKEN"])
		req.Header.Set("Content-Type", "application/json")
		res, err := client.Do(req)
		if err != nil {
			f.SetCellValue(sheetName, "E"+fmt.Sprint(idx+1), "Failed")
			return fmt.Errorf("request failed")
		}
		defer res.Body.Close()
		var data map[string]interface{}
		err = json.NewDecoder(res.Body).Decode(&data)
		if err != nil {
			f.SetCellValue(sheetName, "E"+fmt.Sprint(idx+1), "Failed")
			return fmt.Errorf("could not decode response")
		}
		f.SetCellValue(sheetName, "E"+fmt.Sprint(idx+1), data["msg"])
	}
	f.SaveAs(fileName)
	return nil
}

func handlesInsert(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return fmt.Errorf("enter valid subject")
	}
	postBody, err := json.Marshal(map[string]string {
		"subjectCode": ctx.Args().Get(0),
		"subjectName": ctx.Args().Get(1),
	})
	if err != nil {
		return fmt.Errorf("could not create json object")
	}
	responseBody := bytes.NewBuffer(postBody)
	client := &http.Client{}
	req, err := http.NewRequest("POST", env["WEBSITE"]+"/admin/records", responseBody)
	if err != nil {
		return fmt.Errorf("request formulation failed")
	}
	req.Header.Add("Authorization", "Bearer " + env["TOKEN"])
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed")
	}
	defer res.Body.Close()
	var data map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return fmt.Errorf("could not decode response")
	}
	fmt.Println(data["msg"])
	return nil
}

func registerUsers(ctx *cli.Context) error {
	if ctx.NArg() != 2 || len(ctx.String("degree")) == 0 {
		return fmt.Errorf("enter valid arguments to function")
	}
	if ctx.String("degree") != "B.Tech" && ctx.String("degree") != "M.Tech" && ctx.String("degree") != "PhD" {
		return fmt.Errorf("invalid degree")
	}
	fileName, sheetName := ctx.Args().Get(0), ctx.Args().Get(1)
	f, err := excelize.OpenFile(fileName)
	if err != nil {
		return fmt.Errorf("cannot open input file")
	}
	defer f.Close()
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("cannot read rows from the sheet")
	}
	for idx, row := range rows {
		postBody, err := json.Marshal(map[string]string {
			"name": row[0],
			"entryNumber": row[1],
			"degree": ctx.String("degree"),
		})
		if err != nil {
			f.SetCellValue(sheetName, "C"+fmt.Sprint(idx+1), "Not inserted")
			return fmt.Errorf("could not formulate json")
		}
		responseBody := bytes.NewBuffer(postBody)
		client := &http.Client{}
		req, err := http.NewRequest("POST", env["WEBSITE"] + "/admin/register/user", responseBody)
		if err != nil {
			f.SetCellValue(sheetName, "C"+fmt.Sprint(idx+1), "Not inserted")
			return fmt.Errorf("could not formulate request")
		}
		req.Header.Add("Authorization", "Bearer " + env["TOKEN"])
		req.Header.Set("Content-Type", "application/json")
		res, err := client.Do(req)
		if err != nil {
			f.SetCellValue(sheetName, "C"+fmt.Sprint(idx+1), "Not inserted")
			return fmt.Errorf("request failed")
		}
		defer res.Body.Close()
		var data map[string]interface{}
		err = json.NewDecoder(res.Body).Decode(&data)
		if err != nil {
			f.SetCellValue(sheetName, "C"+fmt.Sprint(idx+1), "Not inserted")
			return fmt.Errorf("could not decode response")
		}
		f.SetCellValue(sheetName, "C"+fmt.Sprint(idx+1), data["msg"])
	}
	f.SaveAs(fileName)
	return nil
}