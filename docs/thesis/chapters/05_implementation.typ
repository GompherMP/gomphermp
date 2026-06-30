
= Capítulo 5. OE2: Implementación del compilador y runtime de la herramienta GompherMP

En este capítulo se detalla el trabajo realizado para cumplir con el segundo objetivo específico, el cual comprende la implementación de la infraestructura central de GompherMP. A diferencia del primer objetivo, centrado en la definición formal y el diseño arquitectónico, este objetivo se materializa en componentes ejecutables y dotados de pruebas automatizadas. Los resultados aquí presentados toman como punto de partida los artefactos producidos en el OE1 (la especificación de directivas, el diseño de módulos y la especificación de la CLI) y los traducen a una implementación operativa siguiendo el patrón de compilador _source-to-source_ establecido.


== R4: Módulo de gestión de goroutines y reparto de trabajo

Para cumplir con este resultado, se implementó el módulo de gestión de goroutines y reparto de trabajo, responsable de materializar las directivas de paralelismo estructurado de GompherMP sobre las primitivas nativas de concurrencia del lenguaje Go. Este módulo es el primero de los tres subsistemas que conforman la librería de runtime y constituye el soporte de ejecución sobre el cual operan las construcciones `parallel`, `for` y `parallel for` formalizadas en la especificación R1.

La implementación reside en el paquete `pkg/runtime/` del repositorio del proyecto, principalmente en los archivos `parallel.go` y `pool.go`. El módulo expone un conjunto de funciones públicas que el motor de transformación invoca para reescribir las directivas del código fuente original. La función `Parallel` instancia un equipo de goroutines y orquesta su ejecución dentro de una región paralela, registrando cada goroutine en un contexto de equipo compartido que habilita la sincronización mediante barreras. Las funciones `For` y `ParallelFor` distribuyen el espacio de iteración de un bucle entre los miembros del equipo aplicando una política de planificación estática por defecto, mientras que las variantes `ForStaticChunked` y `ParallelForStaticChunked` permiten un reparto estático por bloques de tamaño configurable. Adicionalmente, el módulo provee `ForDynamic` y `ParallelForDynamic`, que implementan reparto dinámico mediante un contador atómico compartido para escenarios con cargas heterogéneas, así como `Sections` y `ParallelSections` para la distribución de bloques de código independientes entre el equipo.

El paralelismo se sustenta en un _pool_ de goroutines persistentes, implementado en `pool.go`, que evita el costo de crear y destruir hilos en cada región paralela. El tamaño del equipo se inicializa por defecto al valor de `runtime.GOMAXPROCS(0)`, que corresponde al número de núcleos lógicos disponibles. Este valor puede reconfigurarse en tiempo de ejecución mediante la función `SetPoolSize`, lo que permite controlar el grado de paralelismo aplicado a la ejecución.

La suite de pruebas asociada al módulo se compone de pruebas unitarias que cubren cada función pública del paquete, incluyendo casos límite (cero iteraciones, iteraciones negativas, configuraciones de hilos inválidas), pruebas de distribución efectiva del trabajo entre múltiples goroutines y pruebas de estrés bajo el detector de carreras del compilador (`go test -race`) para verificar la ausencia de condiciones de carrera. La cobertura de instrucciones alcanzada sobre el módulo es del 100%, verificada mediante la herramienta nativa `go tool cover`. El código fuente completo y el informe técnico de cobertura se encuentran disponibles en el repositorio en línea del proyecto, referenciado en los Anexos G y H respectivamente.

#figure(
  image("../figures/fig5_1_coverage_goroutines.png"),
  caption: [Extracto del informe de cobertura de pruebas del módulo de gestión de goroutines],
) <fig:figura-5-1-extracto-del-inform>


== R5: Módulo de mecanismos de sincronización

Para este resultado, se implementó el módulo de mecanismos de sincronización, responsable de coordinar la ejecución de las goroutines dentro de una región paralela. Este módulo provee las primitivas que el código transformado utiliza para implementar las directivas de sincronización especificadas en R1, garantizando la consistencia de los accesos a memoria compartida y la correcta orquestación de bloques de ejecución exclusivos.

