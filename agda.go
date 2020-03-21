package main

import (
	"bufio"
	"encoding/json"
	"errors"
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
	return a.writeCommand(fmt.Sprintf(`Cmd_make_case %d noAgdaRange "%s"`, goalIdx, varName))
}

func (a *Agda) RefineHole(goalIdx int, content string) error {
	return a.writeCommand(fmt.Sprintf(`Cmd_refine %d noAgdaRange "%s"`, goalIdx, content))
}

func (*Agda) Kill() {
}

func parseResponse(response string) (Response, error) {
	debugPrint("parsing: %s", response)
	var unkownResp map[string]interface{}
	if err := json.Unmarshal([]byte(response), &unkownResp); err != nil {
		return nil, err
	} else {
		debugPrint("parsing intermediate map: %v", unkownResp)
		switch unkownResp["kind"] {
		case "DisplayInfo":
			if info, err := parseDisplayInfo(unkownResp["info"]); err != nil {
				return nil, err
			} else {
				return Resp_DisplayInfo{Info: info}, nil
			}
		case "ClearHighlighting":
			return Resp_ClearHighlighting{}, nil
		case "DoneAborting":
			return Resp_DoneAborting{}, nil
		case "DoneExiting":
			return Resp_DoneExiting{}, nil
		case "ClearRunningInfo":
			return Resp_ClearRunningInfo{}, nil
		case "RunningInfo":
			var resp Resp_RunningInfo
			if err := json.Unmarshal([]byte(response), &resp); err != nil {
				return nil, err
			} else {
				return resp, nil
			}
		case "Status":
			var resp Resp_Status
			if err := json.Unmarshal([]byte(response), &resp); err != nil {
				return nil, err
			} else {
				return resp, nil
			}
		case "JumpToError":
			var resp Resp_JumpToError
			if err := json.Unmarshal([]byte(response), &resp); err != nil {
				return nil, err
			} else {
				return resp, nil
			}
		case "InteractionPoints":
			var resp Resp_InteractionPoints
			if err := json.Unmarshal([]byte(response), &resp); err != nil {
				return nil, err
			} else {
				return resp, nil
			}
		case "GiveAction":
			var resp Resp_GiveAction
			if err := json.Unmarshal([]byte(response), &resp); err != nil {
				return nil, err
			} else {
				return resp, nil
			}
		case "MakeCase":
			var resp Resp_MakeCase
			if err := json.Unmarshal([]byte(response), &resp); err != nil {
				return nil, err
			} else {
				return resp, nil
			}
		case "SolveAll":
			var resp Resp_SolveAll
			if err := json.Unmarshal([]byte(response), &resp); err != nil {
				return nil, err
			} else {
				return resp, nil
			}
		}
		return nil, nil
	}
}

func parseDisplayInfo(thing interface{}) (DisplayInfo, error) {
	if infoMap, ok := thing.(map[string]interface{}); !ok {
		return nil, errors.New("DisplayInfo should be an (JSON) object")
	} else {
		switch infoMap["kind"] {
		case "CompilationOk":
			return Info_CompilationOk{Warnings: infoMap["warnings"].(string), Errors: infoMap["errors"].(string)}, nil
		default:
			return nil, errors.New(fmt.Sprintf("unknown DiplayInfo %v", thing))
		}
	}
}

type Status struct {
	// Are implicit arguments displayed
	ShowImplicitArguments bool
	// Has the module been successfully type checked?
	Checked bool
}

// find $AGDA_SRCDIR -type f | xargs grep -n '^data CommandState'
// find $AGDA_SRCDIR -type f | xargs grep -n '^instance EncodeTCM CommandState'
type CommandState struct {
	InteractionPoints []InteractionId
	currentFile       interface{}
}

type Name string

// find $AGDA_SRCDIR -type f | xargs grep -n '^data NameInScope'
// find $AGDA_SRCDIR -type f | xargs grep -n '^instance EncodeTCM NameInScope'
type NameInScope bool

// find $AGDA_SRCDIR -type f | xargs grep -n '^data ResponseContextEntry'
// find $AGDA_SRCDIR -type f | xargs grep -n '^instance EncodeTCM ResponseContextEntry'
type ResponseContextEntry struct {
	OriginalName Name
	ReifiedName  Name
	Binding      interface{}
	InScope      bool
}

// find $AGDA_SRCDIR -type f | xargs grep -n '^data Position'
// find $AGDA_SRCDIR -type f | xargs grep -n '^instance EncodeTCM (Position'
type Position struct {
	Pos  int
	Line int
	Col  int
}

type Interval struct {
	Start Position
	End   Position
}

// find $AGDA_SRCDIR -type f | xargs grep -n '^data Range'
// find $AGDA_SRCDIR -type f | xargs grep -n '^instance EncodeTCM Range'
type AgdaRange []Interval

// find $AGDA_SRCDIR -type f | xargs grep -n '^newtype ProblemId'
// find $AGDA_SRCDIR -type f | xargs grep -n '^instance EncodeTCM ProblemId'
type ProblemId uint

// find $AGDA_SRCDIR -type f | xargs grep -n '^newtype InteractionId'
// find $AGDA_SRCDIR -type f | xargs grep -n '^instance EncodeTCM InteractionId'
type InteractionId struct {
	Id        uint
	Range AgdaRange
}

// find $AGDA_SRCDIR -type f | xargs grep -n '^data GiveResult'
// find $AGDA_SRCDIR -type f | xargs grep -n '^instance EncodeTCM GiveResult'
type GiveResult struct {
	Str   string
	Paren bool
}

