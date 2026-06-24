= Descripción del Módulo

El módulo de mecanismos de sincronización constituye el segundo subsistema de la librería de runtime de GompherMP. Provee las primitivas que el código transformado utiliza para implementar las directivas de sincronización especificadas en R1, garantizando la consistencia de los accesos a memoria compartida y la correcta orquestación de bloques de ejecución exclusivos.

== Ubicación

El módulo reside en el paquete público `pkg/runtime/` del repositorio. El siguiente cuadro detalla los archivos que componen el módulo:

#figure(
  table(
    columns: (auto, 1fr),
    align: (left, left),
    [*Archivo*],            [*Responsabilidad*],
    [`sync.go`],            [Implementa las primitivas públicas (`Critical`, `Single`, `Master`, `Barrier`), la barrera cíclica reutilizable (`cyclicBarrier`) y la gestión interna de mutexes nombrados.],
    [`sync_test.go`],       [Suite completa de pruebas unitarias del módulo.],
  ),
  caption: [Archivos que componen el módulo de mecanismos de sincronización],
)

== Primitivas públicas

El módulo expone cuatro funciones que el motor de transformación invoca para traducir las directivas de sincronización del programa original:

#figure(
  table(
    columns: (auto, 1fr),
    align: (left, left),
    [*Primitiva*], [*Propósito*],
    [`Critical`],    [Garantiza exclusión mutua sobre un bloque de código. Soporta modalidad anónima (mutex global) y modalidad nominal (mutex por nombre), permitiendo que regiones críticas con nombres distintos se ejecuten en paralelo.],
    [`Single`],      [Garantiza que el cuerpo se ejecute exactamente una vez dentro del equipo mediante una elección por _compare-and-swap_ (CAS, operación atómica que actualiza un valor solo si todavía coincide con el esperado) sobre el token `singleFlag` del team context: la primera goroutine que gana el CAS ejecuta el bloque, las demás lo omiten. El token se reinicia en la barrera implícita de cierre, de modo que una misma región puede contener varios `single`.],
    [`Master`],      [Ejecuta condicionalmente un bloque únicamente cuando es invocado desde la goroutine maestra del equipo (`threadID == 0`), sin imponer una barrera implícita posterior.],
    [`Barrier`],     [Establece un punto de sincronización mediante la barrera cíclica reutilizable del equipo, garantizando que ninguna goroutine continúe hasta que todas hayan alcanzado el punto. Se comporta como no-op cuando es invocado fuera de una región paralela.],
  ),
  caption: [Primitivas públicas del módulo de mecanismos de sincronización],
)

== Modelo de sincronización

El diseño del módulo se apoya en las primitivas de sincronización de la librería estándar de Go: `sync.Mutex` para `Critical` y `sync.RWMutex` para proteger el registro interno de mutexes nombrados, delegando al runtime nativo del lenguaje las garantías de orden de memoria requeridas. `Barrier`, en cambio, se respalda en una *barrera cíclica* propia (`cyclicBarrier`, construida sobre `sync.Cond` y un contador de generación). A diferencia de un `sync.WaitGroup` (de un solo uso), la barrera cíclica es reutilizable: como el equipo es persistente y atraviesa varias construcciones de reparto dentro de la misma región (`for`, `single`, `sections`, …), cada una con su propia barrera implícita, se requiere un mecanismo que pueda dispararse repetidamente. El contador de generación distingue las rondas sucesivas y evita los despertares espurios.

La modalidad nominal de `Critical` mantiene un mapa global de mutexes indexado por nombre, lo que permite que regiones críticas con identificadores distintos operen sin contención mutua. Por ejemplo, dos bloques marcados con `//gompher critical(lockA)` y `//gompher critical(lockB)` pueden ejecutarse simultáneamente, mientras que dos bloques con el mismo nombre se serializan correctamente.

La primitiva `Barrier` consulta el team context registrado para la goroutine que la invoca y opera sobre la barrera cíclica asociada al equipo. Si la goroutine no pertenece a ningún equipo activo (es decir, fue invocada fuera de una región `Parallel`), la primitiva se comporta como no-op, evitando bloqueos o errores en código que mezcla regiones paralelas con código secuencial.

== Metodología de pruebas

La suite de pruebas del módulo se organiza por primitiva pública, cubriendo para cada una el comportamiento correcto bajo condiciones de concurrencia intensiva, la independencia entre instancias de la misma primitiva (por ejemplo, locks nombrados distintos) y los casos límite específicos del modelo (invocación fuera de una región paralela, equipos de distintos tamaños). Las pruebas usan contadores compartidos sometidos a actualizaciones concurrentes para verificar empíricamente la ausencia de condiciones de carrera, y aplican límites de tiempo para detectar posibles bloqueos por errores de implementación.