La implementación reside en el archivo `sync.go` del paquete `pkg/runtime/` y expone cuatro funciones públicas que cubren los mecanismos de sincronización del subconjunto OpenMP soportado. La función `Critical` garantiza exclusión mutua sobre un bloque de código, soportando tanto la modalidad anónima (mediante un mutex global compartido) como la modalidad nominal (mediante mutexes asociados a un identificador). La función `Single` garantiza que el cuerpo se ejecute exactamente una vez dentro de un equipo: el primer goroutine que gana una operación atómica de comparación e intercambio (_compare-and-swap_) sobre un testigo del equipo ejecuta el bloque, y la barrera implícita que cierra la construcción reinicia dicho testigo para la siguiente región. La función `Master`#footnote[La directiva master fue deprecada en OpenMP 5.1 (2021) en favor de masked y removida en OpenMP 6.0. Se conserva en GompherMP por ser la forma canónica en la literatura de referencia sobre la que se modeló el subconjunto soportado.] ejecuta condicionalmente un bloque únicamente en la goroutine maestra del equipo, sin imponer una barrera implícita posterior. Finalmente, la función `Barrier` establece un punto de sincronización explícito y reutilizable para el equipo, construido sobre una variable de condición (`sync.Cond`) y un contador de generación, de modo que ninguna goroutine continúe su ejecución hasta que todas hayan alcanzado el punto y la misma barrera pueda emplearse de nuevo en las sucesivas construcciones de una región paralela.

El diseño del módulo se apoya en las primitivas de sincronización de la librería estándar de Go, específicamente `sync.Mutex` y `sync.Cond`, complementadas con operaciones atómicas del paquete `sync/atomic` para la elección del ejecutor en las construcciones de bloque único. Esta decisión permite delegar al runtime nativo del lenguaje las garantías de orden de memoria requeridas por el modelo de concurrencia, evitando duplicar mecanismos de bajo nivel ya provistos por el ecosistema.

La suite de pruebas del módulo verifica el comportamiento correcto de cada primitiva sobre regiones paralelas reales, incluyendo casos de protección contra condiciones de carrera (mediante incrementos concurrentes sobre contadores compartidos), independencia entre locks nombrados, exclusividad del bloque maestro respecto a las goroutines no maestras, ausencia de barrera implícita en `Master`, y comportamiento correcto del `Barrier` bajo distintos tamaños de equipo. Adicionalmente, se incluyen pruebas de tiempo límite que detectan posibles bloqueos por errores de implementación. La cobertura de instrucciones alcanzada sobre el módulo es del 100%, verificada mediante la herramienta nativa `go tool cover`. El código fuente completo y el informe técnico de cobertura se encuentran disponibles en el repositorio en línea del proyecto, referenciado en el Anexo G, y en el Anexo I respectivamente.

#figure(
  image("../figures/fig5_2_coverage_sync.png"),
  caption: [Extracto del informe de cobertura de pruebas de los mecanismos de sincronización],
) <fig:figura-5-2-extracto-del-inform>


== R6: Módulo de soporte para paralelismo de tareas

Para cumplir con este resultado, se implementó el módulo de soporte para paralelismo de tareas, responsable de materializar las directivas de paralelismo no estructurado de GompherMP sobre las primitivas nativas de concurrencia de Go. Este módulo implementa un modelo de ejecución asíncrono orientado a tareas con dependencias de datos explícitas, correspondiente a las directivas `task`, `taskwait`, `taskgroup` y `taskloop` formalizadas en la especificación R1.