// find $AGDA_SRCDIR -type f | xargs grep -n '^data MakeCaseVariant'
// find $AGDA_SRCDIR -type f | xargs grep -n '^instance EncodeTCM MakeCaseVariant'
type MakeCaseVariant string

// find $AGDA_SRCDIR -type f | xargs grep -n '^data Rewrite'
// find $AGDA_SRCDIR -type f | xargs grep -n '^instance EncodeTCM Rewrite'
type Rewrite string

// find $AGDA_SRCDIR -type f | xargs grep -n '^data CPUTime'
// find $AGDA_SRCDIR -type f | xargs grep -n '^instance EncodeTCM CPUTime'
type CPUTime string

// find $AGDA_SRCDIR -type f | xargs grep -n '^data ComputeMode'
// find $AGDA_SRCDIR -type f | xargs grep -n '^instance EncodeTCM ComputeMode'
type ComputeMode string

// find $AGDA_SRCDIR -type f | xargs grep -n '^data OutputForm'
// find $AGDA_SRCDIR -type f | xargs grep -n '^instance EncodeTCM OutputForm'
type OutputForm struct {
	AgdaRange  AgdaRange
	Problems   []ProblemId
	Constraint OutputConstraint
}

type OutputConstraint struct {
	Comparison     string
	ConstraintObjs interface{}
}

// find $AGDA_SRCDIR -type f | xargs grep -n '^data DisplayInfo'
// find $AGDA_SRCDIR -type f | xargs grep -n '^instance EncodeTCM DisplayInfo'
type DisplayInfo interface{}

type Info_CompilationOk struct {
	Warnings string
	Errors   string
}

type Info_Constraints struct {
	Constraints []OutputForm
}

type Info_AllGoalsWarnings struct {
	Warnings       string
	Errors         string
	VisibleGoals   []OutputConstraint
	InvisibleGoals []OutputConstraint
}

type Info_Time struct {
	Time CPUTime
}

type Info_Error struct {
	Message string
}

type Info_Into_NotFound struct{}

type Info string

// find $AGDA_SRCDIR -type f | xargs grep -n '^ *| *Info_Auto'
type Info_Auto struct {
	Info Info
}

type Type string

type NamedType struct {
	Name Name
	Term Type
}

// find $AGDA_SRCDIR -type f | xargs grep -n '^  *| *Info_ModuleContents'
type Info_ModuleContents struct {
	Contents  []NamedType
	Names     []Name
	Telescope []DomType
}

type BareName string

type DomType struct {
	Dom       string
	Name      BareName
	Finite    interface{}
	Chohesion string
	Relevance string
	Hiding    string
}

// find $AGDA_SRCDIR -type f | xargs grep -n '^  *| *Info_SearchAbout'
type Info_SearchAbout struct {
	Results [][]NamedType
	Search  string
}

// find $AGDA_SRCDIR -type f | xargs grep -n '^  *| *Info_WhyInScope'
type Info_WhyInScope struct {
	Thing    string
	Filepath string
	Message  string
}

// find $AGDA_SRCDIR -type f | xargs grep -n '^  *| *Info_NormalForm'
type Info_NormalForm struct {
	CommandState CommandState
	ComputeMode  ComputeMode
	Time         CPUTime
	Expr         string
}

// find $AGDA_SRCDIR -type f | xargs grep -n '^  *| *Info_InferredType'
type Info_InferredType struct {
	CommandState CommandState
	Time         CPUTime
	Expr         string
}

// find $AGDA_SRCDIR -type f | xargs grep -n '^  *| *Info_Context'
type Info_Context struct {
	InteractionPoint InteractionId
	Context          []ResponseContextEntry
}

// find $AGDA_SRCDIR -type f | xargs grep -n '^  *| *Info_Version'
type Info_Version struct {
	Version string
}

// find $AGDA_SRCDIR -type f | xargs grep -n '^  *| *Info_GoalSpecific'
type Info_GoalSpecific struct {
	InteractionPoint InteractionId
	GoalInfo         GoalDisplayInfo
}

type GoalDisplayInfo interface{}

type Goal_HelperFunction struct {
	Signature interface{}
}

type Goal_NormalForm struct {
	ComputeMode ComputeMode
	Expr        string
}

type Goal_GoalType struct {
	Rewrite     string
	TypeAux     interface{}
	Expr        string
	Type        string
	Boundary    []string
	OutputForms []string
}

type Goal_CurrentGoal struct {
	Rewrite string
	Type    string
}

type Goal_InferredType struct {
	Expr string
}

type Response interface{}

type Resp_HighlightingInfo struct{}

type Resp_DisplayInfo struct {
	Info DisplayInfo
}

type Resp_ClearHighlighting struct{}

type Resp_DoneAborting struct{}

type Resp_DoneExiting struct{}

type Resp_ClearRunningInfo struct{}

type Resp_RunningInfo struct {
	DebugLevel int
	Message    string
}

type Resp_Status struct {
	Status Status
}

type Resp_JumpToError struct {
	Filepath string
	Position Position
}

type Resp_InteractionPoints struct {
	InteractionPoints []InteractionId
}

type Resp_GiveAction struct {
	InteractionPoint InteractionId
	GiveResult       GiveResult
}

type Solution struct {
	InteractionPoint InteractionId
	Expression       string
}

type Resp_SolveAll struct {
	Solutions []Solution
}

type Resp_MakeCase struct {
	InteractionPoint InteractionId
	Variant          string
	Clauses          []string
}