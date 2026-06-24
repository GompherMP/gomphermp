= Resultados de Ejecución

La verificación de la suite se realizó en dos pasos. Primero se ejecutó `go test -coverprofile=runtime_cov.out ./pkg/runtime/...`, que ejecuta las pruebas y genera un archivo de perfil con los datos crudos de cobertura. Luego, ese archivo se procesó con `go tool cover -func=runtime_cov.out` para obtener la tabla de cobertura por función presentada en la sección 3.4. Adicionalmente, la suite completa fue ejecutada con `go test -race` para verificar la ausencia de condiciones de carrera bajo el detector de carreras nativo del lenguaje.

== Resumen cuantitativo

#figure(
  table(
    columns: (auto, auto),
    align: (left, right),
    [*Métrica*],                                            [*Valor*],
    [Total de pruebas ejecutadas],                          [18],
    [Pruebas exitosas],                                     [18],
    [Pruebas fallidas],                                     [0],
    [Pruebas exitosas con detector de carreras activo],     [18],
    [Cobertura total de instrucciones],                     [100.0%],
    [Funciones del módulo con cobertura del 100%],          [8 de 8],
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
    [`Single`],     [4],   [Verifican que el cuerpo se ejecuta exactamente una vez por la elección por CAS, que el token se reinicia para permitir varios `single` en la misma región, y la corrección bajo concurrencia.],
    [`Master`],     [4],   [Verifican la exclusividad del bloque en la goroutine maestra, la ausencia de barrera implícita, el control correcto del `threadID` y la continuación de todas las goroutines tras la invocación.],
    [`Barrier`],    [6],   [Verifican la sincronización efectiva del equipo, el comportamiento no-op fuera de una región paralela, el funcionamiento con distintos tamaños de equipo, la reutilización de la barrera cíclica en rondas sucesivas, la ausencia de deadlock y la visibilidad de las operaciones previas a la barrera.],
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
    [`newCyclicBarrier`],     [100.0%],
    [`cyclicBarrier.wait`],   [100.0%],
    [`cyclicBarrier.waitThen`], [100.0%],
    [*Total del módulo*],     [*100.0%*],
  ),
  caption: [Cobertura agregada por función pública e interna del módulo],
)

== Salida directa de la herramienta de cobertura

A continuación se reproduce la salida del comando `go tool cover -func=runtime_cov.out` filtrada al módulo evaluado. Cada línea representa una función con su porcentaje de cobertura individual. Por razones de formato, las rutas se muestran relativas al directorio del módulo.

#figure(
  ```
pkg/runtime/sync.go:18:    Critical               100.0%
pkg/runtime/sync.go:34:    getNamedLock           100.0%
pkg/runtime/sync.go:56:    Single                 100.0%
pkg/runtime/sync.go:76:    Master                 100.0%
pkg/runtime/sync.go:89:    Barrier                100.0%
pkg/runtime/sync.go:111:   newCyclicBarrier       100.0%
pkg/runtime/sync.go:118:   wait                   100.0%
pkg/runtime/sync.go:132:   waitThen               100.0%
  ```,
  caption: [Salida del comando `go tool cover -func=runtime_cov.out` filtrada al módulo],
)

= Conclusión

El módulo de mecanismos de sincronización alcanza una cobertura del 100% de instrucciones ejecutables, con 18 pruebas que cubren las cuatro primitivas públicas y la barrera cíclica interna, su semántica específica (elección por CAS en `Single`, exclusividad maestra, reutilización de la barrera, no-op fuera de región paralela) y sus garantías de concurrencia bajo carga. La suite completa pasa adicionalmente bajo el detector de carreras de Go, verificando la ausencia de condiciones de carrera incluso bajo escenarios con cientos de goroutines compitiendo simultáneamente por los mismos recursos protegidos.
