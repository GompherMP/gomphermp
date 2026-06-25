
= Capítulo 4. OE1: Diseño y especificación de la arquitectura de la herramienta GompherMP

En este capítulo se detalla el trabajo realizado para cumplir con el primer objetivo específico, el cual comprende la definición formal de las directivas, la estructura modular del sistema y la interfaz de interacción con el usuario. Todos estos elementos han sido validados por el experto en programación paralela mencionado en capítulos anteriores, dichas validaciones se pueden encontrar en el Anexo F.


== R1: Documento de especificación de directivas y cláusulas

Como parte del cumplimiento del primer objetivo específico, se elaboró el documento de especificación técnica que define formalmente la gramática, sintaxis y semántica de las construcciones de paralelismo que soporta GompherMP. El propósito de este entregable es establecer el estándar para la adaptación de las directivas OpenMP al ecosistema nativo del lenguaje Go.

Para alcanzar este resultado, la selección de las directivas y cláusulas se realizó mediante un análisis técnico enfocado en cubrir los dos paradigmas principales del estándar OpenMP: el paralelismo estructurado y el paralelismo basado en tareas. En lugar de intentar abarcar la totalidad del estándar, se aplicó un criterio de delimitación de alcance (_scoping_) apropiado para la magnitud del proyecto. Se priorizaron los constructos fundamentales que resultan críticos para demostrar la viabilidad técnica de la traducción y paralelización de código hacia Go, garantizando que el subconjunto elegido sea altamente representativo para aplicaciones concurrentes de propósito general.

El documento de especificación aborda el diseño del lenguaje mediante 3 categorías:

- *Directivas de paralelismo estructurado:* Se detallan los constructos base (tales como `parallel`, `for` y `sections`), los cuales determinan la creación de regiones concurrentes y la distribución equitativa de cargas de trabajo iterativas o en bloques entre múltiples goroutines.

- *Directivas de paralelismo de tareas:* Abarcan los constructos orientados a cargas de trabajo irregulares, dinámicas o asíncronas (tales como `task`, `taskgroup` y `taskwait`). Estas directivas definen la generación independiente de tareas y la orquestación de su ejecución en función de dependencias explícitas.

- *Cláusulas de gestión de datos y sincronización:* Se establecen los mecanismos estandarizados para controlar el alcance, el aislamiento de memoria y la visibilidad de las variables dentro del entorno concurrente. Esto incluye la definición estricta de cláusulas como `shared`, `private`, `firstprivate` y `reduction`, garantizando la consistencia y evitando condiciones de carrera.

Para la revisión completa de la gramática y los ejemplos de uso de cada directiva, se recomienda revisar el Anexo C, donde se proporciona el enlace de acceso directo al documento.

A modo de ilustración, la Figura 1 muestra un extracto del documento de especificación, evidenciando el diseño de la sintaxis para el constructo de paralelismo iterativo.

#figure(
  image("../figures/fig4_1_directive_spec.png"),
  caption: [Extracto del documento de especificación detallando la directiva parallel],
) <fig:figura-4-1-extracto-del-docume>


== R2: Diseño de la arquitectura de la herramienta

Este entregable define la arquitectura de la herramienta GompherMP. El documento cumple tres funciones fundamentales: describir los módulos del software y sus interacciones, detallar el flujo general del programa, y explicar a profundidad cómo funcionan los mecanismos internos, específicamente los algoritmos para la transformación de código.

El diseño del sistema se estructura mediante cuatro módulos principales que interactúan secuencialmente a partir de un punto de entrada central (main):

- *Módulo CLI:* Encargado de mapear y procesar los argumentos recibidos por la línea de comandos.

- *Módulo Parser:* Transforma el código fuente original, enriquecido con directivas de GompherMP, en una representación intermedia.

- *Módulo Transformer:* Constituye el núcleo de la herramienta; inspecciona la representación intermedia, valida su sintaxis GompherMP y ejecuta las transformaciones algorítmicas correspondientes a cada directiva o cláusula.

