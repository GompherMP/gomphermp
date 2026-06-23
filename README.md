# GompherMP

An implementation of parallelism in the Go programming language using OpenMP-like clauses.

## About This Project
GompherMP is a source-to-source compiler for Go. You annotate ordinary Go code with `//gompher` comment directives, and the tool rewrites the program into plain Go that calls the GompherMP runtime, so it compiles and runs with the standard Go toolchain. The goal is explicit, structured concurrent execution on top of Go's native goroutines and synchronization primitives.

## Repository Structure
* `/cmd/gompher`: the command-line tool.
* `/internal/parser`: extracts the `//gompher` directives and clauses from the source.
* `/internal/transformer`: rewrites the AST, injecting the runtime calls.
* `/internal/printer`: serializes the transformed AST back to Go source.
* `/pkg/runtime`: the concurrency runtime the generated code calls.
* `/examples`: runnable example programs, one per construct.
* `/docs/specs`: technical specification of the directives, formal syntax, and runtime behavior.
* `/docs/reports`: test-coverage reports per module.

## Installation

**Prerequisites:** Go 1.22.2 or newer.

Build the `gompher` binary from the repository root:

```bash
make build      # produces ./gompher
```

Or with the Go toolchain directly:

```bash
go build -o gompher ./cmd/gompher/main.go
```

Optionally place the binary on your `PATH` (for example, `mv gompher /usr/local/bin/`) so you can call `gompher` from anywhere.

## Usage

Annotate a Go file with `//gompher` directives, then transpile and compile it in one step:

```bash
gompher build path/to/program.go
```

This parses the directives, rewrites the program, and runs `go build` on the result, producing an executable named after the source file (here, `./program`).

### Options

| Flag | Description |
|---|---|
| `-o`, `--output <path>` | Output binary path (defaults to `./<file>`). |
| `-v`, `--verbose` | Print the pipeline phases and the directives detected. |
| `-k`, `--keep-temp` | Keep the generated intermediate `.go` file (useful for inspecting the output). |
| `-h`, `--help` | Show usage and the available flags. |
| `--version` | Print the tool version. |

### Example

Input (`hello.go`):

```go
package main

import "fmt"

func main() {
	//gompher parallel
	{
		fmt.Println("hello from the parallel team")
	}
	fmt.Println("all goroutines done")
}
```

Transpile, compile, and run:

```bash
gompher build hello.go
./hello
```

To inspect the generated Go source, add `-k` and open the `gompher_*.go` file left next to the input.

## Writing GompherMP directives

A directive is a `//gompher` comment placed immediately above the statement it applies to. The supported directives are:

* Structured parallelism: `parallel`, `for`, `parallel for`, `sections` / `section`, `parallel sections`, `single`, `master`.
* Synchronization: `critical [name]`, `barrier`, `atomic`.
* Tasking: `task`, `taskwait`, `taskgroup`, `taskloop`.

Data-sharing clauses (`private`, `firstprivate`, `shared`, `lastprivate`, `reduction`), the `schedule` clause for loops, `depend` for tasks, and `grainsize` for `taskloop` are written after the directive, for example:

```go
//gompher parallel for reduction(+:sum) schedule(dynamic, 4)
for i := 0; i < n; i++ {
	sum += work(i)
}
```

The `/examples` directory contains a runnable program for each construct, and `/docs/specs` documents the full syntax and semantics.

## Development

```bash
make deps       # first-time setup
make build      # build ./gompher
make test       # run the full test suite
```

## Authors
* Jorge David Alejandro Contreras
* Patricia Natividad Cántaro Márquez
