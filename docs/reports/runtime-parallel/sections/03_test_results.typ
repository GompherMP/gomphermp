= Resultados de Ejecución

La verificación de la suite se realizó en dos pasos. Primero se ejecutó `go test -coverprofile=runtime_cov.out ./pkg/runtime/...`, que ejecuta las pruebas y genera un archivo de perfil con los datos crudos de cobertura. Luego, ese archivo se procesó con `go tool cover -func=runtime_cov.out` para obtener la tabla de cobertura por función presentada en la sección 3.4. Adicionalmente, la suite completa fue ejecutada con `go test -race` para verificar la ausencia de condiciones de carrera bajo el detector de carreras nativo del lenguaje.

== Resumen cuantitativo

#figure(
  table(
    columns: (auto, auto),
    align: (left, right),
    [*Métrica*],                                            [*Valor*],
    [Total de pruebas ejecutadas],                          [53],
    [Pruebas exitosas],                                     [53],
    [Pruebas fallidas],                                     [0],
    [Pruebas exitosas con detector de carreras activo],     [53],
    [Cobertura total de instrucciones],                     [100.0%],
    [Funciones del módulo con cobertura del 100%],          [21 de 21],
  ),
  caption: [Resumen cuantitativo de la ejecución de la suite],
)

== Distribución de pruebas por primitiva

#figure(
  table(
    columns: (auto, auto, 1fr),
    align: (left, right, left),
    table.header([*Primitiva*], [*Cantidad*], [*Propósito*]),
    [`Parallel`],                  [6],   [Verifican la ejecución correcta del equipo de workers, asignación de identificadores, barrera implícita, autocorrección de configuración y serialización del paralelismo anidado.],
    [`For` / `ParallelFor`],       [12],  [Verifican el reparto estático de iteraciones bajo entradas válidas y degeneradas, y la combinación de creación de equipo y reparto en una sola invocación.],
    [`ForDynamic` / `ParallelForDynamic`], [12], [Verifican el reparto dinámico con contador atómico compartido, distribución efectiva entre workers y robustez bajo estrés con detector de carreras.],
    [`ForStaticChunked` / `ParallelForStaticChunked`], [5], [Verifican el reparto block-cyclic de `schedule(static, chunk)`: la distribución round-robin exacta de los trozos, la cobertura completa del espacio, el caso secuencial sin equipo y el acotamiento de chunk no positivo.],
    [`Sections` / `ParallelSections`], [8], [Verifican la distribución dinámica de bloques independientes, incluyendo los casos de menos secciones que workers y viceversa, y el caso compuesto que provisiona el equipo.],
    [Pool],                        [4],   [Verifican la pre-instanciación del pool durante `init()`, la persistencia de los workers entre regiones paralelas sucesivas y la ejecución concurrente real de los workers.],
    [`SetPoolSize`],               [4],   [Verifican el redimensionamiento del pool, la operatividad del nuevo pool, el acotamiento de valores no positivos y la terminación limpia de los workers reemplazados.],
    [`CurrentTeamSize`],           [2],   [Verifican el valor retornado fuera de toda región paralela (`1`) y desde dentro de un job en ejecución con un team context registrado.],
  ),
  caption: [Distribución de pruebas por primitiva del módulo],
)

== Cobertura detallada por función

#figure(
  table(
    columns: (auto, auto),
    align: (left, right),
    table.header([*Función*], [*Cobertura*]),
    [`Parallel`],             [100.0%],
    [`For`],                  [100.0%],
    [`ParallelFor`],          [100.0%],
    [`ForDynamic`],           [100.0%],
    [`ParallelForDynamic`],   [100.0%],
    [`ForStaticChunked`],     [100.0%],
    [`ParallelForStaticChunked`], [100.0%],
    [`Sections`],             [100.0%],
    [`ParallelSections`],     [100.0%],
    [`registerInTeam`],       [100.0%],
    [`unregisterFromTeam`],   [100.0%],
    [`getCurrentTeam`],       [100.0%],
    [`newTeam`],              [100.0%],
    [`init` (pool)],          [100.0%],
    [`newPool`],              [100.0%],
    [`poolWorker`],           [100.0%],
    [`submit`],               [100.0%],
    [`getPool`],              [100.0%],
    [`PoolSize`],             [100.0%],
    [`CurrentTeamSize`],      [100.0%],
    [`SetPoolSize`],          [100.0%],
    [*Total del módulo*],     [*100.0%*],
  ),
  caption: [Cobertura agregada por función pública e interna del módulo],
)

== Salida directa de la herramienta de cobertura

A continuación se reproduce la salida del comando `go tool cover -func=runtime_cov.out` filtrada al módulo evaluado. Cada línea representa una función con su porcentaje de cobertura individual. Por razones de formato, las rutas se muestran relativas al directorio del módulo.

#figure(
  ```
pkg/runtime/parallel.go:37:    registerInTeam            100.0%
pkg/runtime/parallel.go:47:    unregisterFromTeam        100.0%
pkg/runtime/parallel.go:56:    getCurrentTeam            100.0%
pkg/runtime/parallel.go:66:    newTeam                   100.0%
pkg/runtime/parallel.go:78:    Parallel                  100.0%
pkg/runtime/parallel.go:115:   For                       100.0%
pkg/runtime/parallel.go:143:   ParallelFor               100.0%
pkg/runtime/parallel.go:158:   ForDynamic                100.0%
pkg/runtime/parallel.go:193:   ParallelForDynamic        100.0%
pkg/runtime/parallel.go:209:   ForStaticChunked          100.0%
pkg/runtime/parallel.go:241:   ParallelForStaticChunked  100.0%
pkg/runtime/parallel.go:255:   Sections                  100.0%
pkg/runtime/parallel.go:284:   ParallelSections          100.0%
pkg/runtime/pool.go:40:        init                      100.0%
pkg/runtime/pool.go:47:        newPool                   100.0%
pkg/runtime/pool.go:64:        poolWorker                100.0%
pkg/runtime/pool.go:79:        submit                    100.0%
pkg/runtime/pool.go:84:        getPool                   100.0%
pkg/runtime/pool.go:92:        PoolSize                  100.0%
pkg/runtime/pool.go:98:        CurrentTeamSize           100.0%
pkg/runtime/pool.go:109:       SetPoolSize               100.0%
  ```,
  caption: [Salida del comando `go tool cover -func=runtime_cov.out` filtrada al módulo],
)

= Conclusión

El módulo de gestión de goroutines y reparto de trabajo alcanza una cobertura del 100% de instrucciones ejecutables, con 53 pruebas que cubren las nueve primitivas públicas de reparto de trabajo, las tres funciones de configuración del pool, sus casos límite y sus garantías de concurrencia. La suite completa pasa adicionalmente bajo el detector de carreras de Go, verificando la ausencia de condiciones de carrera incluso bajo escenarios de estrés con miles de iteraciones, contención máxima sobre los contadores atómicos compartidos y reemplazos sucesivos del pool en caliente.