- *Módulo Printer:* Toma la representación intermedia transformada y genera el código fuente final equivalente en el lenguaje Go estándar.

A modo de ilustración, la Figura 2 presenta el diagrama de arquitectura diseñado para la herramienta, detallando el caso de uso principal y las interacciones entre los módulos durante la conversión de un archivo con sintaxis enriquecida a uno estándar.

#figure(
  image("../figures/fig4_2_modules_diagram.png"),
  caption: [Diagrama de módulos e interacciones],
) <fig:figura-4-2-diagrama-de-modulos>

El diseño de esta arquitectura se alcanzó tomando como referencia el _pipeline_ de fases clásico utilizado en el desarrollo de compiladores convencionales, adaptándolo a las necesidades específicas de un compilador _source-to-source_. De este modo, se aplicó un mapeo directo de responsabilidades: la función del programa conductor (driver) fue asignada al main; las fases de tokenización y análisis sintáctico (_tokenizer_/parser) se encapsularon en el módulo Parser para la generación del Árbol Sintáctico Abstracto (AST); la etapa de análisis semántico y reestructuración se consolidó en el módulo Transformer; y, finalmente, la fase de generación de código (codegen) fue delegada al módulo Printer. Este enfoque basado en patrones arquitectónicos clásicos de procesamiento de lenguajes garantiza alta cohesión, bajo acoplamiento y facilita la extensibilidad de la herramienta.

Respecto al flujo general del programa, el proceso central recae en el módulo transformador, el cual opera mediante una inspección exhaustiva y secuencial de cada nodo de la representación intermedia. Al detectar un nodo asociado a la sintaxis enriquecida de GompherMP, el sistema evalúa su validez contextual; si la directiva es válida, se aplica el algoritmo de reemplazo para sustituir el nodo original por su versión transformada, mientras que, en caso de detectar inconsistencias, se interrumpe la ejecución de forma segura para emitir el error correspondiente.

En cuanto a los mecanismos internos de transformación de código, el proceso algorítmico se basa fundamentalmente en abstraer el cuerpo de cada constructo paralelo dentro de una nueva función autogenerada con un identificador único e irrepetible. Posteriormente, el sistema inserta una invocación a una función de runtime personalizada que recibe dicha función encapsulada como argumento; son precisamente estas funciones de runtime las que asumen el trabajo pesado, encargándose de distribuir y orquestar dinámicamente las iteraciones o tareas entre múltiples hilos para garantizar su ejecución paralela.

Para sostener estas ejecuciones, la arquitectura interna del runtime se aleja de un modelo de creación dinámica convencional que introduciría un alto _overhead_. En su lugar, el sistema se apoya en dos subsistemas principales: un Gestor de Pool de Goroutines con hilos persistentes para manejar directivas estructuradas de forma óptima; y un Planificador de Tareas asíncrono que da soporte al paralelismo no estructurado. Este planificador construye un Grafo Acíclico Dirigido (DAG) en tiempo de ejecución para resolver y garantizar de forma segura las dependencias antes de la ejecución de cada tarea.

A modo de ilustración, la Figura 3 presenta un diagrama del funcionamiento del gestor de _pool_ de Goroutines, y la Figura 4 presenta un ejemplo de un grafo acíclico dirigido a ser coordinado por el planificador de tareas.

#figure(
  image("../figures/fig4_3_goroutine_pool.png"),
  caption: [Funcionamiento del gestor de pool de goroutines],
) <fig:figura-4-3-funcionamiento-del>

#figure(
  image("../figures/fig4_4_task_dag.png"),
  caption: [Ejemplo de un grafo acíclico dirigido de tareas],
) <fig:figura-4-4-ejemplo-de-un-grafo>

