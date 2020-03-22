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

const menuText = `Get Case Refine Next Goal

{{template "displayInfo" .DisplayInfo}}
{{ with .Error }}{{ .Error }}{{ end }}

{{ define "displayInfo" }}{{ with .Goals}}Goals:
{{ . }}{{ end }}{{ with .Warnings}}Warnings:
{{ . }}{{ end }}{{ with .Errors}}Errors:
{{ . }}{{ end }}{{ with .Payload}}Payload:
{{ . }}{{ end }}{{ end }}
`

type Menu struct {
	menuWin         *acme.Win
	agdaWin         *acme.Win
	template        *template.Template
	agdaInteraction *Agda
	DisplayInfo     DisplayInfo
	Error           error
}

func NewMenu(agdaInteraction *Agda, agdaWin *acme.Win) (*Menu, error) {
	var menu Menu
	var err error
	if menu.template, err = template.New("menu").Parse(menuText); err != nil {
		return nil, errors.Unwrap(fmt.Errorf("cannot parse menu templates: %w", err))
	}
	if menu.menuWin, err = acme.New(); err != nil {
		return nil, errors.Unwrap(fmt.Errorf("cannot open new acme menuWindow: %w", err))
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
					if err := menu.agdaWin.Ctl("put"); err != nil {
						log.Printf("could save file: %s", err)
					}
					if err := menu.agdaInteraction.LoadFile(); err != nil {
						log.Printf("could not load file: %s", err)
					}
				case "Case":
					if err := SelectGoal(menu.agdaWin); err != nil {
						log.Printf("could not select goal: %s", err)
						return
					}
					start, end, err := menu.agdaWin.ReadAddr()
					if err != nil {
						log.Printf("could not read select goal address: %s", err)
						return
					}
					goalRanges, err := GoalRanges(menu.agdaWin)
					if err != nil {
						log.Printf("could not get goal rages: %s", err)
						return
					}
					goalIdx := -1
					for i, goalRange := range goalRanges {
						if goalRange.Start == start && goalRange.End == end {
							goalIdx = i
							break
						}
					}
					if goalIdx == -1 {
						log.Printf("move dot inside a goal")
					}
					goalContent := menu.agdaWin.Selection()
					goalContent = goalContent[2 : len(goalContent)-2] // drop {! and !}
					if err := menu.agdaInteraction.CaseSplit(goalIdx, goalContent); err != nil {
						log.Printf("could not load file: %s", err)
					}
				case "Refine":
					if err := SelectGoal(menu.agdaWin); err != nil {
						log.Printf("could not select goal: %s", err)
						return
					}
					start, end, err := menu.agdaWin.ReadAddr()
					if err != nil {
						log.Printf("could not read select goal address: %s", err)
						return
					}
					goalRanges, err := GoalRanges(menu.agdaWin)
					if err != nil {
						log.Printf("could not get goal rages: %s", err)
						return
					}
					goalIdx := -1
					for i, goalRange := range goalRanges {
						if goalRange.Start == start && goalRange.End == end {
							goalIdx = i
							break
						}
					}
					if goalIdx == -1 {
						log.Printf("move dot inside a goal")
					}
					goalContent := menu.agdaWin.Selection()
					goalContent = goalContent[2 : len(goalContent)-2] // drop {! and !}
					if err := menu.agdaInteraction.RefineHole(goalIdx, goalContent); err != nil {
						log.Printf("could not load file: %s", err)
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
