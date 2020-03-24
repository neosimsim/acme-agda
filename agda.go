package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
)

const prompt = "JSON> "

type Agda struct {
	filename  string
	stdout    io.ReadCloser
	stderr    io.ReadCloser
	stdin     io.WriteCloser
	responses <-chan Response
}

func NewAgda(agdaCmdPath, filename string) (*Agda, error) {
	agdaCmd := exec.Command(agdaCmdPath, "--interaction-json")
	stdin, err := agdaCmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := agdaCmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := agdaCmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := agdaCmd.Start(); err != nil {
		return nil, err
	}
	responses := make(chan Response)
	go func(res chan<- Response) {
		reader := bufio.NewReader(stdout)
		for {
			if line, err := reader.ReadString('\n'); err != nil {
				log.Printf("error reading agda output line: %s", err)
			} else {
				// drop the prompt
				if response, err := parseResponse(strings.TrimPrefix(line, prompt)); err != nil {
					log.Printf("error parsing response: %s", err)
				} else {
					res <- response
				}
			}
		}
	}(responses)
	return &Agda{filename: filename, stdin: stdin, stdout: stdout, stderr: stderr, responses: responses}, nil
}

func (a *Agda) Responses() <-chan Response {
	return a.responses
}

func (a *Agda) writeCommand(cmdFmt string, vals ...interface{}) error {
	cmdString := fmt.Sprintf(`IOTCM "%s" None Direct (%s)
`, a.filename, fmt.Sprintf(cmdFmt, vals...)) // The new line is important
	debugPrint("sending command: %s", cmdString)
	_, err := io.WriteString(a.stdin, cmdString)
	return err
}

func (a *Agda) Load(args ...string) error {
	return a.writeCommand(`Cmd_load "%s" [%s]`, a.filename, strings.Join(args, ","))
}

func (a *Agda) Compile(args ...string) error {
	return a.writeCommand(`Cmd_compile agda "%s" [%s]`, a.filename, strings.Join(args, ","))
}

func (a *Agda) Constraints() error {
	return a.writeCommand("Cmd_constraints")
}

func (a *Agda) Metas() error {
	return a.writeCommand("Cmd_metas")
}

func (a *Agda) ShowModule(moduleName string) error {
	return a.writeCommand(`Cmd_show_module_contents_toplevel AsIs "%s"`, moduleName)
}

func (a *Agda) SolveAll() error {
	return a.writeCommand(`Cmd_solveAll AsIs`)
}

func (a *Agda) SolveOne(goalIdx uint, arg string) error {
	return a.writeCommand(`Cmd_solveOne AsIs %d noRange "%s"`, goalIdx, arg)
}

func (a *Agda) AutoAll() error {
	return a.writeCommand(`Cmd_autoAll AsIs`)
}

func (a *Agda) AutoOne(goalIdx uint, arg string) error {
	return a.writeCommand(`Cmd_autoOne %d noRange "%s"`, goalIdx, arg)
}

// ...

func (a *Agda) MakeCase(goalIdx uint, arg string) error {
	return a.writeCommand(`Cmd_make_case %d noRange "%s"`, goalIdx, arg)
}

func (a *Agda) Refine(goalId uint, content string) error {
	return a.writeCommand(`Cmd_refine %d noRange "%s"`, goalId, content)
}