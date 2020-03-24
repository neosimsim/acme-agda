// Provides functions to select and edit text in an Acme window.
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"9fans.net/go/acme"
)

type Range struct {
	Start int
	End   int
}

// Returns the Acme window from which the application was executed,
// i.e. the window with ID matching the evironment variable winid.
func CallingWindow() (*acme.Win, error) {
	if id, err := strconv.Atoi(os.Getenv("winid")); err != nil {
		return nil, err
	} else {
		return acme.Open(id, nil)
	}
}

func WindowName(win *acme.Win) (string, error) {
	if windows, err := acme.Windows(); err != nil {
		return "", err
	} else {
		for _, winInfo := range windows {
			if win.ID() == winInfo.ID {
				return winInfo.Name, nil
			}
		}
		return "", errors.New("could not determine window name")
	}
}

const goalAddress = `/( \?( |$)|{!.*!})`

// Sets dot to the the next goal
func NextGoal(win *acme.Win) error {
	err := win.Ctl("addr=dot")
	if err != nil {
		return err
	}
	err = win.Addr(goalAddress) // The regex might not be correct every time, but for now we hope it suffice.
	if err != nil {
		return err
	}
	err = win.Ctl("dot=addr")
	if err != nil {
		return err
	}
	return win.Ctl("show")
}

func ReplaceSelection(win *acme.Win, text string) error {
	err := win.Ctl("addr=dot")
	if err != nil {
		return err
	}
	_, err = win.Write("data", []byte(text))
	return err
}

// For some reasons, I do not understand yet, writing the address the first
// time has no effect. After calling this function everything works as I expect.
func ResetAddr(win *acme.Win) error {
	return win.Addr("#0")
}

// Returns the InteractionPoint which is selected by dot, i.e dot is completely included in the range of the interaction point.
// Sets addr to dot.
func SelectedInteractionPoint(win *acme.Win, interactionPoints []InteractionId) (InteractionId, error) {
	if err := win.Ctl("addr=dot"); err != nil {
		return InteractionId{}, errors.New(fmt.Sprintf("could not set addr to dot %s", err))
	}
	selectionStart, selectionEnd, err := win.ReadAddr()
	debugPrint("lookup goal selected by %d %d", selectionStart, selectionEnd)
	if err != nil {
		log.Printf("could not read select goal address: %s", err)
		return InteractionId{}, err
	}
	for _, interactionPoint := range interactionPoints {
		debugPrint("comparing interactionPoint with Range %T%v", interactionPoint.Range, interactionPoint.Range)
		if interactionPoint.Range[0].Start.Pos-1 <= selectionStart && selectionEnd <= interactionPoint.Range[0].End.Pos-1 {
			return interactionPoint, nil
		}
	}
	return InteractionId{}, errors.New("No interaction point selected. Maybe you need to reload the file.")
}

func SelectedInteraction(win *acme.Win, interactionPoints []InteractionId) (interactionPoint InteractionId, interactionContent string, err error) {
	// which goal is hit by selection?
	interactionPoint, err = SelectedInteractionPoint(win, interactionPoints)
	if err != nil {
		err = fmt.Errorf("move dot inside a goal: %w", err)
		return
	}
	debugPrint("set address to #%d,#%d", interactionPoint.Range[0].Start.Pos-1, interactionPoint.Range[0].End.Pos-1)
	err = win.Addr("#%d,#%d", interactionPoint.Range[0].Start.Pos-1, interactionPoint.Range[0].End.Pos-1)
	if err != nil {
		err = fmt.Errorf("could not set interactionPoint address: %s", err)
		return
	}
	err = win.Ctl("dot=addr")
	if err != nil {
		err = fmt.Errorf("could set dot to interactionPoint: %s", err)
		return
	}
	debugPrint("read selection")
	interactionContent = win.Selection()
	if interactionContent == "?" {
		interactionContent = ""
	} else {
		interactionContent = interactionContent[2 : len(interactionContent)-2] // drop {! and !}
	}
	return
}