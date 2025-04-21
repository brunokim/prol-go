package main

import (
	"fmt"
	"io"
	"iter"
	"log"

	"github.com/brunokim/prol-go/kif"
	"github.com/brunokim/prol-go/prol"
	"github.com/chzyer/readline"
)

type shellState int

const (
	queryState shellState = iota
	solutionsState
)

type shell struct {
	state shellState
	db    *prol.Database
	rl    *readline.Instance
	// State for solutionsState
	nextSolution  func() (prol.Solution, bool)
	stopSolutions func()
	solutionErrFn func() error
}

func (s *shell) prompt() (string, error) {
	switch s.state {
	case queryState:
		s.rl.SetPrompt("?- ")
	case solutionsState:
		s.rl.SetPrompt(";) continue .) stop ")
	}
	return s.rl.Readline()
}

func (s *shell) readQuery(text string) error {
	query, err := s.db.Query(text)
	if err != nil {
		return err
	}
	solutions, errFn := s.db.Solve(query.(prol.Clause))
	s.nextSolution, s.stopSolutions = iter.Pull(solutions)
	s.solutionErrFn = errFn
	s.state = solutionsState
	return s.printSolution()
}

func (s *shell) printSolution() error {
	solution, ok := s.nextSolution()
	if !ok {
		return s.abortSolutions()
	}
	if len(solution) == 0 {
		fmt.Fprint(s.rl.Stdout(), "yes ")
		return nil
	}
	fmt.Fprintf(s.rl.Stdout(), "%v ", solution)
	return nil
}

func (s *shell) abortSolutions() error {
	err := s.solutionErrFn()
	s.nextSolution, s.stopSolutions, s.solutionErrFn = nil, nil, nil
	s.state = queryState
	return err
}

func main() {
	fmt.Println("prol shell (Ctrl+D to exit)")
	db := prol.Prelude()
	db.Logger = kif.NewStderrLogger()
	db.Logger.LogLevel = kif.INFO
	rl, err := readline.NewEx(&readline.Config{
		HistoryFile: ".prol-history",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer rl.Close()
	rl.CaptureExitSignal()
	log.SetOutput(rl.Stdout())
	shell := &shell{state: queryState, db: db, rl: rl}
	for {
		text, err := shell.prompt()
		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			break
		}
		if text == "" {
			continue
		}
		switch shell.state {
		case queryState:
			if err := shell.readQuery(text); err != nil {
				log.Println(err)
			}
		case solutionsState:
			switch text {
			case ";":
				if err := shell.printSolution(); err != nil {
					log.Println(err)
				}
			case ".":
				if err := shell.abortSolutions(); err != nil {
					log.Println(err)
				}
			default:
				fmt.Println("input error: expecting ';' or '.'")
			}
		}
	}
}
