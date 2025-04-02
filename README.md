# Prol-Go

# Development

Profiling test execution

```
go test github.com/brunokim/prol-go/prol -run TestPreludeLists -cpuprofile=cpu_prof.out -memprofile=mem_prof.out

go tool pprof -http=localhost:8080 -no_browser cpu_prof.out
go tool pprof -http=localhost:8080 -no_browser mem_prof.out
```
