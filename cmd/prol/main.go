package main

import (
	"fmt"
	"io"
	"iter"
	"log"

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
	s.rl.SetPrompt("?- ")
}

func (s *shell) setSolutionsState() {
	s.state = solutionsState
	s.rl.SetPrompt("")
}

func (s *shell) readQuery(text string) error {
	query, err := s.db.Query(text)
	if err != nil {
		return err
	}
	solutions, errFn := s.db.Solve(query.(prol.Clause))
	s.nextSolution, s.stopSolutions = iter.Pull(solutions)
	s.solutionErrFn = errFn
	s.setSolutionsState()
	return s.printSolution()
}

func (s *shell) printf(format string, args ...any) {
	fmt.Fprintf(s.rl.Stdout(), format, args...)
}

func (s *shell) printSolution() error {
	solution, ok := s.nextSolution()
	s.printf("\n")
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
	s.nextSolution, s.stopSolutions, s.solutionErrFn = nil, nil, nil
	s.state = queryState
	return err
}

func (s *shell) runeListener(line []rune, pos int, key rune) ([]rune, int, bool) {
	if s.state != solutionsState {
		// No change.
		return nil, 0, false
	}
	switch key {
	case 0, '\n':
		// First call, or enter, do nothing.
		return nil, 0, false
	case ';':
		if err := s.printSolution(); err != nil {
			log.Println(err)
		}
		return nil, 0, true
	case '.':
		if err := s.abortSolutions(); err != nil {
			log.Println(err)
		}
		return nil, 0, true
	default:
		s.rl.SetPrompt(";) continue .) stop ")
		return nil, 0, true
	}
}

func main() {
	fmt.Println("prol shell (Ctrl+D to exit)")
	db := prol.Prelude()
	db.Logger = kif.NewStderrLogger()
	db.Logger.LogLevel = kif.INFO
	shell := &shell{db: db}
	// Configure readline.
	rl, err := readline.NewFromConfig(&readline.Config{
		HistoryFile: ".prol-history",
		Listener:    shell.runeListener,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer rl.Close()
	rl.CaptureExitSignal()
	log.SetOutput(rl.Stdout())
	shell.rl = rl
	shell.setQueryState()
	// Read loop.
	for {
		text, err := shell.rl.ReadLine()
		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			break
		}
		if text == "" {
			continue
		}
		if shell.state != queryState {
			log.Println("invalid state", shell.state)
			break
		}
		if err := shell.readQuery(text); err != nil {
			log.Println(err)
		}
	}
}
