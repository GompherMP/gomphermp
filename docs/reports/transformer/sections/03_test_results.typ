= Resultados de Ejecución

== Resumen cuantitativo

#figure(
  table(
    columns: (auto, auto),
    align: (left, right),
    [*Métrica*],                                            [*Valor*],
    [Total de pruebas ejecutadas],                          [190],
    [Pruebas exitosas],                                     [190],
    [Pruebas fallidas],                                     [0],
    [Pruebas exitosas con detector de carreras activo],     [190],
    [Cobertura total de instrucciones],                     [98.5%],
  ),
  caption: [Resumen cuantitativo de la ejecución de la suite],
)

== Distribución de pruebas por área

#figure(
  table(
    columns: (auto, auto, 1fr),
    align: (left, right, left),
    table.header([*Área*], [*Cantidad*], [*Propósito*]),
    [Parallel, for y parallel for], [51], [Verifican la reescritura de las regiones y bucles paralelos, las tres políticas de planificación y la normalización de la forma canónica del bucle.],
    [Sections y parallel sections], [17], [Verifican la generación de una closure por sección y la distribución de los bloques independientes.],
    [Single, master, critical y barrier], [30], [Verifican las construcciones de sincronización, incluyendo las regiones críticas nominales y el punto de barrera.],
    [Atomic], [22], [Verifican la traducción de las operaciones atómicas de actualización, escritura y lectura a las primitivas del runtime.],
    [Task, taskwait, taskgroup y taskloop], [49], [Verifican el paralelismo de tareas, las dependencias de datos y la generación de tareas por bloque de iteraciones.],
    [Infraestructura de AST y validación], [21], [Verifican los constructores y mutadores de AST, la resolución de tipos, la validación contextual y el despacho.],
  ),
  caption: [Distribución de pruebas por área del módulo],
)

== Cobertura detallada por archivo

#figure(
  table(
    columns: (auto, auto, auto),
    align: (left, right, right),
    [*Archivo*],            [*Funciones*], [*Cobertura*],
    [`transformer.go`],     [1],   [96.0%],
    [`validate.go`],        [3],   [100.0%],
    [`parallel.go`],        [1],   [100.0%],
    [`loop.go`],            [8],   [97.4%],
    [`loopform.go`],        [8],   [100.0%],
    [`sections.go`],        [4],   [100.0%],
    [`single.go`],          [1],   [100.0%],
    [`master.go`],          [1],   [100.0%],
    [`critical.go`],        [1],   [100.0%],
    [`barrier.go`],         [1],   [100.0%],
    [`atomic.go`],          [5],   [100.0%],
    [`task.go`],            [3],   [91.4%],
    [`taskloop.go`],        [1],   [96.6%],
    [`clauses.go`],         [22],  [100.0%],
    [`typeinfo.go`],        [1],   [94.4%],
    [`astbuild.go`],        [18],  [100.0%],
    [`astreplace.go`],      [8],   [99.2%],
    [`imports.go`],         [3],   [100.0%],
    [*Total del módulo*],   [*90*], [*98.5%*],
  ),
  caption: [Cobertura agregada por archivo del módulo],
)

== Funciones por debajo del 100%

La cobertura restante corresponde a guardas defensivas: ramas de retorno de error
para condiciones que el flujo del programa no alcanza en operación normal (por
ejemplo, un nodo de AST que no es del tipo esperado, o un nodo objetivo ausente
del árbol tras haberse extraído de él). Se mantienen por robustez aunque ningún
programa de entrada válido las dispara.

#figure(
  ```
internal/transformer/transformer.go:13:    Transform                  96.0%
internal/transformer/loop.go:39:           transformLoopDirective     96.7%
internal/transformer/loop.go:115:          transformLoopWithClauses   95.5%
internal/transformer/task.go:18:           transformTask              92.3%
internal/transformer/task.go:132:          transformTaskwait          80.0%
internal/transformer/taskloop.go:18:       transformTaskloop          96.6%
internal/transformer/typeinfo.go:21:       resolveVarType             94.4%
internal/transformer/astreplace.go:104:    replaceStmtWithPrefix      95.2%
  ```,
  caption: [Funciones con guardas defensivas no alcanzadas por la suite],
)
