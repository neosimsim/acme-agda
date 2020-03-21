package main

import (
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
							debugPrint("response %T%v", r, r)
							switch r.(type) {
							case Resp_MakeCase:
								makeCase := r.(Resp_MakeCase)
								line := makeCase.InteractionPoint.Range[0].Start.Line
								if err := editWin.Addr("%d", line); err != nil {
									log.Printf("could write read addr for MakeCase: %s", err)
								} else {
									clauses := strings.Join(makeCase.Clauses, "\n") + "\n"
									if _, err := editWin.Write("data", []byte(clauses)); err != nil {
										log.Printf("could write result of GiveAction: %s", err)
									}
								}
							case Resp_DisplayInfo:
								menu.DisplayInfo = r.(Resp_DisplayInfo).Info
								menu.Error = nil
								menu.Redraw()
							case Resp_GiveAction:
								giveAction := r.(Resp_GiveAction)
								if giveAction.GiveResult.Str != "" {
									start := giveAction.InteractionPoint.Range[0].Start.Pos
									end := giveAction.InteractionPoint.Range[0].End.Pos
									if err := editWin.Addr("#%d,#%d", start-1, end-1); err != nil {
										log.Printf("could not write addr for GiveAction: %s", err)
									} else {
										if _, err := editWin.Write("data", []byte(giveAction.GiveResult.Str)); err != nil {
											log.Printf("could Write result of GiveAction: %s", err)
										}
									}
								}
							case Resp_InteractionPoints:
								menu.InteractionPoints = r.(Resp_InteractionPoints).InteractionPoints
							case Resp_JumpToError:
							default:
								debugPrint("unhandled response: %T %v", r, r)
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
