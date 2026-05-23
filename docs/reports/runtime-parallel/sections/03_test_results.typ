= Resultados de Ejecución

La verificación de la suite se realizó en dos pasos. Primero se ejecutó `go test -coverprofile=runtime_cov.out ./pkg/runtime/...`, que ejecuta las pruebas y genera un archivo de perfil con los datos crudos de cobertura. Luego, ese archivo se procesó con `go tool cover -func=runtime_cov.out` para obtener la tabla de cobertura por función presentada en la sección 3.4. Adicionalmente, la suite completa fue ejecutada con `go test -race` para verificar la ausencia de condiciones de carrera bajo el detector de carreras nativo del lenguaje.

== Resumen cuantitativo

#figure(
  table(
    columns: (auto, auto),
    align: (left, right),
    [*Métrica*],                                            [*Valor*],
    [Total de pruebas ejecutadas],                          [32],
    [Pruebas exitosas],                                     [32],
    [Pruebas fallidas],                                     [0],
    [Pruebas exitosas con detector de carreras activo],     [32],
    [Cobertura total de instrucciones],                     [100.0%],
    [Funciones del módulo con cobertura del 100%],          [9 de 9],
  ),
  caption: [Resumen cuantitativo de la ejecución de la suite],
)

== Distribución de pruebas por primitiva

#figure(
  table(
    columns: (auto, auto, 1fr),
    align: (left, right, left),
    table.header([*Primitiva*], [*Cantidad*], [*Propósito*]),
    [`Parallel`],        [5],   [Verifican la creación correcta del equipo de goroutines, asignación de identificadores, barrera implícita y autocorrección de configuración.],
    [`For`],             [6],   [Verifican el reparto estático de iteraciones bajo entradas válidas y degeneradas.],
    [`ParallelFor`],     [5],   [Verifican la combinación de creación de equipo y reparto estático en una sola invocación.],
    [`ForDynamic`],      [10],  [Verifican el reparto dinámico con contador atómico compartido, distribución efectiva entre goroutines y robustez bajo estrés con detector de carreras.],
    [`Sections`],        [6],   [Verifican la distribución dinámica de bloques independientes, incluyendo los casos de menos secciones que hilos y viceversa.],
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
    [`Sections`],             [100.0%],
    [`getGoroutineID`],       [100.0%],
    [`registerInTeam`],       [100.0%],
    [`unregisterFromTeam`],   [100.0%],
    [`getCurrentTeam`],       [100.0%],
    [*Total del módulo*],     [*100.0%*],
  ),
  caption: [Cobertura agregada por función pública e interna del módulo],
)

== Salida directa de la herramienta de cobertura

A continuación se reproduce la salida del comando `go tool cover -func=runtime_cov.out` filtrada al módulo evaluado. Cada línea representa una función con su porcentaje de cobertura individual. Por razones de formato, las rutas se muestran relativas al directorio del módulo.

#figure(
  ```
pkg/runtime/parallel.go:28:    getGoroutineID         100.0%
pkg/runtime/parallel.go:37:    registerInTeam         100.0%
pkg/runtime/parallel.go:45:    unregisterFromTeam     100.0%
pkg/runtime/parallel.go:53:    getCurrentTeam         100.0%
pkg/runtime/parallel.go:63:    For                    100.0%
pkg/runtime/parallel.go:103:   Parallel               100.0%
pkg/runtime/parallel.go:134:   ParallelFor            100.0%
pkg/runtime/parallel.go:173:   ForDynamic             100.0%
pkg/runtime/parallel.go:213:   Sections               100.0%
total:                         (statements)           100.0%
  ```,
  caption: [Salida del comando `go tool cover -func=runtime_cov.out` filtrada al módulo],
)

= Conclusión

El módulo de gestión de goroutines y reparto de trabajo alcanza una cobertura del 100% de instrucciones ejecutables, con 32 pruebas que cubren las cinco primitivas públicas, sus casos límite y sus garantías de concurrencia. La suite completa pasa adicionalmente bajo el detector de carreras de Go, verificando la ausencia de condiciones de carrera incluso bajo escenarios de estrés con miles de iteraciones y contención máxima sobre los contadores atómicos compartidos.
