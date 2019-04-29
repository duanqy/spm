package spm

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/mholt/caddy/caddyfile"
	"io"
	"io/ioutil"
	"os/user"
	"path/filepath"
)

type Parser struct {
	filename string
	r        io.Reader
	cfg      []byte
}

func NewParser(r io.Reader) *Parser {
	return &Parser{r: r}
}

func (p *Parser) Parse() (jobs []Task, err error) {
	p.cfg, err = ioutil.ReadAll(p.r)
	if err != nil {
		return nil, err
	}
	d := caddyfile.NewDispenser(p.filename, bytes.NewBuffer(p.cfg))

	return parseTasks(d)
}

func parseTasks(d caddyfile.Dispenser) ([]Task, error) {
	var task Task
	tasks := make([]Task, 0, 10)
	for d.Next() {
		val := d.Val()
		args := d.RemainingArgs()
		switch val {
		case "task":
			var task Task
			if len(args) >= 1 {
				task.Name = args[0]
			}
			for d.NextBlock() {
				val := d.Val()
				args := d.RemainingArgs()
				if err := updateTask(&task, d, val, args); err != nil {
					return nil, err
				}
			}
			tasks = append(tasks, task)
		default:
			if err := updateTask(&task, d, val, args); err != nil {
				return nil, err
			}
		}
	}
	if task.Valid() {
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func updateTask(task *Task, d caddyfile.Dispenser, key string, args []string) error {
	switch key {
	case "name":
		if task.Name != "" {
			return fmt.Errorf("set name two times")
		}
		if len(args) != 1 {
			return d.ArgErr()
		}
		if args[0] == "" {
			return errors.New("task name can not be empty")
		}
		task.Name = args[0]
	case "command":
		if len(task.Command) > 0 {
			return fmt.Errorf("set command two times")
		}
		if len(args) < 1 {
			return d.ArgErr()
		}
		task.Command = args
	case "user":
		if task.User != "" {
			return fmt.Errorf("set user two times")
		}
		if len(args) != 1 {
			return d.ArgErr()
		}
		if _, err := user.Lookup(args[0]); err != nil {
			return err
		}
		task.User = args[0]
	case "group":
		if task.Group != "" {
			return fmt.Errorf("set group two times")
		}
		if len(args) != 1 {
			return d.ArgErr()
		}
		if _, err := user.LookupGroup(args[0]); err != nil {
			return err
		}
		task.Group = args[0]
	case "env":
		if task.Env == nil {
			task.Env = make([]string, 0, 10)
		}
		task.Env = append(task.Env, args...)
	case "dir":
		if task.Dir != "" {
			return fmt.Errorf("set dir two times")
		}
		if len(args) != 1 {
			return d.ArgErr()
		}
		if !filepath.IsAbs(args[0]) {
			return fmt.Errorf("dir %s is not an absolute address", args[0])
		}
		task.Dir = args[0]
	case "chroot":
		if task.Chroot != "" {
			return fmt.Errorf("set chroot two times")
		}
		if len(args) != 1 {
			return d.ArgErr()
		}
		if !filepath.IsAbs(args[0]) {
			return fmt.Errorf("chroot %s is not an absolute address", args[0])
		}
		task.Chroot = args[0]
	default:
		return errors.New("unsupported directive " + key)
	}
	return nil
}