La implementación reside en el paquete `pkg/runtime/` y se organiza en dos archivos complementarios. El archivo `task.go` expone cuatro funciones públicas: `Task`, que somete una función como tarea asíncrona y la vincula al árbol de la tarea activa; `Taskwait`, que bloquea hasta que concluyan todas las hijas directas de la tarea actual; `Taskgroup`, que ejecuta un cuerpo y aguarda el subárbol completo de descendientes a cualquier profundidad; y `Taskloop`, que distribuye un espacio de iteraciones como tareas independientes parametrizadas por un tamaño de bloque. El archivo `depend.go` extiende el modelo con la función `TaskWithDepend`, que acepta tokens de dirección de memoria para los roles in, out e inout y garantiza el orden _happens-before_ entre tareas mediante un esquema de seguimiento por frontera: por cada token, el registro mantiene únicamente el canal del último escritor activo y los canales de los lectores activos, colectando atómicamente las señales de espera necesarias sin construir un grafo explícito. Adicionalmente, `Taskgroup` incorpora un contador de anidamiento que dispara la limpieza del registro de dependencias al retornar el grupo más externo, evitando el crecimiento no acotado del mapa entre regiones de tareas sucesivas.

#figure(
  image("../figures/fig5_3_coverage_tasks.png"),
  caption: [Extracto del informe de cobertura de pruebas del módulo de soporte para paralelismo de tareas],
) <fig:figura-5-3-extracto-del-inform>


== R7: Analizador sintáctico de las directivas GompherMP

Se implementó el módulo Parser, el componente de entrada del compilador GompherMP. Este módulo es el responsable de transformar el código fuente Go anotado con comentarios de la forma \/\/gompher \<directiva\> en una representación intermedia tipada que el motor de transformación puede consumir directamente.

La implementación reside en el directorio `internal/parser/` del repositorio y se apoya en las facilidades de la librería estándar de Go para el análisis sintáctico, evitando duplicar funcionalidad nativa del lenguaje. El parser delega la construcción del árbol sintáctico abstracto (AST) a la función `parser.ParseFile` de la librería `go/parser`, y a continuación realiza su propio análisis sobre dicho AST para extraer las directivas relevantes. Esta estrategia garantiza compatibilidad estricta con la gramática de Go: cualquier programa válido sigue siendo procesable por GompherMP sin alterar la semántica del lenguaje anfitrión.

El procesamiento del módulo se organiza en seis etapas que ejecutan secuencialmente sobre cada archivo de entrada:

1. Análisis sintáctico del código fuente Go mediante la librería estándar, obteniendo el AST nativo y la lista completa de comentarios del archivo.

2. Construcción de un `CommentMap` que asocia cada comentario con el nodo del AST que físicamente le sigue. Este mecanismo, provisto por la librería estándar, permite vincular cada directiva con la sentencia que anota sin recurrir a heurísticas frágiles basadas en números de línea.

3. Recorrido del AST identificando los comentarios cuyo prefijo es \/\/gompher y construyendo, para cada uno, una estructura tipada que captura el tipo de directiva, sus cláusulas, su posición en el código fuente y el nodo del AST que gobierna.

4. Validación semántica de cada directiva, comprendiendo cuatro categorías de reglas: compatibilidad del tipo de nodo objetivo (por ejemplo, \/\/gompher `for` solo es válida sobre un bucle `for`), adyacencia física entre el comentario y su nodo (sin líneas en blanco intermedias), legalidad de las cláusulas declaradas para la directiva en cuestión y exigencia de listas no vacías para cláusulas que operan sobre variables.

5. Validación de contexto jerárquico, que verifica relaciones estructurales entre directivas (principalmente que toda directiva `section` aparezca dentro de un bloque `sections`).

6. Ordenamiento final del conjunto de directivas extraídas según su línea de aparición en el código fuente, preservando el orden de ejecución original para que el motor de transformación procese el archivo de manera cronológica.

El diseño del parser se basa en dos interfaces selladas: `Directive` y `Clause`. Estas solo pueden ser implementadas dentro del paquete. La interfaz `Directive` agrupa los quince tipos concretos correspondientes a las directivas soportadas por GompherMP (`parallel`, `for`, `parallel for`, `sections`, `parallel sections`, `section`, `single`, `master`, `critical`, `barrier`, `atomic`, `task`, `taskwait`, `taskgroup` y `taskloop`), mientras que `Clause` agrupa los ocho tipos correspondientes a las cláusulas (`private`, `firstprivate`, `lastprivate`, `shared`, `reduction`, `schedule`, `depend` y `grainsize`).

