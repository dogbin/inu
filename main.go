package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/dogbin/inu/dogbin"

	"github.com/atotto/clipboard"
	"github.com/urfave/cli"
)

const (
	appAuthorName  = "Till Kottmann"
	appAuthorEmail = "me@deletescape.ch"
	appCopyright   = "(c) 2019 " + appAuthorName
	appName        = "inu"
	appVersion     = "v0.1.3"
)

var (
	file            string
	server          string
	apiKey			string
	slug            string
	clipboardOutput bool
	jsonOutput      bool
)

func main() {
	fileFlag := cli.StringFlag{
		Name:        "file, f",
		Usage:       "A file to upload to dogbin",
		TakesFile:   true,
		Destination: &file,
	}
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
	jsonFlag := cli.BoolFlag{
		Name:        "json, j",
		Usage:       "Outputs the result as JSON",
		Destination: &jsonOutput,
	}
	clipboardFlag := cli.BoolFlag{
		Name:        "copy, c",
		Usage:       "Additionally puts the created URL in your clipboard",
		Destination: &clipboardOutput,
	}
	apiKeyFlag := cli.StringFlag{
		Name:        "key, k",
		Usage:       "The dogbin api key to use",
		Value:       "",
		EnvVar:      "DOGBIN_KEY",
		FilePath:    "~/.inu/key",
		Destination: &apiKey,
	}

	app := cli.NewApp()
	app.Name = appName
	app.Usage = "Use dogbin/hastebin right from your terminal"
	app.Copyright = appCopyright
	app.Authors = []cli.Author{
		{
			Name:  appAuthorName,
			Email: appAuthorEmail,
		},
	}
	app.Version = appVersion
	app.EnableBashCompletion = true
	app.Action = put
	app.Flags = []cli.Flag{
		serverFlag,
		slugFlag,
		fileFlag,
		jsonFlag,
		clipboardFlag,
		apiKeyFlag,
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
				clipboardFlag,
				apiKeyFlag,
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
				cli.BoolFlag{
					Name:        "copy, c",
					Usage:       "Additionally puts the retrieved content in your clipboard",
					Destination: &clipboardOutput,
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
	}
}

func put(c *cli.Context) error {
	info, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("unable to stat os.Stdin: %w", err)
	}

	var content string
	if info.Mode()&os.ModeNamedPipe != 0 {
		content = readStdin()
		if c.NArg() == 1 {
			slug = c.Args()[0]
		}
	} else if file != "" {
		buf, err := ioutil.ReadFile(file)
		if err != nil {
			return fmt.Errorf("unable to read the file '%s': %w", file, err)
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

	result, err := dogbin.NewServer(server, strings.TrimSpace(apiKey)).Put(slug, content)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	if clipboardOutput {
		if err = clipboard.WriteAll(result.Url); err != nil {
			return fmt.Errorf("unable to write the output into the clipboard: %w", err)
		}
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	fmt.Println(result.Url)
	return nil
}

func get(c *cli.Context) error {
	if c.NArg() == 1 {
		slug = c.Args()[0]
	}

	if slug == "" {
		return cli.ShowCommandHelp(c, "get")
	}

	pasteURL := slug
	if strings.ContainsRune(pasteURL, '/') {
		// convert slug to url to attempt to extract path + server from it
		if !strings.HasPrefix(pasteURL, "http") && !strings.HasPrefix(pasteURL, "/") {
			pasteURL = "https://" + pasteURL
		}
		u, err := url.Parse(pasteURL)
		if err == nil {
			if path := u.Path[1:]; path != "" {
				pasteURL = path
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

	if strings.ContainsRune(pasteURL, '.') {
		pasteURL = strings.SplitN(pasteURL, ".", 2)[0]
	}

	doc, err := dogbin.NewServer(server, strings.TrimSpace(apiKey)).Get(pasteURL)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	if clipboardOutput {
		if err := clipboard.WriteAll(doc.Content); err != nil {
			return err
		}
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")

		return enc.Encode(doc)
	}

	fmt.Println(doc.Content)
	return nil
}

func readStdin() string {
	var input []byte
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input = append(input, scanner.Bytes()...)
	}

	return string(input)
}
