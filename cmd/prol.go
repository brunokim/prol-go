package main

import (
	"bufio"
	"fmt"
	"iter"
	"log"
	"os"

	"github.com/brunokim/prol-go/prol"
)

type shellState int

const (
	queryState shellState = iota
	solutionsState
)

type shell struct {
	state         shellState
	db            *prol.Database
	nextSolution  func() (prol.Solution, bool)
	stopSolutions func()
	solutionErrFn func() error
}

func (s *shell) prompt() {
	if s.state == queryState {
		fmt.Print("?- ")
	}
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
		fmt.Print("yes ")
		return nil
	}
	fmt.Printf("%v ", solution)
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
	scanner := bufio.NewScanner(os.Stdin)
	db := prol.Prelude()
	shell := &shell{state: queryState, db: db}
	shell.prompt()
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			shell.prompt()
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
		shell.prompt()
	}
	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}
