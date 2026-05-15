# Parser

The parser is the compiler frontend for GompherMP. It takes raw Go source code as input, builds a standard Go Abstract Syntax Tree (AST), and extracts all `//gompher` pragma directives — pairing each one with the Go syntax node it annotates.

The output is a `ParseResult` that the transformer consumes to rewrite the source into concurrent Go code.

---

## Pipeline

```
 Raw Go source
        |
        v
 go/parser.ParseFile()
        Builds the full Go AST and collects all comments.
        |
        v
 ast.NewCommentMap()
        Associates each comment to the nearest Go AST node it precedes.
        |
        v
 AST walk + directive extraction
        Scans for //gompher comments, parses each into a typed Directive struct,
        and pairs it with its corresponding Go AST node.
        |
        v
 Semantic validation
        Verifies node type compatibility, comment adjacency, and clause validity
        for each directive.
        |
        v
 Hierarchical context validation
        Verifies that every //gompher section is contained inside a //gompher sections block.
        |
        v
 Sort by source line
        Restores top-to-bottom execution order.
        |
        v
 ParseResult{ FileSet, File, []AnnotatedNode }
```

---

## Key Types

### `ParseResult`

The value returned by `Parse()`. Contains everything the transformer needs.

| Field | Type | Description |
|---|---|---|
| `FileSet` | `*token.FileSet` | Spatial index mapping token positions to source file locations |
| `File` | `*ast.File` | Root node of the full Go AST |
| `Nodes` | `[]AnnotatedNode` | All extracted GompherMP directives, in source order |

---

### `AnnotatedNode`

The fundamental unit produced by the parser. Each instance pairs a parsed directive with its location in the Go AST.

| Field | Type | Description |
|---|---|---|
| `Directive` | `Directive` | The parsed directive (e.g. `ParallelDirective`, `ForDirective`) |

The `Directive` itself holds the corresponding `ast.Node` — the block or statement the directive governs.

---

### `Directive`

An interface implemented by all 14 GompherMP directive types. Each concrete type embeds a `pos` (source position and line number) and, where applicable, an `ast.Node` pointing to the Go syntax it annotates.

| Directive | Attached Node | Accepted Clauses |
|---|---|---|
| `ParallelDirective` | `*ast.BlockStmt` | private, firstprivate, shared |
| `ForDirective` | `*ast.ForStmt` | private, firstprivate, schedule |
| `ParallelForDirective` | `*ast.ForStmt` | private, firstprivate, lastprivate, shared, reduction, schedule |
| `SectionsDirective` | `*ast.BlockStmt` | private, firstprivate, lastprivate, reduction |
| `SectionDirective` | `*ast.BlockStmt` | none |
| `SingleDirective` | `*ast.BlockStmt` | private, firstprivate |
| `MasterDirective` | `*ast.BlockStmt` | none |
| `CriticalDirective` | `*ast.BlockStmt` | none (optional lock name) |
| `BarrierDirective` | none | none |
| `AtomicDirective` | `*ast.ExprStmt` / `*ast.AssignStmt` / `*ast.IncDecStmt` | none (optional mode) |
| `TaskDirective` | `*ast.BlockStmt` | private, firstprivate, depend |
| `TaskwaitDirective` | none | none |
| `TaskgroupDirective` | `*ast.BlockStmt` | none |
| `TaskloopDirective` | `*ast.ForStmt` | private, firstprivate, grainsize |

`BarrierDirective` and `TaskwaitDirective` carry no `ast.Node` — they are synchronization points with no associated code block.

---

### `Clause`

An interface implemented by all clause types. Clauses are the arguments attached to a directive and control data-sharing, scheduling, and task dependency behavior.

| Clause | Syntax | Fields |
|---|---|---|
| `PrivateClause` | `private(x, y)` | `Vars []string` |
| `FirstPrivateClause` | `firstprivate(x)` | `Vars []string` |
| `LastPrivateClause` | `lastprivate(x)` | `Vars []string` |
| `SharedClause` | `shared(x, y)` | `Vars []string` |
| `ReductionClause` | `reduction(+:sum)` | `Operator string`, `Vars []string` |
| `ScheduleClause` | `schedule(static, 4)` | `Kind string`, `Chunk string` |
| `DependClause` | `depend(in:x, y)` | `DepType string`, `Vars []string` |
| `GrainsizeClause` | `grainsize(8)` | `Size string` |

---

## Directive Syntax Rules

- A directive must be placed on the line **directly above** the block or statement it annotates, with no blank lines or intervening comments between them.
- `barrier` and `taskwait` are standalone — they do not annotate any block.
- `critical` accepts an optional lock name using parenthesis syntax: `//gompher critical(mylock)`.
- `atomic` accepts an optional mode: `read`, `write`, or `update`. The default when omitted is `update`.
- `section` must appear inside a `sections` block — it cannot be used standalone.
- Clause validation is enforced at parse time. Providing an unsupported clause for a given directive is an error.

---

## Semantic Validations

Beyond parsing the directive text, the parser enforces four categories of semantic rules. Each violation produces a parse-time error with an explicit message and line number.

| Validation | What it checks | Example rejected input |
|---|---|---|
| **Node type** | The directive is attached to a compatible Go AST node | `//gompher for` placed over a `*ast.BlockStmt` instead of a `*ast.ForStmt` |
| **Adjacency** | The directive comment is exactly one line above its target | A blank line between `//gompher parallel` and its `{ ... }` block |
| **Hierarchical context** | `section` directives live inside a `sections` block | `//gompher section` placed at the top of a function body |
| **Non-empty clause arguments** | Variable-list clauses contain at least one variable | `//gompher parallel private()` |

These validations run after the AST walk and before the final sort. A failure short-circuits the pipeline and returns the error to the caller of `Parse`.

---

## Running Tests

The parser ships with a full test suite covering directive parsing, clause parsing, full integration over real Go source, semantic validations, interface contracts, and internal helpers.

### Run all tests

```bash
go test ./internal/parser/...
```

### Run with verbose output

```bash
go test -v ./internal/parser/...
```

Shows each test case name and its result (`PASS` or `FAIL`).

### Generate a coverage profile

```bash
go test -coverprofile=parser_cov.out ./internal/parser/...
```

This runs the tests and writes raw coverage data to `parser_cov.out`. The summary line in stdout reports the overall percentage.

### View per-function coverage

```bash
go tool cover -func=parser_cov.out
```

Prints a table with each function and method in the module alongside its individual coverage percentage.

### View line-by-line coverage as HTML

```bash
go tool cover -html=parser_cov.out -o coverage.html
```

Generates an HTML file showing the source code annotated with covered (green) and uncovered (red) lines. Open `coverage.html` in a browser.

### Run a specific test

```bash
go test -v -run TestParse_ParallelBlock ./internal/parser/
```

The `-run` flag accepts a regex, so `-run "TestParse_.*"` runs every integration test, for example.
