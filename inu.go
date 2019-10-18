package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/dogbin/inu/dogbin"

	"github.com/urfave/cli"
)

var server string
var slug string
var file string
var jsonOutput bool

func main() {
	serverFlag := cli.StringFlag{
		Name:        "server, r",
		Usage:       "The dogbin/hastebin server to use",
		Value:       "del.dog",
		EnvVar:      "DOGBIN_SERVER",
		FilePath:    "~/.inu/server",
		Destination: &server,
	}
	slugFlag := cli.StringFlag{
		Name:        "slug, s",
		Usage:       "The slug to use instead of the server generated one [haste doesn't support this]",
		Destination: &slug,
	}
	fileFlag := cli.StringFlag{
		Name:        "file, f",
		Usage:       "A file to upload to dogbin",
		TakesFile:   true,
		Destination: &file,
	}
	jsonFlag := cli.BoolFlag{
		Name:        "json, j",
		Usage:       "Outputs the result as JSON",
		Destination: &jsonOutput,
	}

	app := cli.NewApp()
	app.Name = "inu"
	app.Usage = "Use dogbin/hastebin right from your terminal"
	app.Copyright = "(c) 2019 Till Kottmann"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Till Kottmann",
			Email: "me@deletescape.ch",
		},
	}
	app.Version = "0.0.1"
	app.EnableBashCompletion = true
	app.Action = put
	app.Flags = []cli.Flag{
		serverFlag,
		slugFlag,
		fileFlag,
		jsonFlag,
	}
	app.Commands = []cli.Command{
		{
			Name:    "put",
			Aliases: []string{"up", "p", "u", ""},
			Usage:   "Create a new paste",
			Action:  put,
			Flags: []cli.Flag{
				serverFlag,
				slugFlag,
				fileFlag,
				jsonFlag,
			},
		},
		{
			Name:    "get",
			Aliases: []string{"show", "s"},
			Usage:   "Obtains the contents of a paste",
			Action:  get,
			Flags: []cli.Flag{
				serverFlag,
				cli.StringFlag{
					Name:        "slug, s",
					Usage:       "The slug of the paste to retrieve",
					Destination: &slug,
				},
				jsonFlag,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
	}
}

func put(c *cli.Context) error {
	info, _ := os.Stdin.Stat()

	var content string
	if info.Mode()&os.ModeNamedPipe != 0 {
		content = readStdin()
		if c.NArg() == 1 {
			slug = c.Args()[0]
		}
	} else if file != "" {
		buf, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		content = string(buf)
		if c.NArg() == 1 {
			slug = c.Args()[0]
		}
	} else {
		if c.NArg() == 1 {
			content = c.Args()[0]
		} else if c.NArg() == 2 {
			slug = c.Args()[0]
			content = c.Args()[1]
		}
	}

	result, err := dogbin.New(server).Put(slug, content)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")

		return enc.Encode(result)
	} else {
		fmt.Println(result.Url)
	}
	return nil
}

func get(c *cli.Context) error {

	if c.NArg() == 1 {
		slug = c.Args()[0]
	}

	if slug == "" {
		return cli.ShowCommandHelp(c, "get")
	}

	var tmp = slug

	if strings.ContainsRune(tmp, '/') {
		// convert slug to url to attempt to extract path + server from it
		if !strings.HasPrefix(tmp, "http") && !strings.HasPrefix(tmp, "/") {
			tmp = "https://" + tmp
		}
		u, err := url.Parse(tmp)
		if err == nil {
			if path := u.Path[1:]; path != "" {
				tmp = path
			}
			u.Path = ""
			u.RawQuery = ""
			u.RawPath = ""
			u.Fragment = ""
			srv := u.String()
			if srv != "" {
				server = srv
			}
		}
	}

	if strings.ContainsRune(tmp, '.') {
		tmp = strings.SplitN(tmp, ".", 2)[0]
	}

	doc, err := dogbin.New(server).Get(tmp)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")

		return enc.Encode(doc)
	} else {
		fmt.Println(doc.Content)
	}
	return nil
}

func readStdin() string {
	reader := bufio.NewReader(os.Stdin)
	var input []rune

	for {
		ch, _, err := reader.ReadRune()
		if err != nil && err == io.EOF {
			break
		}
		input = append(input, ch)
	}

	return string(input)
}
