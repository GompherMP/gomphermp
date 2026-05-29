= Cobertura Funcional

Esta sección demuestra que cada primitiva pública del módulo cuenta con pruebas dedicadas que verifican su comportamiento bajo entradas válidas, casos límite y condiciones de concurrencia.

Por brevedad, los nombres de las pruebas en las tablas siguientes omiten el prefijo `Test`. Las pruebas residen en `pkg/runtime/parallel_test.go` y `pkg/runtime/pool_test.go`.

== Primitivas y pruebas asociadas

#figure(
  table(
    columns: (auto, 1fr),
    align: (left + horizon, left + horizon),
    table.header([*Primitiva*], [*Prueba*]),

    table.cell(rowspan: 6)[`Parallel`],
                                            [`Parallel_AllThreadsExecute`],
                                            [`Parallel_CorrectThreadIDs`],
                                            [`Parallel_ImplicitBarrier`],
                                            [`Parallel_WithClampedPoolSize`],
                                            [`Parallel_NestedSerializes`],
                                            [`Parallel_NestedTeamSizeIsOne`],

    table.cell(rowspan: 6)[`For`],
                                            [`For_BasicExecution`],
                                            [`For_CorrectIterationValues`],
                                            [`For_ZeroIterations`],
                                            [`For_NegativeIterations`],
                                            [`For_WithClampedPoolSize`],
                                            [`For_FewerIterationsThanThreads`],

    table.cell(rowspan: 5)[`ParallelFor`],
                                            [`ParallelFor_DistributesWork`],
                                            [`ParallelFor_CorrectValues`],
                                            [`ParallelFor_ZeroIterations`],
                                            [`ParallelFor_NegativeIterations`],
                                            [`ParallelFor_WithClampedPoolSize`],

    table.cell(rowspan: 10)[`ForDynamic`],
                                            [`ForDynamic_AllIterationsExecute`],
                                            [`ForDynamic_CorrectValues`],
                                            [`ForDynamic_ChunkSizeOne`],
                                            [`ForDynamic_ChunkSizeLargerThanIterations`],
                                            [`ForDynamic_ZeroIterations`],
                                            [`ForDynamic_NegativeIterations`],
                                            [`ForDynamic_InvalidChunkSize`],
                                            [`ForDynamic_WithClampedPoolSize`],
                                            [`ForDynamic_DistributesAcrossGoroutines`],
                                            [`ForDynamic_StressNoRace`],

    table.cell(rowspan: 6)[`Sections`],
                                            [`Sections_AllSectionsExecute`],
                                            [`Sections_FewerSectionsThanThreads`],
                                            [`Sections_MoreSectionsThanThreads`],
                                            [`Sections_EmptyList`],
                                            [`Sections_WithClampedPoolSize`],
                                            [`Sections_DifferentBodies`],

    table.cell(rowspan: 4)[Pool],
                                            [`Pool_InitializedAtStartup`],
                                            [`PoolSize_DefaultsToGOMAXPROCS`],
                                            [`Pool_WorkersReusedAcrossSubmissions`],
                                            [`Pool_WorkersExecuteConcurrently`],

    table.cell(rowspan: 4)[`SetPoolSize`],
                                            [`SetPoolSize_Resizes`],
                                            [`SetPoolSize_NewSizeProcessesJobs`],
                                            [`SetPoolSize_ClampsNonPositive`],
                                            [`SetPoolSize_OldWorkersExitCleanly`],

    table.cell(rowspan: 2)[`CurrentTeamSize`],
                                            [`CurrentTeamSize_OutsideParallel`],
                                            [`CurrentTeamSize_InsideTeam`],
  ),
  caption: [Mapeo de primitivas a pruebas que verifican su funcionamiento],
)

== Comportamientos verificados por categoría

Las pruebas se agrupan en cinco categorías transversales, aplicadas a cada primitiva según corresponda:

#figure(
  table(
    columns: (auto, 1fr),
    align: (left + horizon, left + horizon),
    table.header([*Categoría*], [*Comportamiento verificado*]),

    [Ejecución correcta],
        [Cada iteración o bloque se ejecuta exactamente el número de veces esperado y recibe los valores correctos.],

    [Manejo de casos límite],
        [Las primitivas retornan de forma segura sin ejecutar trabajo cuando reciben entradas degeneradas (cero iteraciones, iteraciones negativas, listas vacías).],

    [Autocorrección de configuración inválida],
        [Cuando `SetPoolSize` o `chunkSize` reciben valores no positivos, el runtime los acota al mínimo seguro de uno (`PoolSize == 1`) sin abortar la ejecución, y las primitivas continúan funcionando correctamente bajo esa configuración degradada.],

    [Persistencia del pool],
        [El pool se pre-instancia durante la inicialización del paquete, los mismos workers atienden jobs de regiones paralelas sucesivas (no se recrean entre invocaciones), y los workers de un pool anterior terminan limpiamente al ser reemplazados por `SetPoolSize`.],

    [Garantías de concurrencia],
        [`Parallel` garantiza la barrera implícita al finalizar y serializa correctamente las invocaciones anidadas en un equipo virtual de tamaño uno; `ForDynamic` reparte trabajo efectivamente entre múltiples workers; el pool ejecuta sus workers de manera concurrente; y todas las primitivas operan sin condiciones de carrera bajo el detector de carreras (`go test -race`).],
  ),
  caption: [Categorías de comportamiento cubiertas por la suite de pruebas],
)
