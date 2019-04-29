package spm

import (
	"bufio"
	"fmt"
	"github.com/hpcloud/tail"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
)

type Manager struct {
	mu    sync.Mutex // protects following
	Tasks map[string]Task
}

func NewManager() *Manager {
	return &Manager{
		Tasks: make(map[string]Task),
	}
}

func (m *Manager) StartAll(tasks []Task) {
	for _, task := range tasks {
		m.Start(task)
	}
}

func setupCommand(task Task, cmd []string, pw io.Writer) (*exec.Cmd, error) {
	c := exec.Command(cmd[0], cmd[1:]...)
	if err := setupUserAndGroup(c, task); err != nil {
		return nil, fmt.Errorf("failed to set up running user: %s", err)
	}
	c.Stderr = pw
	c.Stdout = pw
	c.Env = make([]string, 0, 100)
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "HOME=") {
			c.Env = append(c.Env, e)
		}
	}
	c.Env = append(c.Env, task.Env...)
	if task.Dir != "" {
		c.Dir = task.Dir
	}
	return c, nil
}

func (m *Manager) Start(task Task) {
	if !task.Valid() {
		log.Println("task", task.Name, "包含无效命令")
		return
	}

	_, exists := m.Tasks[task.Name]
	if exists {
		log.Println(fmt.Sprintf("wont start task '%s' because already running", task.Name))
		return
	}

	logging, err := NewLogging(task.Name)
	if err != nil {
		log.Fatal(err)
	} else {
		task.Logger = logging
	}

	pr, pw, err := os.Pipe()
	if err != nil {
		log.Fatal(err)
	}
	// read command's stdout line by line
	in := bufio.NewScanner(pr)
	go func() {
		if err := task.Logger.Output(in); err != nil {
			log.Fatal(err)
		}
		fmt.Println("logger stoped")
	}()

	for i := range task.Need {
		cmd, err := setupCommand(task, task.Need[i], pw)
		if err != nil {
			log.Fatalf("setup command %s faield, error:%s\n", strings.Join(task.Need[i], " "), err)
		}
		if err := cmd.Run(); err != nil {
			log.Fatalf("execute command %s failed, error: %s\n", strings.Join(task.Need[i], " "), err)
		}
	}
	c, err := setupCommand(task, task.Command, pw)
	if err != nil {
		log.Fatalf("execute command %s of task %s failed, error: %s\n", strings.Join(task.Command, " "), task.Name, err)
	}
	task.NotifyEnd = make(chan bool)
	task.Cmd = c

	m.mu.Lock()
	m.Tasks[task.Name] = task
	m.mu.Unlock()

	log.Println(fmt.Sprintf("task `%s` has been started", task.Name))
	if err := c.Start(); err != nil {
		log.Println(err)
		go m.taskEnded(task)
		return
	}

	go func() {
		if err := c.Wait(); err != nil {
			log.Println(err)
		}
		m.taskEnded(task)
	}()
}

func (m *Manager) taskEnded(task Task) {
	m.mu.Lock()
	delete(m.Tasks, task.Name)
	m.mu.Unlock()
	if err := task.Logger.Close(); err != nil {
		log.Println("close task.Logger:", err)
	}
	log.Println(fmt.Sprintf("task `%s` ended", task.Name))
	task.NotifyEnd <- true
}

func (m *Manager) Stop(task string) {
	m.mu.Lock()
	j, exists := m.Tasks[task]
	m.mu.Unlock()
	if !exists {
		return
	}
	//pid, _ := syscall.Getpgid(j.Cmd.Process.Pid)
	//if err := syscall.Kill(-pid, syscall.SIGTERM);err != nil {
	//	log.Println(err)
	//}
	if err := j.Cmd.Process.Signal(syscall.SIGTERM); err != nil {
		log.Println(err)
	}
	<-j.NotifyEnd
}

func (m *Manager) StopAll() {
	var wg sync.WaitGroup
	for task := range m.Tasks {
		wg.Add(1)
		go func(task string) {
			m.Stop(task)
			wg.Done()
		}(task)
	}

	wg.Wait()
}

func (m *Manager) List() (tasks []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for task := range m.Tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// ReadLog reads last n lines of the file that corresponds to job.
func (m *Manager) ReadLog(task string, n int) (lines []string) {
	_, exists := m.Tasks[task]
	if !exists {
		lines = append(lines, "task "+task+" is not running")
		return
	}

	filename := m.Tasks[task].Logger.FileName()
	lines = append(lines, fmt.Sprintf("tail %s of %d lines", filename, n))
	t, err := tail.TailFile(filename, tail.Config{Follow: false})
	if err != nil {
		lines = append(lines, err.Error())
		return
	}
	for line := range t.Lines {
		lines = append(lines, line.Text)
	}
	return lines
}
