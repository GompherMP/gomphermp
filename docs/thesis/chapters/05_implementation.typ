
= Capítulo 5. OE2: Implementación del compilador y runtime de la herramienta GompherMP

En este capítulo se detalla el trabajo realizado para cumplir con el segundo objetivo específico, el cual comprende la implementación de la infraestructura central de GompherMP. A diferencia del primer objetivo, centrado en la definición formal y el diseño arquitectónico, este objetivo se materializa en componentes ejecutables y dotados de pruebas automatizadas. Los resultados aquí presentados toman como punto de partida los artefactos producidos en el OE1 (la especificación de directivas, el diseño de módulos y la especificación de la CLI) y los traducen a una implementación operativa siguiendo el patrón de compilador source-to-source establecido.


== R4: Módulo de gestión de goroutines y reparto de trabajo

Para cumplir con este resultado, se implementó el módulo de gestión de goroutines y reparto de trabajo, responsable de materializar las directivas de paralelismo estructurado de GompherMP sobre las primitivas nativas de concurrencia del lenguaje Go. Este módulo es el primero de los tres subsistemas que conforman la librería de runtime y constituye el soporte de ejecución sobre el cual operan las construcciones parallel, for y parallel for formalizadas en la especificación R1.

La implementación reside en el paquete pkg/runtime/ del repositorio del proyecto, principalmente en los archivos parallel.go y pool.go. El módulo expone un conjunto de funciones públicas que el motor de transformación invoca para reescribir las directivas del código fuente original. La función Parallel instancia un equipo de goroutines y orquesta su ejecución dentro de una región paralela, registrando cada goroutine en un contexto de equipo compartido que habilita la sincronización mediante barreras. Las funciones For y ParallelFor distribuyen el espacio de iteración de un bucle entre los miembros del equipo aplicando una política de planificación estática por defecto, mientras que las variantes ForStaticChunked y ParallelForStaticChunked permiten un reparto estático por bloques de tamaño configurable. Adicionalmente, el módulo provee ForDynamic y ParallelForDynamic, que implementan reparto dinámico mediante un contador atómico compartido para escenarios con cargas heterogéneas, así como Sections y ParallelSections para la distribución de bloques de código independientes entre el equipo.

El paralelismo se sustenta en un pool de goroutines persistentes, implementado en pool.go, que evita el costo de crear y destruir hilos en cada región paralela. El tamaño del equipo se inicializa por defecto al valor de runtime.GOMAXPROCS(0) —es decir, el número de núcleos lógicos disponibles— y puede reconfigurarse en tiempo de ejecución mediante la función SetPoolSize, lo que permite controlar el grado de paralelismo aplicado a la ejecución.

La suite de pruebas asociada al módulo se compone de pruebas unitarias que cubren cada función pública del paquete, incluyendo casos límite (cero iteraciones, iteraciones negativas, configuraciones de hilos inválidas), pruebas de distribución efectiva del trabajo entre múltiples goroutines y pruebas de estrés bajo el detector de carreras del compilador (go test -race) para verificar la ausencia de condiciones de carrera. La cobertura de instrucciones alcanzada sobre el módulo es del 100%, verificada mediante la herramienta nativa go tool cover. El código fuente completo y el informe técnico de cobertura se encuentran disponibles en el repositorio en línea del proyecto, referenciado en los Anexos G y H respectivamente.

#figure(
  image("../figures/fig5_1_coverage_goroutines.png"),
  caption: [Extracto del informe de cobertura de pruebas del módulo de gestión de goroutines],
) <fig:figura-5-1-extracto-del-inform>


== R5: Módulo de mecanismos de sincronización

Para este resultado, se implementó el módulo de mecanismos de sincronización, responsable de coordinar la ejecución de las goroutines dentro de una región paralela. Este módulo provee las primitivas que el código transformado utiliza para implementar las directivas de sincronización especificadas en R1, garantizando la consistencia de los accesos a memoria compartida y la correcta orquestación de bloques de ejecución exclusivos.

La implementación reside en el archivo sync.go del paquete pkg/runtime/ y expone cuatro funciones públicas que cubren los mecanismos de sincronización del subconjunto OpenMP soportado. La función Critical garantiza exclusión mutua sobre un bloque de código, soportando tanto la modalidad anónima (mediante un mutex global compartido) como la modalidad nominal (mediante mutexes asociados a un identificador). La función Single garantiza que el cuerpo se ejecute exactamente una vez dentro de un equipo: el primer goroutine que gana una operación atómica de comparación e intercambio (compare-and-swap) sobre un testigo del equipo ejecuta el bloque, y la barrera implícita que cierra la construcción reinicia dicho testigo para la siguiente región. La función Master ejecuta condicionalmente un bloque únicamente en la goroutine maestra del equipo, sin imponer una barrera implícita posterior. Finalmente, la función Barrier establece un punto de sincronización explícito mediante un grupo de espera dimensionado al tamaño del equipo, garantizando que ninguna goroutine continúe su ejecución hasta que todas hayan alcanzado el punto.

