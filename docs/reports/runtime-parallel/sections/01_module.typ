= Descripción del Módulo

El módulo de gestión de goroutines y reparto de trabajo constituye el primer subsistema de la librería de runtime de GompherMP. Recibe como entrada las invocaciones generadas por el motor de transformación del AST y se encarga de instanciar los equipos de goroutines, distribuir las iteraciones de los bucles paralelos entre los miembros del equipo y ejecutar bloques independientes en paralelo. Provee las primitivas de ejecución que soportan las directivas `parallel`, `for`, `parallel for` y `sections` formalizadas en la especificación R1.

== Ubicación

El módulo reside en el paquete público `pkg/runtime/` del repositorio. El siguiente cuadro detalla los archivos que componen el módulo:

#figure(
  table(
    columns: (auto, 1fr),
    align: (left, left),
    [*Archivo*],            [*Responsabilidad*],
    [`parallel.go`],        [Implementa las primitivas públicas (`Parallel`, `For`, `ParallelFor`, `ForDynamic`, `Sections`) y las estructuras internas que sostienen el team context.],
    [`parallel_test.go`],   [Suite completa de pruebas unitarias del módulo.],
  ),
  caption: [Archivos que componen el módulo de gestión de goroutines y reparto de trabajo],
)

== Primitivas públicas

El módulo expone cinco funciones que el motor de transformación invoca para traducir las directivas del programa original:

#figure(
  table(
    columns: (auto, 1fr),
    align: (left, left),
    [*Primitiva*], [*Propósito*],
    [`Parallel`],     [Instancia un equipo de N goroutines coordinadas por un team context. Cada goroutine recibe su identificador y ejecuta el cuerpo de la región paralela. Aplica una barrera implícita al finalizar.],
    [`For`],          [Reparte un espacio de iteración entre las goroutines del equipo mediante reparto estático en chunks contiguos.],
    [`ParallelFor`],  [Combinación de `Parallel` y `For` en una única invocación para el caso compuesto `parallel for`.],
    [`ForDynamic`],   [Reparte iteraciones mediante un contador atómico compartido que entrega chunks bajo demanda. Implementa la política `schedule(dynamic, chunkSize)`.],
    [`Sections`],     [Distribuye dinámicamente un conjunto de bloques de código independientes entre los miembros del equipo.],
  ),
  caption: [Primitivas públicas del módulo de gestión de goroutines y reparto de trabajo],
)

== Modelo de concurrencia

El tamaño del equipo se determina mediante la variable pública `NumThreads`, inicializada al número de núcleos disponibles reportado por la librería estándar de Go (`runtime.NumCPU()`). Este parámetro puede ser reasignado por el programa antes de invocar cualquier primitiva, lo que permite controlar el grado de paralelismo. Valores no válidos (menores o iguales a cero) son corregidos automáticamente al mínimo seguro de un hilo.

Cada región paralela instancia un team context, una estructura que asocia las goroutines del equipo entre sí mediante un grupo de espera (`sync.WaitGroup`) que sostiene la barrera implícita y permite identificar al equipo en operaciones posteriores como `Barrier()`. El team context se registra por identificador de goroutine en un mapa global protegido por un `sync.RWMutex`, lo que habilita el lookup eficiente desde primitivas de sincronización ajenas al módulo.

== Metodología de pruebas

La suite de pruebas del módulo se organiza por primitiva pública, cubriendo para cada una el comportamiento correcto bajo entradas válidas, el manejo seguro de casos límite (cero iteraciones, iteraciones negativas, configuraciones de hilos inválidas) y la corrección bajo condiciones de concurrencia intensiva. Adicionalmente, se incluyen pruebas que verifican la distribución efectiva del trabajo entre múltiples goroutines y pruebas de estrés diseñadas para detectar condiciones de carrera mediante el detector de carreras (`go test -race`).
