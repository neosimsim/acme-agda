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
							log.Printf("response: %T%v", r, r)
							switch r.(type) {
							case Resp_MakeCase:
								debugPrint("response %T%v", r, r)
								SelectCurrentLine(editWin)
								ReplaceSelection(editWin, fmt.Sprintf("%s\n", strings.Join(r.(Resp_MakeCase).Clauses, "\n")))
							case Resp_DisplayInfo:
								debugPrint("response %T%v", r, r)
								menu.DisplayInfo = r.(Resp_DisplayInfo).Info
								menu.Error = nil
								menu.Redraw()
							case Resp_GiveAction:
								debugPrint("response %T%v", r, r)
							case Resp_JumpToError:
								debugPrint("response %T%v", r, r)
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
