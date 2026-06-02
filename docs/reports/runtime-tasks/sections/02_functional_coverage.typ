= Cobertura Funcional

Esta sección demuestra que cada primitiva pública del módulo cuenta con pruebas dedicadas que verifican su comportamiento bajo entradas válidas, casos límite y condiciones de concurrencia.

Por brevedad, los nombres de las pruebas en las tablas siguientes omiten el prefijo `Test`. Las pruebas residen en `pkg/runtime/task_test.go` y `pkg/runtime/depend_test.go`.

== Primitivas y pruebas asociadas

#figure(
  table(
    columns: (auto, 1fr),
    align: (left + horizon, left + horizon),
    table.header([*Primitiva*], [*Prueba*]),

    table.cell(rowspan: 4)[`Task`],
                                            [`Task_BodyExecutes`],
                                            [`Task_IsAsync`],
                                            [`Task_MultipleTasksExecute`],
                                            [`Task_OutsideTaskContext`],

    table.cell(rowspan: 4)[`Taskwait`],
                                            [`Taskwait_WaitsForDirectChildren`],
                                            [`Taskwait_MultipleChildren`],
                                            [`Taskwait_NoopOutsideTask`],
                                            [`Taskwait_DoesNotWaitGrandchildren`],

    table.cell(rowspan: 4)[`Taskgroup`],
                                            [`Taskgroup_WaitsFullSubtree`],
                                            [`Taskgroup_GrandchildrenIncluded`],
                                            [`Taskgroup_Nested`],
                                            [`Taskgroup_NoDeadlock`],

    table.cell(rowspan: 7)[`Taskloop`],
                                            [`Taskloop_AllIterationsExecute`],
                                            [`Taskloop_CorrectIterationValues`],
                                            [`Taskloop_NegativeIterations`],
                                            [`Taskloop_ZeroIterations`],
                                            [`Taskloop_InvalidGrainsize`],
                                            [`Taskloop_GrainsizeLargerThanIterations`],
                                            [`Taskloop_SingleIteration`],

    table.cell(rowspan: 1)[`Task` (estrés)],
                                            [`Task_StressNoRace`],

    table.cell(rowspan: 15)[`TaskWithDepend`],
                                            [`TaskWithDepend_OutBeforeIn`],
                                            [`TaskWithDepend_InoutChain`],
                                            [`TaskWithDepend_ConsecutiveWriters`],
                                            [`TaskWithDepend_IndependentTokens`],
                                            [`TaskWithDepend_InWithNoWriter`],
                                            [`TaskWithDepend_MultipleReaders`],
                                            [`TaskWithDepend_ReadersDoNotBlockEachOther`],
                                            [`TaskWithDepend_InoutAfterMultipleReaders`],
                                            [`TaskWithDepend_LongInoutChain`],
                                            [`TaskWithDepend_MultipleInTokens`],
                                            [`TaskWithDepend_MultipleOutTokens`],
                                            [`TaskWithDepend_OutThenInThenOut`],
                                            [`TaskWithDepend_InAfterInout`],
                                            [`TaskWithDepend_NoDeadlock`],
                                            [`TaskWithDepend_StressNoRace`],
  ),
  caption: [Mapeo de primitivas a pruebas que verifican su funcionamiento],
)

== Comportamientos verificados por categoría

Las pruebas se agrupan en seis categorías transversales, aplicadas a cada primitiva según corresponda:

#figure(
  table(
    columns: (auto, 1fr),
    align: (left + horizon, left + horizon),
    table.header([*Categoría*], [*Comportamiento verificado*]),

    [Ejecución correcta],
        [Cada tarea, iteración o bloque se ejecuta con los valores correctos y exactamente el número de veces esperado.],

    [Asincronía],
        [`Task` retorna al llamador antes de que su cuerpo haya terminado, sin bloquearlo. Funciona tanto dentro como fuera de cualquier contexto de tarea.],

    [Semántica de espera],
        [`Taskwait` espera únicamente a los hijos directos de la tarea actual, no a sus descendientes más profundos, y es un no-op fuera de todo contexto de tarea. `Taskgroup` espera a todos los descendientes a cualquier nivel de profundidad, incluyendo los de tareas anidadas. Ambas propiedades se verifican incluyendo una prueba de trampa de deadlock: si `Taskwait` esperara a los nietos, bloquearía indefinidamente porque el canal de desbloqueo del nieto sólo se cierra después de que `Taskwait` retorna.],

    [Manejo de casos límite],
        [`Taskloop` retorna de forma segura sin crear ninguna tarea cuando recibe `iterations <= 0`, acota `grainsize <= 0` al mínimo seguro de 1, y maneja correctamente `grainsize > iterations` (una única tarea) e `iterations = 1` (exactamente un cuerpo ejecutado con índice 0).],

    [Ordenamiento por dependencias],
        [Las tareas `out` completan su escritura antes de que las tareas `in` sobre el mismo token lean el valor producido. Las cadenas `inout` se serializan en orden de despacho, incluyendo cadenas largas de 10 tareas. Dos escritores sucesivos sobre el mismo token se ejecutan sin superposición. Una tarea `in` sin escritor previo avanza sin esperar. Las tareas `in` sobre el mismo token no se serializan entre sí (verificado mediante trampa de deadlock). Una tarea `inout` espera tanto al escritor previo como a todos los lectores activos. Una tarea `in` declarada después de una `inout` espera a ésta. Un pipeline de tres etapas (`out → in → out`) preserva el orden observable. Tareas sobre tokens distintos son independientes. Tareas con múltiples tokens `in` esperan a todos sus escritores; tareas con múltiples tokens `out` bloquean a todos sus lectores correspondientes.],

    [Ausencia de carreras de datos],
        [La suite completa pasa bajo el detector de carreras (`go test -race`). La prueba de estrés de `TaskWithDepend` usa incremento no atómico dentro de una cadena de 50 tareas `inout`, por lo que cualquier fallo en la serialización produciría un valor final incorrecto detectable sin el detector de carreras.],
  ),
  caption: [Categorías de comportamiento cubiertas por la suite de pruebas],
)
