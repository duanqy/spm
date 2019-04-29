package spm

import (
	"os/exec"
)

type Task struct {
	Name    string
	Command []string
	Logging *Logging

	NotifyEnd chan bool `json:"-"`
	Cmd       *exec.Cmd `json:"-"`

	Chroot string
	Dir string
	User  string
	Group string
	Env   []string
}

func (t Task)Valid() bool  {
	return t.Name != "" && len(t.Command) > 0
}