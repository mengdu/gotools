package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gookit/color"
	"github.com/mengdu/gotools/library/dirutil"
	"github.com/urfave/cli/v2"
)

func runAction(a string, b string, ignore string, onlyFile bool) error {
	apath, err := filepath.Abs(a)
	if err != nil {
		return err
	}

	bpath, err := filepath.Abs(b)
	if err != nil {
		return err
	}

	fmt.Println("diff", apath, bpath)
	if ignore != "" {
		fmt.Println("Ignore:", color.Yellow.Sprint(ignore))
	}
	arr, err := dirutil.Diff(a, b, ignore, onlyFile)

	if err != nil {
		return err
	}

	fmt.Println("Changed:", color.Yellow.Sprint(len(arr)))

	for _, v := range arr {
		line := ""
		if v.Type == dirutil.CHANGE_REMOVE {
			line = color.Red.Sprint("- " + v.File.RelativePath)
		} else if v.Type == dirutil.CHANGE_CHANGE {
			line = color.Yellow.Sprint("* " + v.File.RelativePath)
		} else {
			line = color.Green.Sprint("+ " + v.File.RelativePath)
		}

		fmt.Println(line)
	}

	return nil
}

func main() {
	app := &cli.App{
		Name:  "difdiff",
		Usage: "Hello world",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "onlyFile",
				Usage: "Only diff file",
			},
			&cli.StringFlag{
				Name:  "ignore",
				Usage: "Ignore file",
			},
		},
		Action: func(c *cli.Context) error {
			return runAction(c.Args().Get(0), c.Args().Get(1), c.String("ignore"), c.Bool("onlyFile"))
		},
	}

	err := app.Run(os.Args)

	if err != nil {
		log.Fatal(err)
	}
}
