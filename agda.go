package main

import (
	"bufio"
	"encoding/json"
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

type Response interface{}

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

func (a *Agda) writeCommand(cmd string) error {
	cmdString := fmt.Sprintf(`IOTCM "%s" None Direct (%s)
`, a.filename, cmd) // The new line is important
	debugPrint("sending command: %s", cmdString)
	_, err := io.WriteString(a.stdin, cmdString)
	return err
}
func (a *Agda) LoadFile(args ...string) error {
	return a.writeCommand(fmt.Sprintf(`Cmd_load "%s" [%s]`, a.filename, strings.Join(args, ",")))
}

func (a *Agda) CaseSplit(goalIdx int, varName string) error {
	return a.writeCommand(fmt.Sprintf(`Cmd_make_case %d noRange "%s"`, goalIdx, varName))
}

func (a *Agda) RefineHole(goalIdx int, content string) error {
	return a.writeCommand(fmt.Sprintf(`Cmd_refine %d noRange "%s"`, goalIdx, content))
}

func (*Agda) Kill() {
}

type InteractionId uint
type RespMakeCase struct {
	Variant string
	// Lines which replace the line of the goal.
	Clauses []string
}

type DisplayInfo struct {
	Goals    string
	Warnings string
	Errors   string
	Payload  string
}

type RespDisplayInfo struct {
	Info DisplayInfo
}

// TODO remove Kind
type RespClearHighlighting struct{}
type RespDoneAborting struct{}
type RespClearRunningInfo struct{}
type RespRunningInfo struct {
	kind       string
	debugLevel int
	message    string
}
type RespStatus struct {
	kind string

	// Are implicit arguments displayed
	showImplicitArguments bool

	// Has the module been successfully type checked?
	Checked bool
}
type RespJumpToError struct {
	FilePath string
	Position int32
}

type RespInteractionPoints struct {
	interactionPoints []InteractionId
}
type RespGiveAction struct {
	InteractionPoint InteractionId
	GiveResult       string // FIXME might also be a bool
}

type RespSolveAll struct {
	solutions map[InteractionId]string
}

func parseResponse(response string) (Response, error) {
	debugPrint("parsing: %s", response)
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		return nil, err
	} else {
		debugPrint("parsing intermediate map: %v", resp)
		switch resp["kind"] {
		case "DisplayInfo":
			var respDisplayInfo RespDisplayInfo
			if err := json.Unmarshal([]byte(response), &respDisplayInfo); err != nil {
				return nil, err
			} else {
				return respDisplayInfo, nil
			}
		case "ClearHighlighting":
			return RespClearHighlighting{}, nil
		case "DoneAborting":
			return RespDoneAborting{}, nil
		case "ClearRunningInfo":
			return RespClearRunningInfo{}, nil
		case "RunningInfo":
			return RespRunningInfo{}, nil
		case "Status":
			return RespStatus{}, nil
		case "JumpToError":
			var respJumpToError RespJumpToError
			if err := json.Unmarshal([]byte(response), &respJumpToError); err != nil {
				return nil, err
			} else {
				return respJumpToError, nil
			}
		case "InteractionPoints":
			return RespInteractionPoints{}, nil
		case "GiveAction":
			var respGiveAction RespGiveAction
			if err := json.Unmarshal([]byte(response), &respGiveAction); err != nil {
				return nil, err
			} else {
				return respGiveAction, nil
			}
		case "MakeCase":
			var respMakeCase RespMakeCase
			if err := json.Unmarshal([]byte(response), &respMakeCase); err != nil {
				return nil, err
			} else {
				return respMakeCase, nil
			}
		case "SolveAll":
			return RespSolveAll{}, nil
		}
		return nil, nil
	}
}
