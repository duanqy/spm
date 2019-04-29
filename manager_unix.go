// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package spm

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
)

func setupUserAndGroup(c *exec.Cmd, task Task) error {
	c.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:true,
		Credential: &syscall.Credential{
			Uid: uint32(os.Geteuid()),
			Gid: uint32(os.Getegid()),
			NoSetGroups:true,
		},
	}
	if task.Chroot != "" {
		c.SysProcAttr.Chroot = task.Chroot
	}

	if task.User != "" {
		u, err := user.Lookup(task.User)
		if err != nil {
			return fmt.Errorf("task '%s' user lookup with error: %s", task.Name, err)
		}
		uid, _ := strconv.Atoi(u.Uid)
		gid, _ := strconv.Atoi(u.Gid)
		c.SysProcAttr.Credential.Uid = uint32(uid)
		c.SysProcAttr.Credential.Gid = uint32(gid)
		c.Env = append(c.Env, "HOME=" + u.HomeDir)
	}
	if task.Group != "" {
		g,err := user.LookupGroup(task.Group)
		if err != nil {
			return fmt.Errorf("task '%s' user lookupGroup with error: %s", task.Name, err)
		}
		gid, _ := strconv.Atoi(g.Gid)
		c.SysProcAttr.Credential.Gid = uint32(gid)
	}
	return nil
}
