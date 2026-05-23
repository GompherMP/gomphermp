= Cobertura Funcional

Esta sección demuestra que cada primitiva pública del módulo cuenta con pruebas dedicadas que verifican su comportamiento bajo entradas válidas, casos límite y condiciones de concurrencia.

Por brevedad, los nombres de las pruebas en las tablas siguientes omiten el prefijo `Test`. Todas las pruebas residen en `pkg/runtime/parallel_test.go`.

== Primitivas y pruebas asociadas

#figure(
  table(
    columns: (auto, 1fr),
    align: (left + horizon, left + horizon),
    table.header([*Primitiva*], [*Prueba*]),

    table.cell(rowspan: 5)[`Parallel`],
                                            [`Parallel_AllThreadsExecute`],
                                            [`Parallel_CorrectThreadIDs`],
                                            [`Parallel_ImplicitBarrier`],
                                            [`Parallel_InvalidNumThreads`],
                                            [`Parallel_NegativeNumThreads`],

    table.cell(rowspan: 6)[`For`],
                                            [`For_BasicExecution`],
                                            [`For_CorrectIterationValues`],
                                            [`For_ZeroIterations`],
                                            [`For_NegativeIterations`],
                                            [`For_InvalidNumThreads`],
                                            [`For_FewerIterationsThanThreads`],

    table.cell(rowspan: 5)[`ParallelFor`],
                                            [`ParallelFor_DistributesWork`],
                                            [`ParallelFor_CorrectValues`],
                                            [`ParallelFor_ZeroIterations`],
                                            [`ParallelFor_NegativeIterations`],
                                            [`ParallelFor_InvalidNumThreads`],

    table.cell(rowspan: 10)[`ForDynamic`],
                                            [`ForDynamic_AllIterationsExecute`],
                                            [`ForDynamic_CorrectValues`],
                                            [`ForDynamic_ChunkSizeOne`],
                                            [`ForDynamic_ChunkSizeLargerThanIterations`],
                                            [`ForDynamic_ZeroIterations`],
                                            [`ForDynamic_NegativeIterations`],
                                            [`ForDynamic_InvalidChunkSize`],
                                            [`ForDynamic_InvalidNumThreads`],
                                            [`ForDynamic_DistributesAcrossGoroutines`],
                                            [`ForDynamic_StressNoRace`],

    table.cell(rowspan: 6)[`Sections`],
                                            [`Sections_AllSectionsExecute`],
                                            [`Sections_FewerSectionsThanThreads`],
                                            [`Sections_MoreSectionsThanThreads`],
                                            [`Sections_EmptyList`],
                                            [`Sections_InvalidNumThreads`],
                                            [`Sections_DifferentBodies`],
  ),
  caption: [Mapeo de primitivas a pruebas que verifican su funcionamiento],
)

== Comportamientos verificados por categoría

Las pruebas se agrupan en cuatro categorías transversales, aplicadas a cada primitiva según corresponda:

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
        [Cuando `NumThreads` o `chunkSize` reciben valores no positivos, las primitivas los corrigen al mínimo seguro de uno sin abortar la ejecución.],

    [Garantías de concurrencia],
        [`Parallel` garantiza la barrera implícita al finalizar; `ForDynamic` reparte trabajo efectivamente entre múltiples goroutines y opera sin condiciones de carrera bajo el detector de carreras (`go test -race`).],
  ),
  caption: [Categorías de comportamiento cubiertas por la suite de pruebas],
)