Las pruebas asociadas al módulo se componen de noventa y nueve pruebas organizadas en seis categorías: parseo de cláusulas en aislamiento, parseo de directivas en aislamiento, integración completa sobre programas Go reales, validaciones semánticas, contratos de interfaz y pruebas internas auxiliares. La cobertura de instrucciones alcanzada sobre el módulo es del 100% verificada mediante la herramienta nativa `go tool cover`. El documento completo, incluyendo el mapeo entre cada directiva y cláusula, y las pruebas que verifican su análisis sintáctico, se encuentra en el Anexo K.

A modo de ilustración, la Figura 9 muestra un extracto del informe de cobertura del parser, evidenciando la trazabilidad entre las directivas formalizadas en R1 y las pruebas automatizadas que validan su soporte.

#figure(
  image("../figures/fig5_4_coverage_parser.png"),
  caption: [Extracto del informe de cobertura del módulo Parser],
) <fig:figura-5-4-extracto-del-inform>


== R8: Motor de transformación del AST

El motor de transformación es el componente central del compilador GompherMP. Recibe la representación intermedia tipada que produce el analizador sintáctico y, por cada directiva reconocida, reescribe el árbol sintáctico abstracto (AST) sustituyendo el nodo anotado por las llamadas a la librería de runtime que materializan su semántica. El resultado es un AST equivalente que, una vez serializado, constituye un programa Go nativo compilable con el _toolchain_ estándar, sin modificación alguna al compilador del lenguaje.

La implementación reside en el directorio `internal/transformer/` del repositorio y se organiza por responsabilidades. Un orquestador (`transformer.go`) recorre los nodos anotados, despacha cada directiva a su manejador e inyecta el _import_ del runtime. Un manejador por construcción (`parallel.go`, `loop.go`, `sections.go`, `single.go`, `master.go`, `critical.go`, `barrier.go`, `atomic.go`, `task.go` y `taskloop.go`) realiza la reescritura específica de cada directiva. La lógica de las cláusulas de gestión de datos (`clauses.go`), el análisis y normalización de la forma canónica de los bucles (`loopform.go`) y la resolución de tipos mediante la librería estándar `go/types` (`typeinfo.go`) concentran la lógica transversal compartida por varios manejadores. Finalmente, los constructores y mutadores de bajo nivel del AST (`astbuild.go`, `astreplace.go` e `imports.go`) proveen las operaciones primitivas de construcción y sustitución de nodos.

Para cada nodo anotado, el motor ejecuta cuatro etapas: el despacho de la directiva hacia su manejador; el análisis del nodo para extraer la información necesaria (variable de inducción y forma del bucle, lista de cláusulas, nombre de la región crítica, entre otros); la construcción de los nodos de reemplazo (la _closure_ con el cuerpo original, las declaraciones de las copias privadas, las capturas previas y la llamada de runtime); y la sustitución del nodo objetivo en el AST junto con la eliminación del comentario consumido.

Más allá de la sustitución directa, el motor implementa las cinco cláusulas de gestión de datos resolviendo el tipo de cada variable con `go/types`, de modo que funciona incluso cuando el tipo se infiere de una asignación corta. La cláusula `private` declara una copia local de valor cero. La cláusula `firstprivate` captura el valor externo una vez antes de la región y lo usa para inicializar la copia. En cuanto a`shared`, no tiene un código específico, pues las _closures_ de Go capturan por referencia. `lastprivate` captura la dirección de la variable y la escribe de vuelta desde la última iteración o sección. `reduction` acumula sobre una copia privada inicializada con la identidad del operador y combina los parciales bajo una sección crítica, admitiendo siete operadores (`+`, `-`, `*`, `&&`, `||`, `max` y `min`). La directiva `atomic` traduce las operaciones de actualización, escritura y lectura a las primitivas atómicas del runtime. Adicionalmente, las directivas de bucle admiten cualquier bucle en forma canónica y no únicamente la forma `for i := 0; i < N; i++`: cuando la forma no es la estrecha, el motor la normaliza sobre el espacio de índices `[0, N)` que distribuye el runtime, reconstruyendo la variable de inducción del usuario al inicio del cuerpo, tal como un compilador de OpenMP baja un bucle canónico.

