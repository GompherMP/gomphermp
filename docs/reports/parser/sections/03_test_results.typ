= Resultados de Ejecución

La verificación de la suite se realizó en dos pasos. Primero se ejecutó `go test -coverprofile=parser_cov.out ./internal/parser/...`, que ejecuta las pruebas y genera un archivo de perfil con los datos crudos de cobertura. Luego, ese archivo se procesó con `go tool cover -func=parser_cov.out` para obtener la tabla de cobertura por función presentada en la sección 3.4.

== Resumen cuantitativo

#figure(
  table(
    columns: (auto, auto),
    align: (left, right),
    [*Métrica*],                                            [*Valor*],
    [Total de pruebas ejecutadas],                          [90],
    [Pruebas exitosas],                                     [90],
    [Pruebas fallidas],                                     [0],
    [Cobertura total de instrucciones],                     [100.0%],
    [Archivos del módulo con cobertura del 100%],           [4 de 4],
    [Funciones del módulo con cobertura del 100%],          [53 de 53],
  ),
  caption: [Resumen cuantitativo de la ejecución de la suite],
)

== Distribución de pruebas por sección

#figure(
  table(
    columns: (auto, auto, 1fr),
    align: (left, right, left),
    [*Sección*],                            [*Cantidad*], [*Propósito*],
    [Parseo de cláusulas],                  [19],         [Verifican el parseo léxico de cada tipo de cláusula en aislamiento mediante `extractClauses`.],
    [Parseo de directivas],                 [39],         [Verifican la construcción de cada tipo de directiva con sus combinaciones de cláusulas mediante `parseDirectiveText`.],
    [Integración completa],                 [15],         [Verifican el flujo completo de `Parse` sobre programas Go reales con directivas anotadas.],
    [Validaciones semánticas],              [10],         [Verifican el rechazo de errores semánticos (tipo de nodo incorrecto, comentarios desalineados, contexto inválido, cláusulas vacías).],
    [Contratos de interfaz],                [3],          [Verifican los métodos no exportados de las interfaces `Directive` y `Clause` para todos los tipos concretos.],
    [Pruebas internas auxiliares],          [4],          [Cubren rutas de error en funciones internas (`validateClauses`, `makeVarListClause`, `buildDirective`).],
  ),
  caption: [Distribución de pruebas por sección del archivo `parser_test.go`],
)

== Cobertura detallada por archivo

#figure(
  table(
    columns: (auto, auto, auto),
    align: (left, right, right),
    [*Archivo*],                       [*Funciones*], [*Cobertura*],
    [`internal/parser/parser.go`],        [13],          [100.0%],
    [`internal/parser/directive.go`],     [28],          [100.0%],
    [`internal/parser/clause.go`],        [8],           [100.0%],
    [`internal/parser/clause_parser.go`], [4],           [100.0%],
    [*Total del módulo*],                 [*53*],        [*100.0%*],
  ),
  caption: [Cobertura agregada por archivo del módulo],
)

== Salida directa de la herramienta de cobertura

A continuación se reproduce la salida completa del comando `go tool cover -func=parser_cov.out`. Cada línea representa una función o método del módulo con su porcentaje de cobertura individual. Los métodos `directiveKind` y `line` aparecen una vez por cada uno de los 14 tipos de directiva, y `clauseKind` aparece una vez por cada uno de los 8 tipos de cláusula.

#figure(
  ```
internal/parser/clause.go:28:   clauseKind             100.0%
internal/parser/clause.go:35:   clauseKind             100.0%
internal/parser/clause.go:42:   clauseKind             100.0%
internal/parser/clause.go:48:   clauseKind             100.0%
internal/parser/clause.go:58:   clauseKind             100.0%
internal/parser/clause.go:67:   clauseKind             100.0%
internal/parser/clause.go:77:   clauseKind             100.0%
internal/parser/clause.go:83:   clauseKind             100.0%
internal/parser/clause_parser.go:34:    extractClauses         100.0%
internal/parser/clause_parser.go:56:    parseNextClause        100.0%
internal/parser/clause_parser.go:120:   makeVarListClause      100.0%
internal/parser/clause_parser.go:137:   splitVars              100.0%
internal/parser/directive.go:52:   directiveKind          100.0%
internal/parser/directive.go:53:   line                   100.0%
internal/parser/directive.go:63:   directiveKind          100.0%
internal/parser/directive.go:64:   line                   100.0%
internal/parser/directive.go:74:   directiveKind          100.0%
internal/parser/directive.go:75:   line                   100.0%
internal/parser/directive.go:85:   directiveKind          100.0%
internal/parser/directive.go:86:   line                   100.0%
internal/parser/directive.go:95:   directiveKind          100.0%
internal/parser/directive.go:96:   line                   100.0%
internal/parser/directive.go:106:  directiveKind          100.0%
internal/parser/directive.go:107:  line                   100.0%
internal/parser/directive.go:116:  directiveKind          100.0%
internal/parser/directive.go:117:  line                   100.0%
internal/parser/directive.go:127:  directiveKind          100.0%
internal/parser/directive.go:128:  line                   100.0%
internal/parser/directive.go:137:  directiveKind          100.0%
internal/parser/directive.go:138:  line                   100.0%
internal/parser/directive.go:148:  directiveKind          100.0%
internal/parser/directive.go:149:  line                   100.0%
internal/parser/directive.go:159:  directiveKind          100.0%
internal/parser/directive.go:160:  line                   100.0%
internal/parser/directive.go:169:  directiveKind          100.0%
internal/parser/directive.go:170:  line                   100.0%
internal/parser/directive.go:179:  directiveKind          100.0%
internal/parser/directive.go:180:  line                   100.0%
internal/parser/directive.go:190:  directiveKind          100.0%
internal/parser/directive.go:191:  line                   100.0%
internal/parser/parser.go:28:      Parse                  100.0%
internal/parser/parser.go:51:      setNode                100.0%
internal/parser/parser.go:96:      getDirectiveLine       100.0%
internal/parser/parser.go:103:     extractAnnotatedNodes  100.0%
internal/parser/parser.go:167:     directiveRequiresNode  100.0%
internal/parser/parser.go:178:     validateNodeType       100.0%
internal/parser/parser.go:201:     validateSectionContext 100.0%
internal/parser/parser.go:230:     isGompherComment       100.0%
internal/parser/parser.go:235:     parseGompherComment    100.0%
internal/parser/parser.go:248:     parseDirectiveText     100.0%
internal/parser/parser.go:263:     buildDirective         100.0%
internal/parser/parser.go:395:     extractKind            100.0%
internal/parser/parser.go:446:     validateClauses        100.0%
total:                             (statements)           100.0%
  ```,
  caption: [Salida completa del comando `go tool cover -func=parser_cov.out`],
)

