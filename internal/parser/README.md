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
 Orphan rescue pass
        Catches barrier and taskwait directives, which have no attached code block
        and are therefore not found by the CommentMap walk.
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
| `AtomicDirective` | `*ast.ExprStmt` / `*ast.AssignStmt` | none (optional mode) |
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
- Clause validation is enforced at parse time. Providing an unsupported clause for a given directive is an error.
