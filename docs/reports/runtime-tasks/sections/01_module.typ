= Descripción del Módulo

El módulo de tareas y dependencias de datos constituye el tercer subsistema de la librería de runtime de GompherMP. Recibe como entrada las invocaciones generadas por el motor de transformación del AST y se encarga de despachar cuerpos de función como goroutines independientes, mantener las relaciones de parentesco entre tareas, sincronizar la ejecución al finalizar subtrees de tareas y garantizar el ordenamiento correcto entre tareas que comparten datos mediante el esquema de dependencias `in`/`out`/`inout`. Provee las primitivas de ejecución que soportan las directivas `task`, `taskwait`, `taskgroup`, `taskloop` y la cláusula `depend` formalizadas en la especificación de GompherMP.

== Ubicación

El módulo reside en el paquete público `pkg/runtime/` del repositorio. El siguiente cuadro detalla los archivos que componen el módulo:

#figure(
  table(
    columns: (auto, 1fr),
    align: (left, left),
    [*Archivo*],            [*Responsabilidad*],
    [`task.go`],            [Implementa la gestión del árbol de tareas: ciclo de vida de cada `taskHandle`, registro de goroutines activas, y las funciones públicas `Task`, `Taskwait`, `Taskgroup` y `Taskloop`.],
    [`task_test.go`],       [Suite completa de pruebas unitarias de las primitivas de gestión de tareas.],
    [`depend.go`],          [Implementa el registro de dependencias de datos: la estructura `depEntry`, la función interna `claimDeps` y la función pública `TaskWithDepend`.],
    [`depend_test.go`],     [Pruebas unitarias de los distintos patrones de dependencia (`in`, `out`, `inout`, tokens independientes y ausencia de escritor previo).],
  ),
  caption: [Archivos que componen el módulo de tareas y dependencias de datos],
)

== Primitivas públicas

El módulo expone cinco funciones que el motor de transformación invoca para traducir las directivas del programa original:

#figure(
  table(
    columns: (auto, 1fr),
    align: (left, left),
    [*Primitiva*], [*Propósito*],
    [`Task`],            [Despacha `body` como una goroutine asíncrona. Si la goroutine invocante pertenece a una tarea activa, la nueva tarea queda registrada como hija de ésta en el árbol de tareas. Retorna inmediatamente.],
    [`Taskwait`],        [Bloquea a la goroutine invocante hasta que todos los hijos directos de la tarea actual hayan terminado. No espera a los descendientes más profundos. Es un no-op si se invoca fuera de cualquier contexto de tarea.],
    [`Taskgroup`],       [Ejecuta `body` de forma síncrona en el hilo invocante - no crea una tarea hija ni se registra en el árbol de tareas del contexto externo. Corresponde a la semántica del estándar OpenMP, donde `taskgroup` es un constructo, no una tarea. Al finalizar `body`, bloquea hasta que todos los descendientes generados dentro del grupo, a cualquier nivel de profundidad, hayan terminado. Proporciona la barrera profunda que `Taskwait` no garantiza.],
    [`Taskloop`],        [Distribuye el rango `[0, iterations)` como tareas independientes, una por cada bloque de `grainsize` iteraciones. `grainsize` se acota a 1 si recibe un valor no positivo. Es un no-op para `iterations <= 0`.],
    [`TaskWithDepend`],  [Despacha `body` como una tarea con ordenamiento por dependencias de datos. `ins`, `outs` e `inouts` son direcciones de variables empleadas como tokens de correlación; nunca se desreferencian. La tarea se registra inmediatamente en el árbol de tareas y la espera efectiva de sus predecesores ocurre dentro de la goroutine antes de ejecutar `body`.],
  ),
  caption: [Primitivas públicas del módulo de tareas y dependencias de datos],
)

== Modelo de concurrencia

=== Árbol de tareas (`task.go`)

Cada tarea está representada por un `taskHandle`, una estructura que contiene un canal `done` (cerrado al finalizar el cuerpo de la tarea), un mutex para proteger la lista de hijos, y dicha lista de punteros a `taskHandle` hijos. El ciclo de vida completo de una tarea sigue cuatro pasos: (1) `newHandle` asigna e inicializa un `taskHandle` con un canal `done` abierto; (2) se registra como hijo de la tarea padre si existe; (3) la goroutine se registra en `taskMap` mediante su ID; y (4) al finalizar, cierra `done` y se desregistra de `taskMap`.

El mapa global `taskMap` asocia identificadores de goroutine (obtenidos mediante `getGoroutineID`, definida en `internal.go` como utilidad compartida del paquete) con sus `taskHandle` activos. El acceso a este mapa está protegido por un `sync.RWMutex` que permite lecturas concurrentes desde `currentTask` y escrituras exclusivas desde `registerTask` y `unregisterTask`.

`waitSubtree` realiza un recorrido recursivo en profundidad sobre el subárbol de `h`: bloquea en el canal `done` de cada nodo hijo antes de descender a sus propios hijos, garantizando que ningún descendiente quede vivo al retornar. Cuando se invoca desde `Taskgroup`, el canal `done` del handle raíz ya está cerrado (el hilo que ejecuta `Taskgroup` lo cierra justo antes de llamar a `waitSubtree`), por lo que la espera retorna inmediatamente en el nodo raíz y la sincronización efectiva ocurre en los nodos hijos y sus descendientes.

=== Registro de dependencias (`depend.go`)

El mapa global `depRegistry` asocia direcciones de variables (representadas como `uintptr`) con entradas `depEntry`. Cada `depEntry` mantiene el canal `done` del último escritor activo sobre esa dirección (`writerDone`) y la lista de canales `done` de todos los lectores activos concurrentes (`readersDone`).

La función `claimDeps` adquiere el mutex global, calcula el conjunto de señales que la tarea debe esperar según la semántica de cada modo, y actualiza el registro bajo el mismo lock para garantizar la atomicidad entre la consulta y el registro:

- *`in` (lectura):* espera al escritor previo si existe; se registra como lector activo para que futuros escritores lo esperen.
- *`out` (escritura):* espera al escritor previo y a todos los lectores activos; reemplaza la entrada como nuevo escritor y limpia la lista de lectores.
- *`inout` (lectura-escritura):* idéntico a `out` - serializa con todos los predecesores y se convierte en el nuevo escritor exclusivo.

La espera efectiva sobre las señales recolectadas ocurre fuera del lock, dentro de la goroutine de la tarea, para no bloquear el ciclo de despacho del hilo que invoca `TaskWithDepend`.

== Metodología de pruebas

La suite de pruebas se organiza por primitiva pública, cubriendo para cada una el comportamiento correcto bajo entradas válidas, el manejo seguro de casos límite (cero iteraciones, grainsize inválido, ausencia de escritor previo) y la corrección bajo condiciones de concurrencia. Las pruebas de `task.go` verifican la asincronía de `Task`, la semántica de espera de `Taskwait` frente a `Taskgroup`, y la distribución correcta de iteraciones por parte de `Taskloop`. Las pruebas de `depend.go` verifican cada patrón de dependencia de forma aislada, asegurando que el ordenamiento observado en tiempo de ejecución coincide con el ordering declarado.
