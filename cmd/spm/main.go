package main

import (
	"fmt"
	"github.com/bytegust/spm"
	"github.com/urfave/cli"
	"log"
	"os"
	"regexp"
)

var procfile string

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	app := cli.NewApp()
	app.Name = "spm - Simple Process Manager"
	app.Usage = "spm [OPTIONS] COMMAND [args...]"
	app.Version = "0.0.1"
	app.Author = "duanquanyong@outlook.com"

	app.Commands = cli.Commands{
		{
			Name:  "daemon",
			Usage: "run spm daemon service",
			Action: startDaemon,
			Subcommands: cli.Commands{
				{
					Name:   "install",
					Action: daemonInstall,
				},
				{
					Name:   "remove",
					Action: daemonRemove,
				},
				{
					Name:   "start",
					Action: daemonStart,
				},
				{
					Name:   "stop",
					Action: daemonStop,
				},
				{
					Name:   "status",
					Action: daemonStatus,
				},
			},
		},
		{
			Name:"start",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "file, f",
					Value:       "./",
					Usage:       "procfile location (e.g. ./spm/cmd/Procfile or ./spm/cmd/)",
					Destination: &procfile,
				},
			},
			Usage:"Starts tasks if present in Procfile",
			Action: startAction,
		},
		{
			Name:"stop",
			Usage:" Stop tasks if currently running",
			Action: stopAction,
		},
		{
			Name:"list",
			Usage:"Lists all running tasks",
			Action: listAction,
		},
		{
			Name:"log",
			Usage:"Prints last 200 lines of task's logfile",
			UsageText:"spm log [task...]",
			Action: logsAction,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func startAction(c *cli.Context)  {
	procfile := getProcfilePath(procfile)
	file, err := os.Open(procfile)
	if err != nil {
		log.Fatal(err)
	}

	p := spm.NewParser(file)
	jobs, err := p.Parse()
	if err != nil {
		log.Fatal(err)
	}

	sock := spm.NewSocket()
	if err := sock.Dial(); err != nil {
		log.Fatal(err)
	}

	var j []spm.Task
	if args := c.Args(); len(args) > 0 {
		for _, arg := range args {
			exist := false
			for _, job := range jobs {
				if job.Name == arg {
					j = append(j, job)
					exist = true
					break
				}
			}
			if !exist {
				fmt.Printf("job %s is not exist in procfile\n", arg)
			}
		}
	} else {
		j = jobs
	}

	if err := sock.Send(spm.Message{
		Command: "start",
		Jobs:    j,
	}); err != nil {
		log.Fatal(err)
	}

	<-sock.Message
	log.Println("done")
}

func stopAction(c *cli.Context)  {
	sock := spm.NewSocket()
	if err := sock.Dial(); err != nil {
		log.Fatal(err)
	}

	if err := sock.Send(spm.Message{
		Command:   "stop",
		Arguments: c.Args(),
	}); err != nil {
		log.Fatal(err)
	}

	<-sock.Message
	log.Println("done")
}

func listAction(c *cli.Context)  {
	sock := spm.NewSocket()
	if err := sock.Dial(); err != nil {
		log.Fatal(err)
	}

	if err := sock.Send(spm.Message{
		Command: "list",
	}); err != nil {
		log.Fatal(err)
	}

	m := <-sock.Message
	fmt.Println("Running jobs:")
	for _, job := range m.JobList {
		fmt.Printf("\t%s\n", job)
	}
	fmt.Println("") // line break
}

func logsAction(c *cli.Context) error {
	if job := c.Args().First(); job == "" {
		return cli.ShowCommandHelp(c, c.Command.Name)
	}

	sock := spm.NewSocket()
	if err := sock.Dial(); err != nil {
		log.Fatal(err)
	}

	if err := sock.Send(spm.Message{
		Command:   "logs",
		Arguments: []string{c.Args().Get(1)},
	}); err != nil {
		log.Fatal(err)
	}

	m := <-sock.Message
	for i := len(m.JobLogs) - 1; i >= 0; i-- {
		fmt.Println(m.JobLogs[i])
	}

	return nil
}


func getProcfilePath(input string) string {
	re := regexp.MustCompile("(/)$|(/Procfile(\\s+?|$))")
	match := re.FindStringSubmatch(input)

	if len(match) > 0 {
		if match[1] == "/" {
			return input + "Procfile"
		} else if match[2] == "/Procfile" {
			return input
		}
	}

	return input + "/Procfile"
}
