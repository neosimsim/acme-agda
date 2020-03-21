package main

import (
	"encoding/json"
)

func parseResponse(response string) (Response, error) {
	debugPrint("parsing: %s", response)
	var unkownResp map[string]interface{}
	if err := json.Unmarshal([]byte(response), &unkownResp); err != nil {
		return nil, err
	} else {
		debugPrint("parsing intermediate map: %v", unkownResp)
		switch unkownResp["kind"] {
		case "DisplayInfo":
			var resp Resp_DisplayInfo
			if err := json.Unmarshal([]byte(response), &resp); err != nil {
				return nil, err
			} else {
				return resp, nil
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
	Id    uint
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

type ConstraintObj int

type OutputConstraint struct {
	Kind           string
	Type           string
	OfType         string
	Comparison     string
	ConstraintObj  ConstraintObj
	ConstraintObjs []ConstraintObj
	Problem        []ProblemId
	Value          string
	Polarities     []string
	Arguments      []string
	Candidates     map[string]string
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

type Info_Intro_NotFound struct{}

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

// Combines attributes of all Info_ struct for easier access.
// Less type safe.
type Info_Union struct {
	Kind	string
	CommandState     CommandState
	ComputeMode      ComputeMode
	Constraints      []OutputForm
	Contents         []NamedType
	Context          []ResponseContextEntry
	Errors           string
	Expr             string
	Filepath         string
	GoalInfo         GoalDisplayInfo
	Info             Info
	InteractionPoint InteractionId
	InvisibleGoals   []OutputConstraint
	Message          string
	Name             Name
	Names            []Name
	Results          [][]NamedType
	Search           string
	Telescope        []DomType
	Term             Type
	Thing            string
	Time             CPUTime
	Version          string
	VisibleGoals     []OutputConstraint
	Warnings         string
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
	Info Info_Union
}

type Resp_DisplayInfo_CompilationOk struct {
	Info Info_CompilationOk
}
type Resp_DisplayInfo_Constraints struct {
	Info Info_Constraints
}
type Resp_DisplayInfo_AllGoalsWarnings struct {
	Info Info_AllGoalsWarnings
}

type Resp_DisplayInfo_Time struct {
	Info Info_Time
}

type Resp_DisplayInfo_Error struct {
	Info Info_Error
}

type Resp_DisplayInfo_Intro_NotFound struct {
	Info Info_Intro_NotFound
}

type Resp_DisplayInfo_Auto struct {
	Info Info_Auto
}

type Resp_DisplayInfo_ModuleContents struct {
	Info Info_ModuleContents
}

type Resp_DisplayInfo_SearchAbout struct {
	Info Info_SearchAbout
}

type Resp_DisplayInfo_WhyInScope struct {
	Info Info_WhyInScope
}

type Resp_DisplayInfo_NormalForm struct {
	Info Info_NormalForm
}

type Resp_DisplayInfo_InferredType struct {
	Info Info_InferredType
}

type Resp_DisplayInfo_Context struct {
	Info Info_Context
}

type Resp_DisplayInfo_Version struct {
	Info Info_Version
}

type Resp_DisplayInfo_GoalSpecific struct {
	Info Info_GoalSpecific
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