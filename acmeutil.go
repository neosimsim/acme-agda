// Provides convenient funcitons to select and edit text
// in an Acme window.
package main

import (
	"errors"
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

func SelectCurrentLine(win *acme.Win) error {
	err := win.Ctl("addr=dot")
	if err != nil {
		return err
	}
	err = win.Addr("-+")
	if err != nil {
		return err
	}
	err = win.Ctl("dot=addr")
	if err != nil {
		return err
	}
	return win.Ctl("show")
}

// Select the goal "under the cursor". Using backwords and forward search from dot.
func SelectGoal(win *acme.Win) error {
	err := win.Ctl("addr=dot")
	if err != nil {
		return err
	}
	err = win.Addr(`-/{!/,/!}/`) // The regex might not be correct every time, but for now we hope it suffice.
	if err != nil {
		return err
	}
	err = win.Ctl("dot=addr")
	if err != nil {
		return err
	}
	return win.Ctl("show")
}

const goalAddress = `/( \?( |$)|{!.*!})`

func GoalRanges(win *acme.Win) ([]Range, error) {
	err := win.Addr("#0")
	if err != nil {
		return nil, err
	}
	prevStart := 0
	err = win.Addr(goalAddress)
	if err != nil {
		return nil, errors.New("no goal found")
	}
	start, end, err := win.ReadAddr()
	if err != nil {
		return nil, err
	}
	ranges := make([]Range, 0, 10)
	for prevStart < start {
		ranges = append(ranges, Range{Start: start, End: end})
		prevStart = start
		err = win.Addr(goalAddress)
		if err != nil {
			return nil, err
		}
		start, end, err = win.ReadAddr()
		if err != nil {
			return nil, err
		}
	}
	return ranges, nil
}

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
