= Descripción del Módulo

El módulo de mecanismos de sincronización constituye el segundo subsistema de la librería de runtime de GompherMP. Provee las primitivas que el código transformado utiliza para implementar las directivas de sincronización especificadas en R1, garantizando la consistencia de los accesos a memoria compartida y la correcta orquestación de bloques de ejecución exclusivos.

== Ubicación

El módulo reside en el paquete público `pkg/runtime/` del repositorio. El siguiente cuadro detalla los archivos que componen el módulo:

#figure(
  table(
    columns: (auto, 1fr),
    align: (left, left),
    [*Archivo*],            [*Responsabilidad*],
    [`sync.go`],            [Implementa las primitivas públicas (`Critical`, `Single`, `Master`, `Barrier`) y la gestión interna de mutexes nombrados.],
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
    [`Single`],      [Provee la primitiva de ejecución única, sobre la cual el motor de transformación construye la coordinación con `sync.Once` y una barrera implícita para garantizar que el cuerpo se ejecute exactamente una vez dentro del equipo.],
    [`Master`],      [Ejecuta condicionalmente un bloque únicamente cuando es invocado desde la goroutine maestra del equipo (`threadID == 0`), sin imponer una barrera implícita posterior.],
    [`Barrier`],     [Establece un punto de sincronización explícito mediante un grupo de espera dimensionado al tamaño del equipo, garantizando que ninguna goroutine continúe hasta que todas hayan alcanzado el punto. Se comporta como no-op cuando es invocado fuera de una región paralela.],
  ),
  caption: [Primitivas públicas del módulo de mecanismos de sincronización],
)

== Modelo de sincronización

El diseño del módulo se apoya en las primitivas de sincronización de la librería estándar de Go: `sync.Mutex` para `Critical`, `sync.WaitGroup` para `Barrier`, y `sync.RWMutex` para proteger el registro interno de mutexes nombrados. Esta decisión permite delegar al runtime nativo del lenguaje las garantías de orden de memoria (happens-before) requeridas por el modelo de concurrencia, evitando reimplementar mecanismos de bajo nivel ya provistos por el ecosistema.

La modalidad nominal de `Critical` mantiene un mapa global de mutexes indexado por nombre, lo que permite que regiones críticas con identificadores distintos operen sin contención mutua. Por ejemplo, dos bloques marcados con `//gompher critical(lockA)` y `//gompher critical(lockB)` pueden ejecutarse simultáneamente, mientras que dos bloques con el mismo nombre se serializan correctamente.

La primitiva `Barrier` consulta el team context registrado para la goroutine que la invoca y opera sobre el grupo de espera asociado al equipo. Si la goroutine no pertenece a ningún equipo activo (es decir, fue invocada fuera de una región `Parallel`), la primitiva se comporta como no-op, evitando bloqueos o errores en código que mezcla regiones paralelas con código secuencial.

== Metodología de pruebas

La suite de pruebas del módulo se organiza por primitiva pública, cubriendo para cada una el comportamiento correcto bajo condiciones de concurrencia intensiva, la independencia entre instancias de la misma primitiva (por ejemplo, locks nombrados distintos) y los casos límite específicos del modelo (invocación fuera de una región paralela, equipos de distintos tamaños). Las pruebas usan contadores compartidos sometidos a actualizaciones concurrentes para verificar empíricamente la ausencia de condiciones de carrera, y aplican límites de tiempo para detectar posibles bloqueos por errores de implementación.
