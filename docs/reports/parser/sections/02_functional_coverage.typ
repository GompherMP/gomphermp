= Cobertura Funcional

Esta sección demuestra que cada una de las directivas y cláusulas declaradas en la especificación del lenguaje GompherMP cuenta con al menos una prueba que verifica su correcto análisis sintáctico.

Por brevedad, los nombres de las pruebas en las tablas omiten su prefijo: las pruebas que comienzan con `Parse` corresponden a `TestParse_*` en `parser_test.go`, las que comienzan con `Directive` corresponden a `TestParseDirectiveText_*`, y las que comienzan con `Clauses` corresponden a `TestExtractClauses_*`.

== Directivas soportadas

#figure(
  table(
    columns: (auto, auto, 1fr),
    align: (left + horizon, left + horizon, left + horizon),
    table.header([*Directiva*], [*Tipo de prueba*], [*Prueba*]),

    table.cell(rowspan: 2)[`parallel`],     [Integración], [`Parse/ParallelBlock`],
                                            [Unitaria],    [`Directive/Parallel`],

    table.cell(rowspan: 2)[`for`],          [Integración], [`Parse/For`],
                                            [Unitaria],    [`Directive/For`],

    table.cell(rowspan: 2)[`parallel for`], [Integración], [`Parse/ParallelFor`],
                                            [Unitaria],    [`Directive/ParallelFor`],

    table.cell(rowspan: 2)[`sections`],     [Integración], [`Parse/Sections`],
                                            [Unitaria],    [`Directive/SectionsInvalidClause`],

    table.cell(rowspan: 2)[`section`],      [Integración], [`Parse/Sections`],
                                            [Unitaria],    [`Directive/SectionRejectsClause`],

    table.cell(rowspan: 2)[`single`],       [Integración], [`Parse/Single`],
                                            [Unitaria],    [`Directive/Single`],

    table.cell(rowspan: 2)[`master`],       [Integración], [`Parse/Master`],
                                            [Unitaria],    [`Directive/MasterRejectsClause`],

    table.cell(rowspan: 3)[`critical`],     [Integración], [`Parse/Critical`],
                                            [Unitaria],    [`Directive/CriticalNamed`],
                                            [Unitaria],    [`Directive/CriticalAnonymous`],

    table.cell(rowspan: 2)[`barrier`],      [Integración], [`Parse/MultipleDirectives`],
                                            [Unitaria],    [`Directive/Barrier`],

    table.cell(rowspan: 6)[`atomic`],       [Integración], [`Parse/Atomic`],
                                            [Integración], [`Parse/AtomicOnAssignStmt`],
                                            [Unitaria],    [`Directive/AtomicUpdate`],
                                            [Unitaria],    [`Directive/AtomicRead`],
                                            [Unitaria],    [`Directive/AtomicWrite`],
                                            [Unitaria],    [`Directive/AtomicDefaultMode`],

    table.cell(rowspan: 2)[`task`],         [Integración], [`Parse/Tasks`],
                                            [Unitaria],    [`Directive/Task`],

    table.cell(rowspan: 2)[`taskwait`],     [Integración], [`Parse/Tasks`],
                                            [Unitaria],    [`Directive/Taskwait`],

    table.cell(rowspan: 2)[`taskgroup`],    [Integración], [`Parse/Taskgroup`],
                                            [Unitaria],    [`Directive/Taskgroup`],

    table.cell(rowspan: 2)[`taskloop`],     [Integración], [`Parse/Taskloop`],
                                            [Unitaria],    [`Directive/Taskloop`],
  ),
  caption: [Mapeo de directivas a pruebas que verifican su análisis sintáctico],
)

== Cláusulas soportadas

#figure(
  table(
    columns: (auto, 1fr),
    align: (left + horizon, left + horizon),
    table.header([*Cláusula*], [*Prueba*]),

    [`private`],      [`Clauses/Private`],
    [`firstprivate`], [`Clauses/FirstPrivate`],
    [`lastprivate`],  [`Clauses/LastPrivate`],
    [`shared`],       [`Clauses/Shared`],

    table.cell(rowspan: 3)[`reduction`], [`Clauses/Reduction_Sum`],
                                         [`Clauses/Reduction_And`],
                                         [`Clauses/Reduction_Max`],

    table.cell(rowspan: 4)[`schedule`],  [`Clauses/ScheduleStatic`],
                                         [`Clauses/ScheduleStaticNoChunk`],
                                         [`Clauses/ScheduleDynamic`],
                                         [`Clauses/ScheduleDynamicNoChunk`],

    table.cell(rowspan: 3)[`depend`],    [`Clauses/Depend_In`],
                                         [`Clauses/Depend_Out`],
                                         [`Clauses/Depend_Inout`],

    [`grainsize`],    [`Clauses/Grainsize`],
  ),
  caption: [Mapeo de cláusulas a pruebas que verifican su análisis sintáctico],
)

== Combinaciones inválidas rechazadas

Además de verificar el análisis correcto, la suite prueba que combinaciones inválidas de directivas y cláusulas son rechazadas con un mensaje de error explícito antes de cualquier transformación.

#figure(
  table(
    columns: (1fr, 1fr),
    align: (left + horizon, left + horizon),
    table.header([*Combinación inválida*], [*Prueba que verifica el rechazo*]),

    [`for` con `reduction`],         [`Directive/ForRejectsReduction`],
    [`single` con `shared`],         [`Directive/SingleRejectsShared`],
    [`parallel` con `depend`],       [`Directive/ParallelRejectsDepend`],
    [`parallel for` con `depend`],   [`Directive/ParallelForRejectsDepend`],
    [`sections` con `depend`],       [`Directive/SectionsRejectsDepend`],
    [`task` con `schedule`],         [`Directive/TaskRejectsSchedule`],
    [`taskloop` con `shared`],       [`Directive/TaskloopRejectsShared`],

    table.cell(rowspan: 5)[Directivas sin cláusulas (`barrier`, `master`, `section`, `taskwait`, `taskgroup`) con cualquier cláusula],
                                     [`Directive/BarrierRejectsClause`],
                                     [`Directive/MasterRejectsClause`],
                                     [`Directive/SectionRejectsClause`],
                                     [`Directive/TaskwaitRejectsClause`],
                                     [`Directive/TaskgroupRejectsClause`],

    [`atomic` con modo inválido],    [`Directive/AtomicInvalidMode`],

    table.cell(rowspan: 2)[`critical` con nombre malformado o vacío],
                                     [`Directive/CriticalMalformedName`],
                                     [`Directive/CriticalEmptyName`],

    table.cell(rowspan: 2)[Cláusula con paréntesis vacíos],
                                     [`Clauses/EmptyPrivate`],
                                     [`Clauses/EmptyShared`],

    table.cell(rowspan: 2)[Directiva desconocida],
                                     [`Directive/Unknown`],
                                     [`Parse/InvalidGompherDirective`],

    table.cell(rowspan: 7)[Cláusula desconocida en cada directiva con cláusulas],
                                     [`Directive/ParallelInvalidClause`],
                                     [`Directive/ForInvalidClause`],
                                     [`Directive/ParallelForInvalidClause`],
                                     [`Directive/SectionsInvalidClause`],
                                     [`Directive/SingleInvalidClause`],
                                     [`Directive/TaskInvalidClause`],
                                     [`Directive/TaskloopInvalidClause`],
  ),
  caption: [Combinaciones inválidas verificadas como rechazadas por el parser],
)
