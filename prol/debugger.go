package prol

import (
	"log"
)

type debugger struct {
	breakpoints map[Indicator]struct{}
}

func newDebugger() *debugger {
	return &debugger{
		breakpoints: make(map[Indicator]struct{}),
	}
}

func (dbg *debugger) putBreakpoint(ind Indicator) {
	dbg.breakpoints[ind] = struct{}{}
}

func (dbg *debugger) clearBreakpoint(ind Indicator) {
	delete(dbg.breakpoints, ind)
}

func (dbg *debugger) checkBreakpoint(ind Indicator) {
	if dbg == nil {
		return
	}
	if _, ok := dbg.breakpoints[ind]; !ok {
		return
	}
	log.Println("should break at", ind)
}
