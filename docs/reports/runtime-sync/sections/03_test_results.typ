= Resultados de Ejecución

La verificación de la suite se realizó en dos pasos. Primero se ejecutó `go test -coverprofile=runtime_cov.out ./pkg/runtime/...`, que ejecuta las pruebas y genera un archivo de perfil con los datos crudos de cobertura. Luego, ese archivo se procesó con `go tool cover -func=runtime_cov.out` para obtener la tabla de cobertura por función presentada en la sección 3.4. Adicionalmente, la suite completa fue ejecutada con `go test -race` para verificar la ausencia de condiciones de carrera bajo el detector de carreras nativo del lenguaje.

== Resumen cuantitativo

#figure(
  table(
    columns: (auto, auto),
    align: (left, right),
    [*Métrica*],                                            [*Valor*],
    [Total de pruebas ejecutadas],                          [15],
    [Pruebas exitosas],                                     [15],
    [Pruebas fallidas],                                     [0],
    [Pruebas exitosas con detector de carreras activo],     [15],
    [Cobertura total de instrucciones],                     [100.0%],
    [Funciones del módulo con cobertura del 100%],          [5 de 5],
  ),
  caption: [Resumen cuantitativo de la ejecución de la suite],
)

== Distribución de pruebas por primitiva

#figure(
  table(
    columns: (auto, auto, 1fr),
    align: (left, right, left),
    table.header([*Primitiva*], [*Cantidad*], [*Propósito*]),
    [`Critical`],   [4],   [Verifican exclusión mutua bajo carga concurrente, independencia entre locks nombrados, serialización dentro de un mismo lock, e independencia entre lock anónimo y locks nombrados.],
    [`Single`],     [2],   [Verifican que la primitiva ejecuta el cuerpo y que puede ser invocada múltiples veces.],
    [`Master`],     [4],   [Verifican la exclusividad del bloque en la goroutine maestra, la ausencia de barrera implícita, el control correcto del `threadID` y la continuación de todas las goroutines tras la invocación.],
    [`Barrier`],    [5],   [Verifican la sincronización efectiva del equipo, el comportamiento no-op fuera de una región paralela, el funcionamiento con distintos tamaños de equipo, la ausencia de deadlock y la visibilidad de las operaciones previas a la barrera.],
  ),
  caption: [Distribución de pruebas por primitiva del módulo],
)

== Cobertura detallada por función

#figure(
  table(
    columns: (auto, auto),
    align: (left, right),
    table.header([*Función*], [*Cobertura*]),
    [`Critical`],             [100.0%],
    [`Single`],               [100.0%],
    [`Master`],               [100.0%],
    [`Barrier`],              [100.0%],
    [`getNamedLock`],         [100.0%],
    [*Total del módulo*],     [*100.0%*],
  ),
  caption: [Cobertura agregada por función pública e interna del módulo],
)

== Salida directa de la herramienta de cobertura

A continuación se reproduce la salida del comando `go tool cover -func=runtime_cov.out` filtrada al módulo evaluado. Cada línea representa una función con su porcentaje de cobertura individual. Por razones de formato, las rutas se muestran relativas al directorio del módulo.

#figure(
  ```
pkg/runtime/sync.go:17:    Critical               100.0%
pkg/runtime/sync.go:33:    getNamedLock           100.0%
pkg/runtime/sync.go:51:    Single                 100.0%
pkg/runtime/sync.go:57:    Master                 100.0%
pkg/runtime/sync.go:65:    Barrier                100.0%
total:                     (statements)           100.0%
  ```,
  caption: [Salida del comando `go tool cover -func=runtime_cov.out` filtrada al módulo],
)

= Conclusión

El módulo de mecanismos de sincronización alcanza una cobertura del 100% de instrucciones ejecutables, con 15 pruebas que cubren las cuatro primitivas públicas, su semántica específica (exclusividad maestra, ausencia de barrera, no-op fuera de región paralela) y sus garantías de concurrencia bajo carga. La suite completa pasa adicionalmente bajo el detector de carreras de Go, verificando la ausencia de condiciones de carrera incluso bajo escenarios con cientos de goroutines compitiendo simultáneamente por los mismos recursos protegidos.