La siguiente figura ilustra la reescritura para el caso de un bucle paralelo con reducción, mostrando el código anotado de entrada y la llamada de runtime sintetizada.

#figure(
  block(stroke: 1pt + luma(180), inset: 10pt, width: 100%)[
    #align(left)[
      *Código anotado de entrada*
      ```go
      sum := 0
      //gompher parallel for reduction(+:sum)
      for i := 0; i < n; i++ {
          sum += f(i)
      }
      ```

      #v(0.8em)
      *Código transpilado de salida*
      ```go
      sum := 0
      _red_sum := &sum
      runtime.Parallel(func(threadID int) {
          var sum int = 0
          runtime.For(threadID, func(i int) {
              sum += f(i)
          }, n)
          runtime.Critical("", func() {
              *_red_sum += sum
          })
      })
      ```
    ]
  ],
  caption: [Reescritura de un bucle paralelo con reducción por el motor de transformación],
  kind: image,
) <fig:transformer-ejemplo>

Las pruebas asociadas al módulo se componen de ciento noventa pruebas organizadas por directiva y por la maquinaria transversal que comparten (constructores y mutadores de AST, resolución de tipos y cláusulas de datos), e incluyen escenarios de extremo a extremo que transpilan y compilan el programa resultante para confirmar que el _toolchain_ estándar de Go lo acepta. La cobertura de instrucciones alcanzada sobre el módulo es del 98.5%, verificada mediante la herramienta nativa `go tool cover`. La fracción no cubierta corresponde a guardas defensivas para condiciones que el flujo de ejecución no alcanza en operación normal. El informe técnico completo se encuentra en el Anexo M.

#figure(
  image("../figures/fig5_5_coverage_transformer.png"),
  caption: [Extracto del mapeo entre cada directiva y la llamada del runtime como parte del informe de cobertura del motor de transformación del AST],
) <fig:cobertura-transformer>

== R9: Herramienta GompherMP (CLI)

La interfaz de línea de comandos es el ejecutable que el usuario invoca para transpilar y compilar un programa. Coordina, en orden, a los módulos de la herramienta: lee el código fuente, lo entrega al analizador sintáctico, valida las reglas contextuales, aplica la transformación del AST, serializa el resultado a un archivo Go y finalmente invoca al compilador estándar de Go para producir el binario. La interfaz oculta estos pasos intermedios y expone una única operación de construcción.

La implementación reside en el paquete `cmd/gompher/` del repositorio y se apoya en el paquete `flag` de la librería estándar de Go para el análisis de las opciones, sin dependencias externas. El comando se invoca como `gompher build \<archivo.go\>` y admite las banderas `-o` (ruta del binario de salida), `-v` (modo detallado, que imprime las fases del flujo y las directivas detectadas), `-k` (conservar el archivo Go intermedio generado), `-h` (ayuda) y `--version`.

El comando de construcción ejecuta una secuencia de seis fases; cualquiera que falle detiene el flujo, emite un mensaje de error descriptivo y termina con un código de salida. Las fases incluyen la lectura, que valida la extensión y la existencia del archivo de entrada; el análisis, que extrae las directivas mediante el módulo Parser; la validación de las reglas contextuales; la transformación del AST; la serialización del AST transformado a un archivo Go temporal; y la compilación, que invoca al compilador estándar de Go sobre dicho archivo para producir el binario. El archivo temporal se elimina al finalizar, salvo que se solicite conservarlo.

Las pruebas asociadas al módulo se componen de veinte pruebas que ejercitan la interfaz a través de su punto de entrada testeable, capturando el código de salida y la salida de texto. Cubren el despacho de comandos, la validación de argumentos y opciones, cada modo de error de las fases, y un conjunto de pruebas de extremo a extremo que ejecutan el flujo completo, incluida la invocación real al compilador de Go, para confirmar que un programa anotado se transpila y compila correctamente. La cobertura de instrucciones alcanzada sobre el módulo es del 90.4%, verificada mediante la herramienta nativa `go tool cover`. La fracción no cubierta corresponde a la función envoltorio `main` y a ramas de manejo de fallos de entrada y salida que no se reproducen de forma portable en una prueba. En la siguiente figura, se muestra un extracto del informe de la CLI, pero el informe técnico completo se encuentra en el Anexo N.