En síntesis, el diseño arquitectónico actúa como un puente semántico para conciliar dos paradigmas de concurrencia dispares: el modelo tradicional de memoria compartida asumido por OpenMP y el modelo de Procesos Secuenciales Comunicantes (CSP) idiomático de Go. Esta fricción se resuelve combinando un análisis estático durante la fase de transformación, para determinar el alcance de las variables, con la inyección de mecanismos de sincronización en tiempo de ejecución (mediante canales y mutexes), lo que previene condiciones de carrera y permite al desarrollador programar bajo un modelo mental puramente secuencial.

Se recomienda revisar el Anexo D para ver la explicación detallada de módulos, interacciones, flujo del programa y los mecanismos de transformación de código.


== R3: Especificación funcional de la Interfaz de Línea de Comandos (CLI)

Este entregable define la interfaz de interacción entre el desarrollador y la herramienta. El diseño de la CLI de GompherMP prioriza la usabilidad y la baja fricción, emulando el comportamiento y la sintaxis de las herramientas estándar del ecosistema de Go como `go build`.

Las funciones clave especificadas para la interfaz incluyen:

- *Gestión de Compilación:* La capacidad de invocar el _pipeline_ completo mediante el comando `gompher build`.

- *Configuración mediante Banderas (Flags):* Implementación de opciones para personalizar la salida (-o), activar el modo detallado o verbose (-v) para depuración del AST, y conservar archivos intermedios (-k) para análisis técnico.

- *Automatización del Pipeline:* La CLI orquesta automáticamente la validación de prerrequisitos, la invocación de los módulos de transformación y la limpieza de artefactos temporales tras una compilación exitosa.

La especificación detallada de los comandos, códigos de error y casos de uso se encuentra documentada en el Anexo E.

Como se observa en la Figura 5, la interfaz proporciona una retroalimentación clara al usuario sobre el estado de la transformación y la detección de directivas en el código fuente.

#figure(
  image("../figures/fig4_cli_flags.png"),
  caption: [Ejemplos del uso de banderas],
) <fig:figura-4-5-ejemplos-del-uso-de>

Para definir la especificación funcional de la CLI, se llevó a cabo un proceso de revisión y homologación basado en las convenciones estándar de usabilidad de interfaces de línea de comandos modernas. Se analizaron los esquemas de ayuda y gestión de parámetros (flags) de herramientas de compilación e intérpretes ampliamente utilizados en la industria. A partir de este análisis, se extrajeron y adaptaron los comandos correspondientes y necesarios para el flujo de trabajo específico de GompherMP, asegurando con ello que la herramienta resultante ofrezca una interfaz intuitiva y una curva de aprendizaje mínima para los desarrolladores habituados a entornos de terminal.


== Discusión de resultados

El diseño arquitectónico de GompherMP revela una notable compatibilidad técnica entre el modelo de OpenMP (C/C++) y los paradigmas de Go. La técnica clásica de _outlining_ de funciones, pilar de las implementaciones de OpenMP en C, se traslada de forma natural a Go mediante el uso de funciones de orden superior y _closures_. Esta similitud permite que el módulo Transformer adopte patrones lógicos de compiladores estándar, demostrando que el modelo de transformación de OpenMP es lo suficientemente agnóstico para ser implementado eficientemente sobre el modelo de concurrencia de Go, mitigando la complejidad de traducir entre ecosistemas tan distintos como _pthreads_ y goroutines.

No obstante, la sintaxis adoptada para estas directivas plantea una ruptura con la práctica convencional de programación en Go. El uso de comentarios con valor semántico (\/\/gompher) puede resultar extraño para el desarrollador habitual, quien bajo condiciones normales asume que los comentarios no afectan el comportamiento del binario. Esta propuesta exige que el usuario ceda el control del flujo de ejecución de sus procesos a una capa de abstracción externa (nuestro runtime), permitiendo que la herramienta reestructure bloques de código ocultando la complejidad interna. Aunque este enfoque facilita la paralelización masiva, introduce una carga cognitiva distinta: el programador deja de gestionar directamente las primitivas de concurrencia nativas para confiar en la interpretación de metadatos insertados en el código fuente.
