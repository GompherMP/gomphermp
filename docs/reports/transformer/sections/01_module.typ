= Descripción del Módulo

El módulo transformador es el motor de reescritura de código de GompherMP. Recibe
como entrada el AST anotado que produce el analizador sintáctico y, para cada
directiva `//gompher` reconocida, sustituye el nodo objetivo por las llamadas a
la librería de runtime que materializan su semántica. El resultado es un AST
equivalente que, una vez serializado, es un programa Go nativo compilable con el
toolchain estándar.

A diferencia de un compilador a código máquina, el transformador opera
enteramente sobre el AST: no genera instrucciones de bajo nivel, sino que inyecta
llamadas a funciones de runtime (`Parallel`, `For`, `Critical`, las primitivas
atómicas, etc.) y declara las copias de variables que exigen las cláusulas de
gestión de datos. La resolución de tipos para esas copias se delega a `go/types`,
y la normalización de los bucles a su forma canónica replica la estrategia que
emplean los compiladores de OpenMP.

== Ubicación

El módulo reside en el paquete `internal/transformer/` del repositorio. El
siguiente cuadro agrupa sus archivos por responsabilidad:

#figure(
  table(
    columns: (auto, 1fr),
    align: (left, left),
    [*Archivo*],            [*Responsabilidad*],
    [`transformer.go`],     [Orquestador: recorre los nodos anotados, despacha cada directiva a su manejador e inyecta el import del runtime.],
    [`validate.go`],        [Validación contextual previa a la reescritura (por ejemplo, que `for` o `single` aparezcan dentro de una región paralela).],
    [`parallel.go`, `loop.go`, `sections.go`, `single.go`, `master.go`, `critical.go`, `barrier.go`, `atomic.go`, `task.go`, `taskloop.go`], [Manejadores de directiva: una unidad por construcción (parallel, for, parallel for, sections, single, master, critical, barrier, atomic, task, taskwait, taskgroup, taskloop).],
    [`loopform.go`],        [Análisis y normalización de la forma canónica del bucle (límite inferior, operador relacional, paso y dirección).],
    [`clauses.go`],         [Maquinaria de las cláusulas de gestión de datos (private, firstprivate, shared, lastprivate, reduction).],
    [`typeinfo.go`],        [Resolución del tipo de una variable de cláusula mediante `go/types`.],
    [`astbuild.go`],        [Constructores puros de nodos de AST (closures, llamadas de runtime, literales, clonación de expresiones).],
    [`astreplace.go`],      [Mutadores del AST: sustitución del nodo objetivo, inserción de sentencias y eliminación del comentario consumido.],
    [`imports.go`],         [Inyección idempotente de los imports requeridos (runtime, unsafe).],
  ),
  caption: [Archivos que componen el módulo transformador],
)

== Pipeline de transformación

El módulo se invoca una vez por archivo, después del análisis sintáctico y de la
validación contextual. Para cada nodo anotado ejecuta cuatro etapas:

#figure(
  table(
    columns: (auto, 1fr),
    align: (left, left),
    [*Etapa*], [*Descripción*],
    [Despacho], [El orquestador identifica el tipo de directiva y lo dirige a su manejador.],
    [Análisis], [El manejador extrae la información necesaria del nodo: variable de inducción y forma del bucle, lista de cláusulas, nombre de la región crítica, etc.],
    [Construcción], [Se arman los nodos de reemplazo: la closure con el cuerpo original, las declaraciones de las copias privadas, las capturas previas y la llamada de runtime.],
    [Sustitución], [El nodo objetivo se reemplaza en el AST por la llamada de runtime, se inyectan las sentencias auxiliares y se elimina el comentario de la directiva.],
  ),
  caption: [Etapas del flujo de transformación por directiva],
)

Al terminar, el orquestador inyecta el import del runtime si se emitió al menos
una llamada, y entrega el AST transformado a la etapa de serialización.

== Metodología de pruebas

La suite de pruebas del módulo se organiza por directiva y por la maquinaria
transversal que comparten (constructores y mutadores de AST, resolución de tipos,
cláusulas de datos). Para cada manejador se verifica la estructura del código
emitido sobre entradas válidas, el manejo de los casos límite y el rechazo
explícito de las construcciones no soportadas con un mensaje descriptivo. Las
pruebas comparan el código generado contra el resultado esperado y, en los
escenarios de extremo a extremo, transpilan y compilan el programa resultante
para confirmar que el toolchain estándar de Go lo acepta.
