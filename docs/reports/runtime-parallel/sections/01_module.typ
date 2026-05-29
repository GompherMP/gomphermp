= Descripción del Módulo

El módulo de gestión de goroutines y reparto de trabajo constituye el primer subsistema de la librería de runtime de GompherMP. Recibe como entrada las invocaciones generadas por el motor de transformación del AST y se encarga de despachar trabajo a un pool persistente de goroutines, distribuir las iteraciones de los bucles paralelos entre los workers del pool y ejecutar bloques independientes en paralelo. Provee las primitivas de ejecución que soportan las directivas `parallel`, `for`, `parallel for` y `sections` formalizadas en la especificación R1.

== Ubicación

El módulo reside en el paquete público `pkg/runtime/` del repositorio. El siguiente cuadro detalla los archivos que componen el módulo:

#figure(
  table(
    columns: (auto, 1fr),
    align: (left, left),
    [*Archivo*],            [*Responsabilidad*],
    [`pool.go`],            [Implementa el pool persistente de goroutines, su ciclo de vida y las funciones públicas de configuración (`PoolSize`, `SetPoolSize`, `CurrentTeamSize`).],
    [`pool_test.go`],       [Pruebas unitarias del pool: inicialización temprana, redimensionamiento, persistencia y ejecución concurrente de los workers.],
    [`parallel.go`],        [Implementa las primitivas públicas (`Parallel`, `For`, `ParallelFor`, `ForDynamic`, `Sections`) y las estructuras internas que sostienen el team context.],
    [`parallel_test.go`],   [Suite completa de pruebas unitarias de las primitivas de reparto de trabajo.],
  ),
  caption: [Archivos que componen el módulo de gestión de goroutines y reparto de trabajo],
)

== Primitivas públicas

El módulo expone cinco funciones que el motor de transformación invoca para traducir las directivas del programa original, complementadas por tres funciones de configuración del pool:

#figure(
  table(
    columns: (auto, 1fr),
    align: (left, left),
    [*Primitiva*], [*Propósito*],
    [`Parallel`],          [Despacha al pool tantos jobs como workers tenga el pool, cada uno asociado a un team context compartido. Cada worker recibe su identificador de hilo y ejecuta el cuerpo de la región paralela. Aplica una barrera implícita al finalizar. Cuando se invoca desde dentro de una región paralela ya activa, serializa la ejecución en un equipo virtual de tamaño uno (paralelismo anidado deshabilitado, consistente con el comportamiento por defecto de OpenMP).],
    [`For`],               [Reparte un espacio de iteración entre los workers del pool mediante reparto estático en chunks contiguos.],
    [`ParallelFor`],       [Combinación de `Parallel` y `For` en una única invocación para el caso compuesto `parallel for`.],
    [`ForDynamic`],        [Reparte iteraciones mediante un contador atómico compartido que entrega chunks bajo demanda. Implementa la política `schedule(dynamic, chunkSize)`.],
    [`Sections`],          [Distribuye dinámicamente un conjunto de bloques de código independientes entre los workers del pool.],
    [`PoolSize`],          [Retorna el tamaño del pool activo.],
    [`SetPoolSize`],       [Reemplaza el pool activo por uno nuevo del tamaño solicitado. Valores no positivos se acotan al mínimo seguro de uno.],
    [`CurrentTeamSize`],   [Retorna el tamaño del equipo al que pertenece la goroutine invocante, o uno si no está dentro de ninguna región paralela.],
  ),
  caption: [Primitivas públicas del módulo de gestión de goroutines y reparto de trabajo],
)

== Modelo de concurrencia

Al inicializarse el paquete (`init()` en `pool.go`), el runtime pre-instancia un pool de goroutines persistentes dimensionado a `runtime.GOMAXPROCS(0)`. Los workers son goroutines de larga vida que consumen jobs de un canal compartido y permanecen activas durante toda la ejecución del programa, eliminando el overhead de creación y destrucción que tendría un enfoque "spawn por región". Esta arquitectura replica el patrón Worker Pool empleado por implementaciones de referencia del runtime OpenMP.

El tamaño del pool se controla mediante la función `SetPoolSize(n)`, que crea un nuevo pool de tamaño `n` y libera los workers del pool anterior cerrando el canal de jobs. Valores no positivos son acotados al mínimo seguro de uno, garantizando que el runtime mantenga un estado operativo incluso ante configuraciones degeneradas.

Cada región paralela instancia un team context, una estructura que asocia los workers participantes mediante un grupo de espera (`sync.WaitGroup`) que sostiene la barrera implícita y permite identificar al equipo en operaciones posteriores como `Barrier()`. Cuando un worker del pool toma un job, se registra en el team context asociado al job por su identificador de goroutine en un mapa global protegido por un `sync.RWMutex`, habilitando el lookup eficiente desde primitivas de sincronización ajenas al módulo. Al finalizar la ejecución del cuerpo del job, el worker se desregistra y queda disponible para participar en regiones paralelas posteriores con team contexts distintos.

== Metodología de pruebas

La suite de pruebas del módulo se organiza por primitiva pública, cubriendo para cada una el comportamiento correcto bajo entradas válidas, el manejo seguro de casos límite (cero iteraciones, iteraciones negativas, configuraciones de hilos inválidas) y la corrección bajo condiciones de concurrencia intensiva. Adicionalmente, se incluyen pruebas que verifican la distribución efectiva del trabajo entre múltiples goroutines y pruebas de estrés diseñadas para detectar condiciones de carrera mediante el detector de carreras (`go test -race`).
