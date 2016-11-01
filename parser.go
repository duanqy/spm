package spm

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"regexp"
	"strings"
)

type Parser struct {
	r io.Reader
}

func NewParser(r io.Reader) *Parser {
	return &Parser{r: r}
}

func (p *Parser) Parse() (jobs []Job, err error) {
	reader := bufio.NewReader(p.r)

PARSE:
	for {
		job := Job{}
		var lines []byte

		for {
			line, _, err := reader.ReadLine()
			if err == io.EOF {
				return jobs, nil
			}
			if err != nil {
				return jobs, err
			}

			// trim leading and trailing spaces
			line = bytes.TrimSpace(line)

			if len(line) == 0 {
				continue PARSE
			} else if len(line) > 0 && line[len(line)-1] == '\\' {
				lines = append(lines, line[:len(line)-1]...)
				continue
			} else {
				lines = append(lines, line...)
			}

			// convert bytes to string
			slines := string(lines)

			// find string that matches regexp
			re := regexp.MustCompile("wait\\s+for\\s+([a-z]+)\\s([a-z-0-9.]+):([0-9]+)(.*)?:")

			matched := re.FindString(slines)
			if matched != "" {
				wait := strings.TrimPrefix(strings.TrimSuffix(matched, ":"), "wait for")
				ports := strings.Split(wait, ",")

				for _, port := range ports {
					p := strings.Split(strings.TrimSpace(port), " ")
					job.WaitSockets = append(job.WaitSockets, WaitSocket{
						Type: strings.TrimSpace(p[0]),
						Addr: strings.TrimSpace(p[1]),
					})
				}

				slines = strings.Join(re.Split(slines, -1), "")
			}

			sp := strings.SplitN(slines, ":", 2)
			if len(sp) < 2 {
				return jobs, errors.New("spm: missing command")
			}

			job.Name = sp[0]
			commandsStr := sp[1]

			if job.Name == "" {
				return jobs, errors.New("spm: invalid name")
			}

			job.Command = strings.Trim(commandsStr, " ")

			jobs = append(jobs, job)
			break
		}
	}
}
