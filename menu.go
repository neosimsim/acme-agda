package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"9fans.net/go/acme"
)

const menuText = `Get Case Refine AutoOne Next Goal
GoalType

{{ with .DisplayInfo}}{{ with .Warnings}}Warnings:
{{ . }}{{ end }}{{ with .Errors}}Errors:
{{ . }}{{ end }}{{ with .Message}}Message:
{{ . }}{{ end }}{{ with .InvisibleGoals}}InvisibleGoals:
{{ range . }}{{ template "outputConstraint" . }}{{ end }}{{ end }}{{ with .VisibleGoals}}VisibleGoals:
{{ range . }}{{ template "outputConstraint" . }}{{ end }}{{ end }}
{{ with .InteractionPoint }}InteractionPoint {{ .Id }}:{{ range .Range }}
{{ $.AgdaFile }}:#{{ decr .Start.Pos }},#{{ decr .End.Pos }}{{ end }}{{ end }}
{{ with .GoalInfo }}{{.Type}}{{ end }}{{ end }}
{{ with .Error }}{{ .Error }}{{ end }}

{{ define "outputConstraint" }}?{{ .ConstraintObj }} : {{ .Type }}
{{ end }}`

type Menu struct {
	menuWin           *acme.Win
	agdaWin           *acme.Win
	template          *template.Template
	agdaInteraction   *Agda
	DisplayInfo       Info_Union
	InteractionPoints []InteractionId
	Error             error
	AgdaFile          string
}

func NewMenu(agdaInteraction *Agda, agdaWin *acme.Win) (*Menu, error) {
	var menu Menu
	var err error

	funcMap := template.FuncMap{
		"decr": func(i int) int {
			return i - 1
		},
	}
	if menu.template, err = template.New("menu").Funcs(funcMap).Parse(menuText); err != nil {
		return nil, errors.Unwrap(fmt.Errorf("cannot parse menu templates: %w", err))
	}
	if menu.menuWin, err = acme.New(); err != nil {
		return nil, errors.Unwrap(fmt.Errorf("cannot open new acme menuWindow: %w", err))
	}
	if menu.AgdaFile, err = WindowName(agdaWin); err != nil {
		return nil, errors.Unwrap(fmt.Errorf("cannot read Agda file from window name: %w", err))
	}
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		return nil, errors.Unwrap(fmt.Errorf("cannot get current directory: %w", err))
	}
	if err = menu.menuWin.Name("%s/+Acme", currentWorkingDirectory); err != nil {
		return nil, errors.Unwrap(fmt.Errorf("cannot set acme menuWindow name: %w", err))
	}
	menu.agdaInteraction = agdaInteraction
	menu.agdaWin = agdaWin
	return &menu, nil
}

func (menu *Menu) Redraw() {
	if err := menu.menuWin.Addr(","); err != nil {
		log.Printf("error writing display address: %s", err)
	} else {
		var builder strings.Builder
		menu.template.Execute(&builder, menu)
		if _, err := menu.menuWin.Write("data", []byte(builder.String())); err != nil {
			log.Printf("error writing display info: %s", err)
		} else {
			if err := menu.menuWin.Addr("0"); err != nil {
				log.Printf("error resetting display address: %s", err)
			} else {
				menu.menuWin.Ctl("dot=addr")
				menu.menuWin.Ctl("show")
			}
		}
	}
}

func (menu *Menu) Loop() {
	for e := range menu.menuWin.EventChan() {
		go func(event *acme.Event) {
			switch event.C2 {
			case 'x', 'X':
				switch string(event.Text) {
				case "Del":
					if err := menu.menuWin.Ctl("delete"); err != nil {
						log.Fatalln("Failed to delete the menuWindow:", err)
					}
					os.Exit(0)
				case "Get":
					debugPrint("loading file")
					if err := menu.agdaWin.Ctl("put"); err != nil {
						log.Printf("could save file: %s", err)
					}
					if err := menu.agdaInteraction.Load(); err != nil {
						log.Printf("could not load file: %s", err)
					}
				case "Case":
					debugPrint("doing case split")
					interactionId, goalContent, err := SelectedInteraction(menu.agdaWin, menu.InteractionPoints)
					if err != nil {
						debugPrint("move dot inside a goal: %s", err)
						menu.Error = errors.New("Move dot inside a goal. Have you loaded the file?")
						menu.Redraw()
						return
					}
					if err := menu.agdaInteraction.MakeCase(interactionId.Id, goalContent); err != nil {
						log.Printf("could not MakeCase: %s", err)
					}
				case "Refine":
					debugPrint("refine goal")
					interactionId, goalContent, err := SelectedInteraction(menu.agdaWin, menu.InteractionPoints)
					if err != nil {
						debugPrint("move dot inside a goal: %s", err)
						menu.Error = errors.New("Move dot inside a goal. Have you loaded the file?")
						menu.Redraw()
						return
					}
					if err := menu.agdaInteraction.Refine(interactionId.Id, goalContent); err != nil {
						log.Printf("could not Refine goal: %s", err)
					}
				case "AutoOne":
					debugPrint("running AutoOne on goal")
					interactionId, goalContent, err := SelectedInteraction(menu.agdaWin, menu.InteractionPoints)
					if err != nil {
						debugPrint("move dot inside a goal: %s", err)
						menu.Error = errors.New("Move dot inside a goal. Have you loaded the file?")
						menu.Redraw()
						return
					}
					if err := menu.agdaInteraction.AutoOne(interactionId.Id, goalContent); err != nil {
						log.Printf("could not Refine goal: %s", err)
					}
				case "GoalType":
					debugPrint("running GoalType")
					interactionId, _, err := SelectedInteraction(menu.agdaWin, menu.InteractionPoints)
					if err != nil {
						debugPrint("move dot inside a goal: %s", err)
						menu.Error = errors.New("Move dot inside a goal. Have you loaded the file?")
						menu.Redraw()
						return
					}
					if err := menu.agdaInteraction.GoalType(interactionId.Id); err != nil {
						log.Printf("could not call for goal type: %s", err)
					}
				case "Next":
					NextGoal(menu.agdaWin)
				case "Goal":
					ReplaceSelection(menu.agdaWin, "{!!}")
				default:
					menu.menuWin.WriteEvent(event)
				}
			default:
				menu.menuWin.WriteEvent(event)
			}
		}(e)
	}
}

func (menu *Menu) Close() {
	menu.menuWin.CloseFiles()
}