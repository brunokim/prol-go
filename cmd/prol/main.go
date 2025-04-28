package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"iter"
	"log"
	"os"
	"strings"

	"github.com/brunokim/prol-go/kif"
	"github.com/brunokim/prol-go/prol"
	"github.com/ergochat/readline"
)

var (
	parserName   = flag.String("parser", "prelude", "Parser to use. One of (bootstrap, prelude).")
	consultPaths = flag.String("consult-paths", "", "Comma-separated paths to consult")
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

type dbFunc func(opts ...any) *prol.Database

var (
	parsers = map[string]dbFunc{
		"bootstrap": func(opts ...any) *prol.Database { return prol.Bootstrap() },
		"prelude":   prol.Prelude,
	}
)

func parseCSVRow(text string) ([]string, error) {
	r := csv.NewReader(strings.NewReader(text))
	return r.Read()
}

func parser() *prol.Database {
	dbFn, ok := parsers[*parserName]
	if !ok {
		log.Fatalf("Invalid parser %q", *parserName)
	}
	db := dbFn()
	db.Logger = kif.NewStderrLogger()
	db.Logger.LogLevel = kif.INFO
	return db
}

func consult(db *prol.Database) {
	paths, err := parseCSVRow(*consultPaths)
	if err != nil {
		log.Fatalf("Invalid consult paths: %v", err)
	}
	for _, path := range paths {
		bs, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Could not consult file: %v", err)
			continue
		}
		err = db.Interpret(string(bs))
		if err != nil {
			log.Printf("Failure to consult file: %v", err)
			continue
		}
		log.Printf("Consulted file %s", path)
	}
}

func main() {
	flag.Parse()
	fmt.Println("prol shell (press Ctrl+D, or type 'exit' to exit)")
	db := parser()
	consult(db)
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
