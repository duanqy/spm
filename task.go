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

	Dir string
	User  string
	Group string
	Env   []string
}

func (j Task)Valid() bool  {
	return j.Name != "" && len(j.Command) > 0
}