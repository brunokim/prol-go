package profiler

import (
	"compress/gzip"
	"fmt"
	"os"

	profilepb "github.com/brunokim/prol-go/proto/profile"
	"google.golang.org/protobuf/proto"
)

type Location struct {
	FuncName   string
	LineNumber int
}

type Profiler struct {
	profile         *profilepb.Profile
	stack           []Location
	stringIndices   map[string]int
	locationIndices map[Location]int
	functionIndices map[string]int
}

type Type struct {
	Name string
	Unit string
}

func NewProfiler(sampleTypes ...Type) (*Profiler, error) {
	if len(sampleTypes) == 0 {
		return nil, fmt.Errorf("empty sample types")
	}
	p := &Profiler{
		profile:         &profilepb.Profile{},
		stringIndices:   make(map[string]int),
		locationIndices: make(map[Location]int),
		functionIndices: make(map[string]int),
	}
	p.stringIndex("")
	// Insert sample types.
	p.profile.SampleType = make([]*profilepb.ValueType, len(sampleTypes))
	for i, sampleType := range sampleTypes {
		p.profile.SampleType[i] = &profilepb.ValueType{
			Type: int64(p.stringIndex(sampleType.Name)),
			Unit: int64(p.stringIndex(sampleType.Unit)),
		}
	}
	return p, nil
}

func (p *Profiler) WriteFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	bs, err := proto.Marshal(p.profile)
	if err != nil {
		return err
	}
	zw, err := gzip.NewWriterLevel(f, gzip.BestSpeed)
	if err != nil {
		return err
	}
	_, err = zw.Write(bs)
	if err != nil {
		return err
	}
	err = zw.Close()
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	return nil
}

func (p *Profiler) Enter(loc Location) {
	p.stack = append(p.stack, loc)
}

func (p *Profiler) Exit() {
	p.stack = p.stack[:len(p.stack)-1]
}

func (p *Profiler) DoSample(values ...int64) {
	n := len(p.profile.SampleType)
	if len(values) != n {
		// TODO: notify about error.
		return
	}
	sample := &profilepb.Sample{
		LocationId: make([]uint64, len(p.stack)),
		Value:      make([]int64, n),
	}
	// Insert locations in reverse order.
	for i, loc := range p.stack {
		sample.LocationId[len(p.stack)-i-1] = uint64(p.locationID(loc))
	}
	copy(sample.Value, values)
	p.profile.Sample = append(p.profile.Sample, sample)
}

func (p *Profiler) stringIndex(s string) int {
	i, ok := p.stringIndices[s]
	if !ok {
		i = len(p.profile.StringTable)
		p.stringIndices[s] = i
		p.profile.StringTable = append(p.profile.StringTable, s)
	}
	return i
}

func (p *Profiler) locationID(loc Location) int {
	i, ok := p.locationIndices[loc]
	if !ok {
		i = len(p.profile.Location)
		p.locationIndices[loc] = i
		p.profile.Location = append(p.profile.Location, &profilepb.Location{
			Id: uint64(i + 1),
			Line: []*profilepb.Line{
				{
					FunctionId: uint64(p.functionID(loc.FuncName)),
					Line:       int64(loc.LineNumber),
				},
			},
		})
	}
	return i + 1
}

func (p *Profiler) functionID(funcName string) int {
	i, ok := p.functionIndices[funcName]
	if !ok {
		i = len(p.profile.Function)
		p.functionIndices[funcName] = i
		p.profile.Function = append(p.profile.Function, &profilepb.Function{
			Id:   uint64(i + 1),
			Name: int64(p.stringIndex(funcName)),
		})
	}
	return i + 1
}