#figure(
  image("../figures/fig5_6_coverage_cli.png"),
  caption: [Extracto del informe de cobertura de la interfaz de línea de comandos],
) <fig:cobertura-cli>

== Discusión de resultados

La implementación del OE2 confirma en la práctica que las decisiones de diseño formuladas en el OE1 son viables. Al construir el runtime utilizando únicamente las herramientas de concurrencia que la librería estándar de Go ya provee (mutexes, grupos de espera y goroutines), se evidencia que el lenguaje ofrece todos los componentes necesarios para soportar un conjunto representativo de las construcciones del estándar OpenMP. La verificación de los módulos mediante el detector de carreras nativo (`go test -race`) bajo escenarios de estrés con miles de iteraciones y contención máxima sobre estructuras compartidas confirma adicionalmente que las garantías de orden de memoria provistas por la librería estándar son suficientes para implementar primitivas de paralelismo correctas, evitando la introducción de las condiciones de carrera y bloqueos que típicamente afectan la implementación manual del paralelismo en este tipo de lenguajes.

Por su parte, el módulo Parser materializa el principio de "fallo temprano" propio de los frontends de compiladores robustos. Las cuatro categorías de validación semántica implementadas (compatibilidad del tipo de nodo objetivo, adyacencia comentario-bloque, contexto jerárquico de directivas anidadas y exigencia de listas no vacías para cláusulas de variables) aseguran que cualquier inconsistencia se detecte y reporte durante el análisis sintáctico, antes que el motor de transformación intente operar sobre estructuras inválidas y produzca errores de difícil diagnóstico. No obstante, el modelo de comentarios con valor semántico (\/\/gompher) hereda una fragilidad inherente al paradigma OpenMP: la dependencia estricta de la adyacencia física entre el comentario y su nodo objetivo, que puede ser interrumpida por líneas en blanco u otros comentarios intercalados. Esta limitación inherente constituye una tensión inevitable al utilizar metadatos textuales como portadores de semántica de ejecución, y refuerza la necesidad de que la documentación dirigida al usuario final enfatice con claridad las reglas de uso correcto de cada directiva.

El motor de transformación valida en la práctica la elección de una arquitectura _source-to-source_ sobre el AST nativo de Go. Al delegar tanto el análisis sintáctico (`go/parser`) como la resolución de tipos (`go/types`) a la librería estándar, el motor evita reimplementar componentes complejos del frontend del lenguaje y garantiza que el código emitido sea Go idiomático, portable y compilable por el _toolchain_ estándar sin modificación alguna. Las decisiones de implementación más delicadas se concentraron en las cláusulas de gestión de datos: la técnica de captura por puntero para `reduction` y `lastprivate` permite materializar la semántica de copia privada y de combinación sin renombrar las variables del cuerpo del usuario, preservando la legibilidad del código generado. Además, la normalización de la forma canónica de los bucles reproduce la estrategia de los compiladores de OpenMP, lo cual amplía las formas admitidas más allá del caso estrecho sin sacrificar la corrección. La limitación asumida es que el archivo generado constituye un artefacto de compilación intermedio, cuyos comentarios se descartan para evitar defectos conocidos de reposicionamiento de `go/printer`, de modo que la trazabilidad para el usuario recae sobre el código fuente anotado original y no sobre la salida transpilada.

Finalmente, la herramienta CLI integra los módulos anteriores en un único flujo coherente y materializa el objetivo central de GompherMP: transformar un archivo `.go` anotado en un binario nativo mediante una sola invocación. El principio de fallo temprano del frontend se extiende hasta este nivel, pues cada fase del flujo (lectura, análisis, validación, transformación, serialización y compilación) reporta su fallo con un mensaje específico y un código de salida distinto de cero. En cuanto a las pruebas de integración de extremo a extremo, constituyen la verificación más relevante del objetivo, pues confirman empíricamente que el código generado es aceptado sin modificaciones por el _toolchain_ del lenguaje anfitrión.