El diseño del módulo se apoya en las primitivas de sincronización de la librería estándar de Go, específicamente sync.Mutex y sync.WaitGroup, complementadas con operaciones atómicas del paquete sync/atomic para la elección del ejecutor en las construcciones de bloque único. Esta decisión permite delegar al runtime nativo del lenguaje las garantías de orden de memoria requeridas por el modelo de concurrencia, evitando duplicar mecanismos de bajo nivel ya provistos por el ecosistema.

La suite de pruebas del módulo verifica el comportamiento correcto de cada primitiva sobre regiones paralelas reales, incluyendo casos de protección contra condiciones de carrera (mediante incrementos concurrentes sobre contadores compartidos), independencia entre locks nombrados, exclusividad del bloque maestro respecto a las goroutines no maestras, ausencia de barrera implícita en Master, y comportamiento correcto del Barrier bajo distintos tamaños de equipo. Adicionalmente, se incluyen pruebas de tiempo límite que detectan posibles bloqueos por errores de implementación. La cobertura de instrucciones alcanzada sobre el módulo es del 100%, verificada mediante la herramienta nativa go tool cover. El código fuente completo y el informe técnico de cobertura se encuentran disponibles en el repositorio en línea del proyecto, referenciado en el Anexo G, y en el Anexo I respectivamente.

#figure(
  image("../figures/fig5_2_coverage_sync.png"),
  caption: [Extracto del informe de cobertura de pruebas de los mecanismos de sincronización],
) <fig:figura-5-2-extracto-del-inform>


== R6: Módulo de soporte para paralelismo de tareas

Para cumplir con este resultado, se implementó el módulo de soporte para paralelismo de tareas, responsable de materializar las directivas de paralelismo no estructurado de GompherMP sobre las primitivas nativas de concurrencia de Go. Este módulo implementa un modelo de ejecución asíncrono orientado a tareas con dependencias de datos explícitas, correspondiente a las directivas task, taskwait, taskgroup y taskloop formalizadas en la especificación R1.

La implementación reside en el paquete pkg/runtime/ y se organiza en dos archivos complementarios. El archivo task.go expone cuatro funciones públicas: Task, que somete una función como tarea asíncrona y la vincula al árbol de la tarea activa; Taskwait, que bloquea hasta que concluyan todas las hijas directas de la tarea actual; Taskgroup, que ejecuta un cuerpo y aguarda el subárbol completo de descendientes a cualquier profundidad; y Taskloop, que distribuye un espacio de iteraciones como tareas independientes parametrizadas por un tamaño de bloque. El archivo depend.go extiende el modelo con la función TaskWithDepend, que acepta tokens de dirección de memoria para los roles in, out e inout y garantiza el orden happens-before entre tareas mediante un esquema de seguimiento por frontera: por cada token, el registro mantiene únicamente el canal del último escritor activo y los canales de los lectores activos, colectando atómicamente las señales de espera necesarias sin construir un grafo explícito. Adicionalmente, Taskgroup incorpora un contador de anidamiento que dispara la limpieza del registro de dependencias al retornar el grupo más externo, evitando el crecimiento no acotado del mapa entre regiones de tareas sucesivas.

#figure(
  image("../figures/fig5_3_coverage_tasks.png"),
  caption: [Extracto del informe de cobertura de pruebas del módulo de soporte para paralelismo de tareas],
) <fig:figura-5-3-extracto-del-inform>


== R7: Analizador sintáctico de las directivas GompherMP

Se implementó el módulo Parser, el componente de entrada del compilador GompherMP. Este módulo es el responsable de transformar el código fuente Go anotado con comentarios de la forma \/\/gompher \<directiva\> en una representación intermedia tipada que el motor de transformación puede consumir directamente.

La implementación reside en el directorio internal/parser/ del repositorio y se apoya en las facilidades de la librería estándar de Go para el análisis sintáctico, evitando duplicar funcionalidad nativa del lenguaje. El parser delega la construcción del árbol sintáctico abstracto (AST) a la función parser.ParseFile de la librería go/parser, y a continuación realiza su propio análisis sobre dicho AST para extraer las directivas relevantes. Esta estrategia garantiza compatibilidad estricta con la gramática de Go: cualquier programa válido sigue siendo procesable por GompherMP sin alterar la semántica del lenguaje anfitrión.

El procesamiento del módulo se organiza en seis etapas que ejecutan secuencialmente sobre cada archivo de entrada:

Análisis sintáctico del código fuente Go mediante la librería estándar, obteniendo el AST nativo y la lista completa de comentarios del archivo.

Construcción de un CommentMap que asocia cada comentario con el nodo del AST que físicamente le sigue. Este mecanismo, provisto por la librería estándar, permite vincular cada directiva con la sentencia que anota sin recurrir a heurísticas frágiles basadas en números de línea.

Recorrido del AST identificando los comentarios cuyo prefijo es \/\/gompher y construyendo, para cada uno, una estructura tipada que captura el tipo de directiva, sus cláusulas, su posición en el código fuente y el nodo del AST que gobierna.

Validación semántica de cada directiva, comprendiendo cuatro categorías de reglas: compatibilidad del tipo de nodo objetivo (por ejemplo, \/\/gompher for solo es válida sobre un bucle for), adyacencia física entre el comentario y su nodo (sin líneas en blanco intermedias), legalidad de las cláusulas declaradas para la directiva en cuestión y exigencia de listas no vacías para cláusulas que operan sobre variables.

Validación de contexto jerárquico, que verifica relaciones estructurales entre directivas (principalmente que toda directiva section aparezca dentro de un bloque sections).

Ordenamiento final del conjunto de directivas extraídas según su línea de aparición en el código fuente, preservando el orden de ejecución original para que el motor de transformación procese el archivo de manera cronológica.

El diseño del parser se basa en dos interfaces selladas: Directive y Clause. Estas solo pueden ser implementadas dentro del paquete. La interfaz Directive agrupa los quince tipos concretos correspondientes a las directivas soportadas por GompherMP (parallel, for, parallel for, sections, parallel sections, section, single, master, critical, barrier, atomic, task, taskwait, taskgroup y taskloop), mientras que Clause agrupa los ocho tipos correspondientes a las cláusulas (private, firstprivate, lastprivate, shared, reduction, schedule, depend y grainsize).

Las pruebas asociadas al módulo se componen de noventa y nueve pruebas organizadas en seis categorías: parseo de cláusulas en aislamiento, parseo de directivas en aislamiento, integración completa sobre programas Go reales, validaciones semánticas, contratos de interfaz y pruebas internas auxiliares. La cobertura de instrucciones alcanzada sobre el módulo es del 100% verificada mediante la herramienta nativa go tool cover. El documento completo, incluyendo el mapeo entre cada directiva y cláusula, y las pruebas que verifican su análisis sintáctico, se encuentra en el Anexo K.

A modo de ilustración, la Figura 9 muestra un extracto del informe de cobertura del parser, evidenciando la trazabilidad entre las directivas formalizadas en R1 y las pruebas automatizadas que validan su soporte.

#figure(
  image("../figures/fig5_4_coverage_parser.png"),
  caption: [Extracto del informe de cobertura del módulo Parser],
) <fig:figura-5-4-extracto-del-inform>


== Discusión de resultados

La implementación del OE2 confirma en la práctica que las decisiones de diseño formuladas en el OE1 son viables. Al construir el runtime utilizando únicamente las herramientas de concurrencia que la librería estándar de Go ya provee (mutexes, grupos de espera y goroutines), se evidencia que el lenguaje ofrece todos los componentes necesarios para soportar un conjunto representativo de las construcciones del estándar OpenMP. La verificación de los módulos mediante el detector de carreras nativo (go test -race) bajo escenarios de estrés con miles de iteraciones y contención máxima sobre estructuras compartidas confirma adicionalmente que las garantías de orden de memoria provistas por la librería estándar son suficientes para implementar primitivas de paralelismo correctas, evitando la introducción de las condiciones de carrera y bloqueos que típicamente afectan la implementación manual del paralelismo en este tipo de lenguajes.

Por su parte, el módulo Parser materializa el principio de "fallo temprano" propio de los frontends de compiladores robustos. Las cuatro categorías de validación semántica implementadas (compatibilidad del tipo de nodo objetivo, adyacencia comentario-bloque, contexto jerárquico de directivas anidadas y exigencia de listas no vacías para cláusulas de variables) aseguran que cualquier inconsistencia se detecte y reporte durante el análisis sintáctico, antes que el motor de transformación intente operar sobre estructuras inválidas y produzca errores de difícil diagnóstico. No obstante, el modelo de comentarios con valor semántico (\/\/gompher) hereda una fragilidad inherente al paradigma OpenMP: la dependencia estricta de la adyacencia física entre el comentario y su nodo objetivo, que puede ser interrumpida por líneas en blanco u otros comentarios intercalados. Este compromiso constituye una tensión inevitable al utilizar metadatos textuales como portadores de semántica de ejecución, y refuerza la necesidad de que la documentación dirigida al usuario final enfatice con claridad las reglas de uso correcto de cada directiva.
