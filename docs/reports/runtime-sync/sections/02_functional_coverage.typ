= Cobertura Funcional

Esta sección demuestra que cada primitiva pública del módulo cuenta con pruebas dedicadas que verifican su comportamiento bajo condiciones de concurrencia y casos límite.

Por brevedad, los nombres de las pruebas en las tablas siguientes omiten el prefijo `Test`. Todas las pruebas residen en `pkg/runtime/sync_test.go`.

== Primitivas y pruebas asociadas

#figure(
  table(
    columns: (auto, 1fr),
    align: (left + horizon, left + horizon),
    table.header([*Primitiva*], [*Prueba*]),

    table.cell(rowspan: 4)[`Critical`],
                                            [`Critical_PreventRaceCondition`],
                                            [`Critical_NamedLocks`],
                                            [`Critical_SameNamedLockSerializes`],
                                            [`Critical_AnonymousVsNamed`],

    table.cell(rowspan: 2)[`Single`],
                                            [`Single_Executes`],
                                            [`Single_ExecutesMultipleTimes`],

    table.cell(rowspan: 4)[`Master`],
                                            [`Master_OnlyMasterExecutes`],
                                            [`Master_NoImplicitBarrier`],
                                            [`Master_CorrectThreadIDCheck`],
                                            [`Master_AllThreadsContinue`],

    table.cell(rowspan: 5)[`Barrier`],
                                            [`Barrier_SynchronizesGoroutines`],
                                            [`Barrier_OutsideParallel`],
                                            [`Barrier_DifferentTeamSizes`],
                                            [`Barrier_NoDeadlock`],
                                            [`Barrier_OrderOfOperations`],
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

    [Protección contra condiciones de carrera],
        [`Critical` previene actualizaciones concurrentes incorrectas sobre contadores compartidos bajo cargas de cientos de goroutines.],

    [Independencia entre instancias],
        [Locks nombrados distintos (`lockA` y `lockB`) operan en paralelo sin contención mutua; el lock anónimo y los locks nombrados son independientes entre sí.],

    [Semántica específica del modelo],
        [`Master` ejecuta el bloque únicamente en `threadID == 0` y no impone barrera implícita; `Barrier` espera correctamente a todos los miembros del equipo y se comporta como no-op fuera de una región paralela.],

    [Robustez bajo carga],
        [`Barrier` no incurre en deadlock bajo distintos tamaños de equipo (1, 2, 4 y 8 goroutines); las primitivas completan su ejecución dentro de límites de tiempo razonables.],
  ),
  caption: [Categorías de comportamiento cubiertas por la suite de pruebas],
)
