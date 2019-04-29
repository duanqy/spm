package main

import (
	"fmt"
	"github.com/bytegust/spm"
	"github.com/takama/daemon"
	"github.com/urfave/cli"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

const (
	// name of the service
	name        = "spm"
	description = "Simple Process Manager"
)

//	dependencies that are NOT required by the service, but might be used
var dependencies = []string{"dummy.service"}

var stdlog, errlog *log.Logger

func init() {
	stdlog = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	errlog = log.New(os.Stderr, "", log.Ldate|log.Ltime)
}

func newDaemon() daemon.Daemon {
	srv, err := daemon.New(name, description, dependencies...)
	if err != nil {
		errlog.Println("Error: ", err)
		os.Exit(1)
	}
	return srv
}

func daemonInstall(ctx *cli.Context) error {
	status, err := newDaemon().Install("daemon")
	if err != nil {
		errlog.Println(status, "\nError: ", err)
	} else {
		fmt.Println(status)
	}
	return err
}

func daemonRemove(ctx *cli.Context) error {
	status, err := newDaemon().Remove()
	if err != nil {
		errlog.Println(status, "\nError: ", err)
	} else {
		fmt.Println(status)
	}
	return err
}

func daemonStart(ctx *cli.Context) error {
	status, err := newDaemon().Start()
	if err != nil {
		errlog.Println(status, "\nError: ", err)
	} else {
		fmt.Println(status)
	}
	return err
}

func daemonStop(ctx *cli.Context) error {
	status, err := newDaemon().Stop()
	if err != nil {
		errlog.Println(status, "\nError: ", err)
	} else {
		fmt.Println(status)
	}
	return err
}

func daemonStatus(ctx *cli.Context) error {
	status, err := newDaemon().Status()
	if err != nil {
		errlog.Println(status, "\nError: ", err)
	} else {
		fmt.Println(status)
	}
	return err
}

func startDaemon(c *cli.Context) {
	manager := spm.NewManager()
	sock := spm.NewSocket()

	// listen for user termination
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	// start listening for cli apps
	go func() {
		if err := sock.Listen(); err != nil {
			log.Fatal(err)
		}
	}()

	// handle incoming cli app connections
	go func() {
		for conn := range sock.Connection {
			go handleMessage(<-conn.Message, conn, manager)
		}
	}()

	log.Println("deamon started")
	defer func() {
		log.Println("deamon ended")
	}()

	for {
		select {
		case conn := <-sock.Connection:
			go handleMessage(<-conn.Message, conn, manager)
		case killSignal := <-interrupt:
			stdlog.Println("Got signal:", killSignal)
			stdlog.Println("Stoping listening")
			err := sock.Close()
			if err != nil {
				log.Println("close sock error: ", err)
			}
			if killSignal == os.Interrupt {
				log.Println("Daemon was interruped by system signal")
			} else {
				log.Println("Daemon was killed")
			}
			manager.StopAll()
			return
		}
	}
}

func handleMessage(mes spm.Message, conn *spm.Socket, manager *spm.Manager) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Println("close conn error: ", err)
		}
	}()

	switch mes.Command {
	case "start":
		go manager.StartAll(mes.Jobs)
	case "list":
		if err := conn.Send(spm.Message{
			JobList: manager.List(),
		}); err != nil {
			log.Println(err)
		}
	case "stop":
		if args := mes.Arguments; len(args) > 0 {
			for _, arg := range args {
				go manager.Stop(arg)
			}
		} else {
			manager.StopAll()
		}
	case "log":
		job := mes.Arguments[0]
		if job == "" {
			_ = conn.Send(spm.Message{
				JobLogs: []string{"task name cannot be empty"},
			})
			return
		}
		n, _ := strconv.ParseUint(mes.Arguments[1], 10, 64)
		if n == 0 {
			n = 200
		}
		if err := conn.Send(spm.Message{
			JobLogs: manager.ReadLog(job, 2),
		}); err != nil {
			log.Println(err)
		}
	}
}
