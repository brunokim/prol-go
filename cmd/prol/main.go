package main

import (
	"fmt"
	"io"
	"iter"
	"log"
	"strings"

	"github.com/brunokim/prol-go/kif"
	"github.com/brunokim/prol-go/prol"
	"github.com/ergochat/readline"
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

func (s *shell) setQueryState() {
	s.state = queryState
	s.nextSolution = nil
	s.stopSolutions = nil
	s.solutionErrFn = nil
	s.rl.SetPrompt("?- ")
}

func (s *shell) setSolutionsState(solutions iter.Seq[prol.Solution], errFn func() error) {
	s.nextSolution, s.stopSolutions = iter.Pull(solutions)
	s.solutionErrFn = errFn
	s.state = solutionsState
	s.rl.SetPrompt("")
}

func (s *shell) readQuery(text string) error {
	query, err := s.db.Query(text)
	if err != nil {
		return err
	}
	s.setSolutionsState(s.db.Solve(query.(prol.Clause)))
	return s.printSolution()
}

func (s *shell) printf(format string, args ...any) {
	fmt.Fprintf(s.rl.Stdout(), format, args...)
}

func (s *shell) printSolution() error {
	solution, ok := s.nextSolution()
	if !ok {
		return s.abortSolutions()
	}
	if len(solution) == 0 {
		s.printf("yes ")
		return nil
	}
	s.printf("%v ", solution)
	return nil
}

func (s *shell) abortSolutions() error {
	err := s.solutionErrFn()
	s.setQueryState()
	return err
}

func (s *shell) mainLoop() {
	for {
		text, err := s.rl.ReadLine()
		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			break
		}
		if text == "" {
			continue
		}
		switch s.state {
		case queryState:
			if err := s.readQuery(text); err != nil {
				log.Println(err)
			}
		case solutionsState:
			switch strings.TrimSpace(text) {
			case ";":
				if err := s.printSolution(); err != nil {
					log.Println(err)
				}
			case ".":
				if err := s.abortSolutions(); err != nil {
					log.Println(err)
				}
			case "exit":
				return
			default:
				s.rl.SetPrompt(";) continue .) stop ")
			}
		}
	}
}

func main() {
	fmt.Println("prol shell (press Ctrl+D, or type 'exit' to exit)")
	db := prol.Prelude()
	db.Logger = kif.NewStderrLogger()
	db.Logger.LogLevel = kif.INFO
	shell := &shell{db: db}
	// Configure readline.
	rl, err := readline.NewFromConfig(&readline.Config{
		HistoryFile: ".prol-history",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer rl.Close()
	rl.CaptureExitSignal()
	log.SetOutput(rl.Stderr())
	shell.rl = rl
	shell.setQueryState()
	shell.mainLoop()
}
