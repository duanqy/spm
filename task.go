package spm

import (
	"os/exec"
)

type Task struct {
	Name    string
	Command []string
	Logger  *Logger

	NotifyEnd chan bool `json:"-"`
	Cmd       *exec.Cmd `json:"-"`

	Chroot string
	Dir string
	User  string
	Group string
	Env   []string
	Need [][]string
}

func (t Task)Valid() bool  {
	return t.Name != "" && len(t.Command) > 0
}