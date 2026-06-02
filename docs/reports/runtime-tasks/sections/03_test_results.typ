= Resultados de Ejecución

La verificación de la suite se realizó en dos pasos. Primero se ejecutó `go test -coverprofile=runtime_cov.out ./pkg/runtime/...`, que ejecuta las pruebas y genera un archivo de perfil con los datos crudos de cobertura. Luego, ese archivo se procesó con `go tool cover -func=runtime_cov.out` para obtener la tabla de cobertura por función presentada en la sección 3.4. Adicionalmente, la suite de tareas y dependencias fue ejecutada con `go test -race` para verificar la ausencia de condiciones de carrera bajo el detector de carreras nativo del lenguaje.

== Resumen cuantitativo

#figure(
  table(
    columns: (auto, auto),
    align: (left, right),
    [*Métrica*],                                            [*Valor*],
    [Total de pruebas ejecutadas (módulo completo)],        [93],
    [Pruebas del módulo de tareas y dependencias],          [35],
    [Pruebas del submódulo de tareas (`task_test.go`)],     [20],
    [Pruebas del submódulo de dependencias (`depend_test.go`)], [15],
    [Pruebas exitosas],                                     [93],
    [Pruebas fallidas],                                     [0],
    [Pruebas exitosas con detector de carreras activo],     [93],
    [Cobertura total de instrucciones],                     [100.0%],
    [Funciones del módulo con cobertura del 100%],          [12 de 12],
  ),
  caption: [Resumen cuantitativo de la ejecución de la suite],
)

== Distribución de pruebas por primitiva

#figure(
  table(
    columns: (auto, auto, 1fr),
    align: (left, right, left),
    table.header([*Primitiva*], [*Cantidad*], [*Propósito*]),
    [`Task`],            [4],   [Verifican que el cuerpo se ejecuta, que la llamada retorna sin esperar al cuerpo, que múltiples tareas concurrentes corren a completitud, y que `Task` funciona fuera de cualquier contexto de tarea.],
    [`Taskwait`],        [4],   [Verifican que `Taskwait` espera a los hijos directos con uno o varios hijos, que es un no-op fuera de todo contexto de tarea, y que no espera a los nietos (trampa de deadlock).],
    [`Taskgroup`],       [4],   [Verifican que `Taskgroup` espera a todo el árbol de descendientes incluyendo nietos, que las invocaciones anidadas funcionan correctamente (la interna espera a sus tareas antes de retornar), y que no produce deadlock con 100 tareas concurrentes.],
    [`Taskloop`],        [7],   [Verifican que todas las iteraciones se ejecutan exactamente una vez con sus índices correctos, que `iterations <= 0` produce cero ejecuciones, que `grainsize <= 0` se acota a 1, que `grainsize > iterations` produce una sola tarea y que `iterations = 1` produce exactamente una ejecución con índice 0.],
    [`Task` (estrés)],   [1],   [1000 tareas concurrentes, cada una escribiendo en su propia ranura atómica; verifica que ninguna se omite ni se duplica bajo máxima concurrencia.],
    [`TaskWithDepend`],  [15],  [Cubren el ordering `out→in`, la cadena `inout`, dos escritores consecutivos, tokens independientes, lector sin escritor previo, lectores concurrentes (verificación de valor y trampa de deadlock para probar no serialización), `inout` esperando a múltiples lectores, cadena larga de 10 tareas `inout`, dependencias múltiples `in` y `out`, pipeline de tres etapas `out→in→out`, `in` después de `inout`, ausencia de deadlock y estrés de 50 tareas `inout` con incremento no atómico.],
  ),
  caption: [Distribución de pruebas por primitiva del módulo],
)

== Cobertura detallada por función

#figure(
  table(
    columns: (auto, auto),
    align: (left, right),
    table.header([*Función*], [*Cobertura*]),
    [`newHandle`],           [100.0%],
    [`currentTask`],         [100.0%],
    [`registerTask`],        [100.0%],
    [`unregisterTask`],      [100.0%],
    [`waitSubtree`],         [100.0%],
    [`Task`],                [100.0%],
    [`Taskwait`],            [100.0%],
    [`Taskgroup`],           [100.0%],
    [`Taskloop`],            [100.0%],
    [`getOrCreateEntry`],    [100.0%],
    [`claimDeps`],           [100.0%],
    [`TaskWithDepend`],      [100.0%],
    [*Total del módulo*],    [*100.0%*],
  ),
  caption: [Cobertura agregada por función pública e interna del módulo],
)

== Salida directa de la herramienta de cobertura

A continuación se reproduce la salida del comando `go tool cover -func=runtime_cov.out` filtrada al módulo evaluado. Cada línea representa una función con su porcentaje de cobertura individual. Por razones de formato, las rutas se muestran relativas al directorio del módulo.

#figure(
  ```
pkg/runtime/depend.go:19:   getOrCreateEntry   100.0%
pkg/runtime/depend.go:31:   claimDeps          100.0%
pkg/runtime/depend.go:75:   TaskWithDepend     100.0%
pkg/runtime/task.go:21:     newHandle          100.0%
pkg/runtime/task.go:28:     currentTask        100.0%
pkg/runtime/task.go:36:     registerTask       100.0%
pkg/runtime/task.go:44:     unregisterTask     100.0%
pkg/runtime/task.go:53:     waitSubtree        100.0%
pkg/runtime/task.go:67:     Task               100.0%
pkg/runtime/task.go:88:     Taskwait           100.0%
pkg/runtime/task.go:106:    Taskgroup          100.0%
pkg/runtime/task.go:117:    Taskloop           100.0%
  ```,
  caption: [Salida del comando `go tool cover -func=runtime_cov.out` filtrada al módulo],
)

= Conclusión

El módulo de tareas y dependencias de datos alcanza una cobertura del 100% de instrucciones ejecutables con 35 pruebas distribuidas entre los dos submódulos. La suite cubre la totalidad de las primitivas públicas (`Task`, `Taskwait`, `Taskgroup`, `Taskloop` y `TaskWithDepend`), sus funciones internas de soporte, sus casos límite y sus garantías de concurrencia. Las pruebas más exigentes incluyen trampas de deadlock que detectarían regresiones semánticas no observables mediante cobertura de líneas: la trampa de `Taskwait_DoesNotWaitGrandchildren` detectaría si `Taskwait` comenzara a esperar a los nietos, y la trampa de `TaskWithDepend_ReadersDoNotBlockEachOther` detectaría si el runtime serializara inadvertidamente tareas `in` entre sí. La prueba de estrés de `TaskWithDepend_StressNoRace` usa incremento no atómico en una cadena de 50 tareas `inout`, verificando la correctitud del ordenamiento por dependencias incluso sin el detector de carreras. La suite completa pasa bajo `go test -race`.
