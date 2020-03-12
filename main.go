package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	agdaCmd = flag.String("with-agda", "agda", "Name or path of the agda compiler")
	debug   = flag.Bool("v", false, "Enable verbose debugging output")
)

const usageFmt = `Usage of %s:

Run this command from an Agda file opened in Acme.

Not all of the Agda interaction mode is supported yet.
Goal selection does not work on edge cases, either.

`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usageFmt, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if editWin, err := CallingWindow(); err != nil {
		log.Fatalf("cannot determine calling window: %v\n", err)
	} else {
		if err := ResetAddr(editWin); err != nil {
			log.Fatalf("cannot determine calling window: %s\n", err)
		}
		if agdaFile, err := WindowName(editWin); err != nil {
			log.Fatalf("cannot determine calling window: %s\n", err)
		} else {
			if a, err := NewAgda(*agdaCmd, agdaFile); err != nil {
				log.Fatalf("unable to start agda: %s", err)
			} else {
				if menu, err := NewMenu(a, editWin); err != nil {
					log.Fatalf("cannot open acme menu: %s\n", err)
				} else {
					defer menu.Close()
					menu.Redraw()
					go func() {
						for r := range a.Responses() {
							switch r.(type) {
							case RespMakeCase:
								SelectCurrentLine(editWin)
								ReplaceSelection(editWin, fmt.Sprintf("%s\n", strings.Join(r.(RespMakeCase).Clauses, "\n")))
							case RespDisplayInfo:
								menu.DisplayInfo = r.(RespDisplayInfo).Info
								menu.Error = nil
								menu.Redraw()
							case RespGiveAction:
								respGiveAction := r.(RespGiveAction)
								if goals, err := GoalRanges(editWin); err != nil {
									log.Printf("error finding goals: %s", err)
								} else {
									goal := goals[respGiveAction.InteractionPoint]
									if err := editWin.Addr("#%d,#%d", goal.Start, goal.End); err != nil {
										log.Printf("error writing goal address: %s", err)
									} else {
										if err := editWin.Ctl("dot=addr"); err != nil {
											log.Printf("error moving to goal: %s", err)
										} else {
											ReplaceSelection(editWin, respGiveAction.GiveResult)
										}
									}
								}
							case RespJumpToError:
								log.Printf("%v", r)
								respJumpToError := r.(RespJumpToError)
								menu.Error = errors.New(fmt.Sprintf("Error at %s:#%d", respJumpToError.FilePath, respJumpToError.Position))
								menu.Redraw()
							default:
								debugPrint("unknown response: %T %v", r, r)
							}
						}
					}()
					menu.Loop()
				}
			}
		}
	}
}

func debugPrint(f string, vals ...interface{}) {
	if *debug {
		log.Printf("DEBUG: "+f, vals...)
	}
}