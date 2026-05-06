
= Capítulo 2. Marco Referencial


== 2.1. Marco Teórico


=== 2.1.1. Computación Paralela


==== 2.1.1.1. Taxonomía de Flynn

De acuerdo con Hennessy y Patterson (2012), una de las clasificaciones más fundamentales y duraderas para las arquitecturas de computadores paralelos fue la propuesta por Michael Flynn en la década de 1960. Este modelo, que sigue siendo una referencia en la actualidad, categoriza los computadores en función de los flujos de instrucciones y de datos que pueden procesar. La taxonomía se divide en cuatro categorías principales:

Single instruction stream, single data stream (SISD): Esta categoría, según Hennessy y Patterson (2012), representa al uniprocesador estándar, donde un único flujo de instrucciones opera sobre un único flujo de datos.

Single instruction stream, multiple data streams (SIMD): En este modelo, una única instrucción es aplicada sobre múltiples flujos de datos de forma simultánea, lo que permite explotar el paralelismo de datos (data-level parallelism). Hennessy y Patterson (2012) señalan que ejemplos de esta arquitectura son las arquitecturas vectoriales y las Unidades de Procesamiento Gráfico (GPUs).

Multiple instruction streams, single data stream (MISD): Aunque lógicamente completa la taxonomía, los autores indican que hasta la fecha no se han construido computadores comerciales que se ajusten a esta categoría (Hennessy & Patterson, 2012).

Multiple instruction streams, multiple data streams (MIMD): Hennessy y Patterson (2012) la describen como la arquitectura más flexible y general, en la que cada procesador ejecuta su propio flujo de instrucciones sobre sus propios datos. Este modelo está orientado de forma natural a explotar el paralelismo de tareas (task-level parallelism) y es la base de la mayoría de los sistemas multinúcleo y clústeres actuales.


==== 2.1.1.2. Concurrencia y Paralelismo

Aunque en la práctica los términos concurrencia y paralelismo a menudo se utilizan de manera intercambiable, es fundamental establecer una distinción conceptual para el propósito de este trabajo. Según Pacheco y Malensek (2022), aunque no existe un acuerdo completo en la comunidad sobre la definición exacta de estos términos, es posible trazar diferencias clave.

La concurrencia se refiere a un programa en el que múltiples tareas pueden estar en progreso en cualquier instante. Esto no implica necesariamente que se estén ejecutando de forma simultánea. Un ejemplo clásico es un sistema operativo multitarea en un computador con un solo núcleo: las tareas avanzan de forma intercalada, dando la ilusión de simultaneidad, pero solo una se ejecuta en un momento dado (Pacheco & Malensek, 2022).

Por otro lado, el paralelismo, de acuerdo con Pacheco y Malensek (2022), describe un programa en el que múltiples tareas cooperan estrechamente para resolver un problema, usualmente ejecutándose de manera simultánea en hardware dedicado para ello. La diferencia fundamental reside en que el paralelismo busca la ejecución simultánea real para acelerar la resolución de un único problema.

De esta manera, todo programa paralelo es inherentemente concurrente, pero no todo programa concurrente es paralelo. Pacheco y Malensek (2022) sugieren que una distinción útil radica en el nivel de acoplamiento: las tareas en un programa paralelo están "estrechamente acopladas", ya que se ejecutan en núcleos que comparten memoria o están conectados por una red de muy alta velocidad.


==== 2.1.1.3. Modelos de Computación Paralela

La forma en que los programas paralelos gestionan el acceso a la memoria y coordinan sus diferentes hilos de ejecución define su modelo de computación. De acuerdo con Pacheco y Malensek (2022), estos modelos determinan cómo se comunican y sincronizan las tareas para resolver un problema de forma cooperativa.

Sistemas de Memoria Compartida: Según Pacheco y Malensek (2022), en los programas de memoria compartida, las variables pueden ser compartidas o privadas. Las variables compartidas pueden ser leídas o escritas por cualquier hilo de ejecución, mientras que las variables privadas normalmente solo pueden ser accedidas por un único hilo. En este paradigma, la comunicación entre los hilos se realiza de forma implícita a través de la modificación de estas variables compartidas.

Sistemas de Memoria Distribuida: En contraste, en los programas de memoria distribuida, los núcleos solo pueden acceder directamente a su propia memoria privada. De acuerdo con Pacheco y Malensek (2022), la forma de comunicación más utilizada en este modelo es el paso de mensajes (message-passing). Es importante destacar que las APIs de memoria distribuida también pueden ser utilizadas en hardware de memoria compartida; para ello, un compilador o una librería se encarga de particionar lógicamente la memoria en espacios de direcciones privados y de implementar la comunicación necesaria. Típicamente, los programas de memoria distribuida se inician como múltiples procesos en lugar de múltiples hilos, ya que los "hilos de ejecución" pueden correr en CPUs con sistemas operativos independientes.

Sistemas Híbridos: Existe también la posibilidad de programar sistemas híbridos, como los clústers de procesadores multinúcleo. Pacheco y Malensek (2022) explican que esto se puede lograr utilizando una combinación de una API de memoria compartida dentro de cada nodo y una API de memoria distribuida para la comunicación entre los nodos. Sin embargo, este enfoque se reserva generalmente para programas que requieren los más altos niveles de rendimiento, ya que la complejidad de utilizar una API híbrida dificulta considerablemente el desarrollo del programa.


==== 2.1.1.4. Modelos de Comunicación y Sincronización

Cuando múltiples procesos o hilos se ejecutan de forma paralela, es indispensable contar con mecanismos que les permitan comunicarse y coordinar sus acciones. Según Pacheco y Malensek (2022), la necesidad de sincronizar los hilos es uno de los problemas fundamentales que deben resolverse para crear un programa paralelo correcto. Existen dos paradigmas principales para lograr esta coordinación: uno basado en el intercambio explícito de información y otro en el acceso controlado a recursos comunes.

Paso de Mensajes y CSP:

El modelo de paso de mensajes se basa en la idea de que los procesos, que no comparten memoria, se comunican enviándose datos explícitamente entre ellos. Uno de los modelos teóricos más influyentes en este paradigma es el de Procesos Secuenciales Comunicantes (Communicating Sequential Processes o CSP), propuesto por C.A.R. Hoare.

Según Hoare (1978), en el modelo CSP, la entrada y la salida son consideradas primitivas básicas de la programación. La comunicación ocurre cuando un proceso nombra a otro como destino para una salida (destination!expression) y el segundo nombra al primero como fuente para una entrada (source?target_variable). Un aspecto fundamental de este modelo es que la comunicación es sincrónica y sin búferes intermedios (unbuffered). Hoare (1978) especifica que un comando de entrada o salida se retrasa hasta que el proceso correspondiente esté listo para realizar la operación complementaria. En ese momento, los comandos se ejecutan simultáneamente y el valor se transfiere. Esta sincronización inherente en el acto de comunicación es una característica central del modelo CSP.

Variables Compartidas y Sincronización:

En el paradigma de memoria compartida, los procesos se comunican de forma implícita al leer y escribir en una misma área de memoria. Esta flexibilidad introduce un desafío significativo conocido como el problema de la sección crítica. Según Silberschatz, Galvin y Gagne (2009), cada proceso tiene un segmento de código, llamado sección crítica, en el que accede a recursos compartidos. Para garantizar la consistencia de los datos, es crucial que cuando un proceso esté ejecutando su sección crítica, ningún otro proceso pueda ejecutar la suya. Una solución a este problema debe satisfacer tres requisitos: exclusión mutua, progreso y espera limitada.

Para resolver este problema, se han desarrollado diversas herramientas de sincronización. Una de las más conocidas es el semáforo. Silberschatz et al. (2009) lo definen como una variable entera a la que solo se puede acceder a través de dos operaciones atómicas estándar: wait() (originalmente P) y signal() (originalmente V). A pesar de su utilidad, los semáforos deben usarse con extremo cuidado, ya que un uso incorrecto, como intercambiar el orden de wait() y signal(), puede violar la exclusión mutua o, peor aún, provocar un interbloqueo (deadlock) (Silberschatz et al., 2009).

Para simplificar la tarea del programador y evitar los errores comunes asociados a los semáforos, se han desarrollado abstracciones de más alto nivel como los monitores. Silberschatz et al. (2009) introducen los monitores como una construcción que busca ofrecer un mecanismo conveniente y efectivo para la sincronización, mitigando los errores de temporización difíciles de detectar que pueden surgir con el uso de semáforos.


==== 2.1.1.5. Leyes y Límites Teóricos del Rendimiento Paralelo

Para comprender y predecir la ganancia de rendimiento potencial al paralelizar un programa, existen modelos teóricos fundamentales que establecen los límites de la aceleración (speedup) que se puede alcanzar.

Ley de Amdahl:

Según Hennessy y Patterson (2012), la Ley de Amdahl es una fórmula que permite calcular la mejora de rendimiento que se puede obtener al optimizar una porción de un sistema computacional. Su principio fundamental establece que la ganancia de rendimiento global está limitada por la fracción de tiempo que la porción mejorada puede ser utilizada.

El speedup o aceleración se define como la relación entre el tiempo de ejecución sin la mejora y el tiempo de ejecución con la mejora (Hennessy & Patterson, 2012).

Para aplicar la ley, se deben considerar dos factores clave:

Fracción mejorada (Fraction enhanced): La fracción del tiempo de cómputo original que puede aprovechar la mejora. Por ejemplo, si una sección paralelizable de un programa tarda 20 segundos de un total de 60, esta fracción es 20/60.

Aceleración de la fracción mejorada (Speedup enhanced): Cuánto más rápida es la porción mejorada. Si un cálculo tardaba 5 segundos y ahora tarda 2, el speedup de esa porción es 5/2.

Con base en esto, Hennessy y Patterson (2012) establecen la fórmula para la aceleración total (Speedup overall) de la siguiente manera:

Esta ley demuestra que, sin importar cuán rápida se haga la porción paralela, la aceleración total siempre estará limitada por la porción del código que debe ejecutarse de forma secuencial.

Ley de Gustafson:

La Ley de Amdahl puede presentar un panorama desalentador para el paralelismo. Sin embargo, Pacheco y Malensek (2022) señalan que una de las principales críticas a esta ley es que no toma en consideración el tamaño del problema. Argumentan que, para muchos problemas del mundo real, a medida que el tamaño del problema aumenta, la fracción del programa que es inherentemente secuencial tiende a disminuir.


=== 2.1.2. Computación de Alto Rendimiento (HPC)


==== 2.1.2.1. Definición y Objetivos de HPC

La evolución de la ciencia y la ingeniería ha estado intrínsecamente ligada a la capacidad de procesamiento computacional. Problemas de gran envergadura, como la creación de modelos climáticos precisos, el estudio del plegamiento de proteínas para comprender enfermedades degenerativas, o el análisis masivo de datos genómicos, demandan un poder de cómputo que excede las capacidades de los sistemas convencionales (Pacheco & Malensek, 2022). Durante décadas, la industria del hardware respondió a esta necesidad incrementando la velocidad de los procesadores de un solo núcleo, una estrategia impulsada por el aumento en la densidad de transistores.

Sin embargo, como señalan Pacheco y Malensek (2022), a principios del siglo XXI este modelo de escalamiento vertical alcanzó un límite físico fundamental. El incremento en la velocidad de los transistores generaba un aumento insostenible en el consumo de energía y la disipación de calor, haciendo inviable seguir construyendo procesadores monolíticos cada vez más rápidos. La respuesta de la industria a este desafío fue un cambio de paradigma hacia el paralelismo: en lugar de un único procesador complejo, se comenzaron a integrar múltiples procesadores más simples, o núcleos, en un solo chip.

Este enfoque es la base de la Computación de Alto Rendimiento (HPC). Según IBM (2025), la HPC puede definirse como una tecnología que utiliza clusters de procesadores potentes que trabajan en paralelo para procesar conjuntos de datos masivos y resolver problemas complejos a velocidades extremadamente altas. El objetivo fundamental de la HPC, por lo tanto, es agregar la potencia de cómputo de múltiples unidades de procesamiento para abordar problemas que, por su escala o complejidad, serían intratables para los sistemas de computación tradicionales, impulsando así la investigación y la innovación en dominios críticos.


== 2.2. Marco Conceptual


=== 2.2.1. Concurrencia pragmática en Go: goroutines, canales y sincronización


==== 2.2.1.1. Descripción General

Go adopta un enfoque inspirado en CSP, donde la composición de tareas y la seguridad de acceso a datos se apoyan en goroutines y canales como mecanismos primarios. La práctica recomendada es coordinar el intercambio de datos mediante comunicación explícita, en lugar de compartir memoria directamente (The Go Programming Language, s. f., Effective Go; The Go Programming Language, s. f., FAQ).


==== 2.2.1.2. Componentes y modelo de ejecución

Goroutines: Son la unidad fundamental de concurrencia en Go. A diferencia de los hilos del sistema operativo, las goroutines son extremadamente ligeras y están gestionadas por el runtime; sus stacks comienzan en pocos kilobytes y crecen o se reducen dinámicamente, lo que permite lanzar cientos de miles en un mismo proceso (The Go Programming Language, s. f., Effective Go).

Scheduler y paralelismo: El runtime multiplexa goroutines sobre hilos del sistema operativo y permite ajustar el grado de paralelismo con GOMAXPROCS, de modo que varias goroutines puedan ejecutarse realmente en paralelo en múltiples CPU lógicas cuando corresponde (The Go Programming Language, s. f., Effective Go).


==== 2.2.1.3. Comunicación y Sincronización

Canales (channels): Son conductos tipados para enviar/recibir valores entre goroutines y materializan la idea “no te comuniques compartiendo memoria; comparte memoria comunicándote” (The Go Programming Language, s. f., Effective Go). Además, el modelo de memoria de Go establece relaciones happens-before para las operaciones de canal, posibilitando diseños que eviten data races cuando se transfiere la propiedad de los datos entre goroutines (The Go Programming Language, 2022, The Go Memory Model).

Primitivas tradicionales en sync: Cuando la coordinación por canales no encaja (p. ej., estado compartido que debe protegerse o secciones críticas muy cortas y frecuentes), Go ofrece primitivas de memoria compartida en sync (The Go Programming Language, s. f., Effective Go; The Go Programming Language, s. f., pkg sync).
Mutex provee exclusión mutua mediante Lock/Unlock. Semánticamente, el desbloqueo de un Mutex sucede-antes del bloqueo posterior del mismo Mutex, garantizando la visibilidad de las escrituras hechas dentro de la sección crítica (The Go Programming Language, 2022, The Go Memory Model; The Go Programming Language, s. f., pkg sync).
 WaitGroup permite esperar a un conjunto de goroutines con el patrón Add/Done/Wait. Un Done() que permite que Wait() retorne sucede-antes del retorno de ese Wait(), por lo que los efectos de las tareas finalizadas son observables después de la espera (The Go Programming Language, 2022, The Go Memory Model; The Go Programming Language, s. f., pkg sync).
Otras utilidades (RWMutex, Cond, Once, Map, Pool y sync/atomic) cubren casos específicos, pero para los objetivos de esta tesis Mutex y WaitGroup resultan las más relevantes para contrastar con el modelo de canales (The Go Programming Language, s. f., Effective Go; The Go Programming Language, s. f., pkg sync).

2.2.1.4. Implicancias para el rendimiento y diseño

El modelo de Go favorece servidores concurrentes e I/O intensivo gracias a la ligereza de las goroutines y a la composición con canales. En cargas marcadamente CPU-bound, la elección entre canales y sync depende del patrón de acceso a datos y del costo de sincronización; en estructuras altamente compartidas, un Mutex bien ubicado puede reducir sobrecarga frente a protocolos de mensajería más elaborados, mientras que en pipelines de procesamiento los canales simplifican el diseño (The Go Programming Language, s. f., Effective Go; The Go Programming Language, 2022, The Go Memory Model).


=== 2.2.2. Paralelismo de Memoria Compartida Dirigido por Directivas: El Estándar OpenMP


==== 2.2.2.1. Definición del Estándar

OpenMP (Open Multi-Processing) es una Interfaz de Programación de Aplicaciones (API) diseñada específicamente para la programación en sistemas de memoria compartida con arquitecturas MIMD. Según Pacheco y Malensek (2022), fue desarrollada por un consorcio de programadores y científicos de la computación que consideraban que escribir programas de alto rendimiento con APIs de bajo nivel como Pthreads era excesivamente difícil. El objetivo era crear un estándar que permitiera desarrollar programas paralelos a un nivel de abstracción más alto.

De acuerdo con Dagum y Menon (1998), OpenMP surgió como una alternativa portable a los modelos de paso de mensajes. Antes de su creación, los desarrolladores se veían forzados a elegir entre utilizar extensiones propietarias de un fabricante de hardware, perdiendo portabilidad, o adoptar modelos más complejos como MPI para asegurar que su código pudiera ejecutarse en diferentes plataformas.

Una de las ventajas más significativas de OpenMP, destacada tanto por Pacheco y Malensek (2022) como por Dagum y Menon (1998), es que permite la paralelización incremental. A diferencia del paso de mensajes, donde todo el programa y sus estructuras de datos deben ser descompuestos para funcionar en paralelo, OpenMP permite a los desarrolladores comenzar por paralelizar pequeñas secciones de un código secuencial existente, como bucles individuales, para obtener mejoras de rendimiento inmediatas con un esfuerzo relativamente bajo.


==== 2.2.2.2. Componentes y Modelo de Ejecución

En su nivel más fundamental, OpenMP está compuesto por tres elementos principales. Según Dagum y Menon (1998), estos son:

Un conjunto de directivas de compilador (conocidas como pragmas en C/C++ o comentarios especiales en Fortran) que instruyen al compilador sobre cómo paralelizar una sección de código.

Una librería de rutinas de tiempo de ejecución (runtime) que gestiona los detalles de la ejecución, como la creación y sincronización de hilos.

Variables de entorno que permiten al usuario controlar aspectos de la ejecución del programa paralelo.

El modelo de ejecución que utiliza OpenMP es conocido como fork/join (bifurcación y unión). Dagum y Menon (1998) describen este proceso de la siguiente manera: un programa OpenMP comienza su ejecución como un único hilo de control, conocido como el hilo maestro. Cuando el hilo maestro encuentra una construcción paralela (la fase de fork), crea un equipo de hilos adicionales. La región de código paralela es entonces ejecutada por todos los hilos del equipo de forma simultánea. Al finalizar la región paralela, los hilos se sincronizan en una barrera implícita y se terminan, dejando únicamente al hilo maestro para continuar con la ejecución del resto del programa (la fase de join). Este ciclo de fork/join puede repetirse múltiples veces a lo largo de la ejecución del programa.

Este modelo, como señalan Dagum y Menon (1998), es lo que facilita la paralelización incremental, ya que el código secuencial original no necesita ser reescrito por completo; solo se marcan con directivas las secciones que se beneficiarán de la ejecución paralela.

OpenMP representa el modelo mental del desarrollador de HPC y el estándar de facto para el paralelismo en sistemas de memoria compartida.

Historia y Propósito: OpenMP (Open Multi-Processing) fue creado a finales de los años 90 por un consorcio de empresas de hardware y software para estandarizar y simplificar la programación paralela en sistemas de memoria compartida. Su propósito es ofrecer una API portable y escalable que permita a los desarrolladores paralelizar código en C, C++ y Fortran de forma incremental y con un esfuerzo mínimo.

Modelo de Ejecución Fork-Join: El modelo de ejecución de OpenMP es el fork-join. Un programa comienza con un único hilo de ejecución (el hilo maestro). Cuando el hilo maestro encuentra una construcción paralela, crea (forks) un equipo de hilos esclavos. La región de código es ejecutada en paralelo por todos los hilos. Al finalizar la región paralela, los hilos esclavos se sincronizan y terminan, y solo el hilo maestro continúa (join). Este modelo está diseñado explícitamente para el paralelismo de bucles y tareas en un espacio de memoria compartido.

Componentes Fundamentales de OpenMP:

Los componentes fundamentales de OpenMP operan en conjunto para ofrecer un modelo de paralelismo declarativo de alto nivel, diseñado para abstraer la complejidad inherente a la programación de hilos manual. El pilar de este sistema son las Directivas (Pragmas), que son instrucciones especiales insertadas en el código fuente, como \#pragma omp parallel. Estas directivas no son parte del lenguaje base, sino que actúan como anotaciones que le indican al compilador qué secciones del algoritmo deben ser ejecutadas en paralelo. Su principal ventaja es que permiten al programador expresar la intención de paralelizar sin necesidad de reescribir la lógica fundamental del programa, manteniendo el código limpio y cercano a su versión secuencial. En el contexto de esta tesis, el enfoque se centra en directivas clave como parallel, que crea un equipo de hilos para ejecutar un bloque de código, y for, una directiva de reparto de trabajo (work-sharing) que distribuye automáticamente las iteraciones de un bucle entre los hilos disponibles.

Para refinar y controlar el comportamiento de estas directivas, se utilizan las Cláusulas. Estas actúan como modificadores que se añaden a los pragmas para gestionar el entorno de datos y la ejecución. Por ejemplo, las cláusulas de visibilidad de datos como private y shared son esenciales para evitar condiciones de carrera, permitiendo al programador especificar si una variable debe tener una copia local para cada hilo o si debe ser compartida entre todos. Cláusulas más avanzadas como reduction simplifican operaciones de agregación comunes (sumas, productos, etc.), automatizando la creación de variables privadas para los resultados parciales y su posterior combinación segura en un resultado final. De este modo, las cláusulas ofrecen un control fino y declarativo sobre aspectos complejos del paralelismo sin contaminar la lógica del algoritmo.

Finalmente, la Librería de Runtime (Runtime Library) complementa el modelo estático de las directivas con un conjunto de funciones que otorgan un control más dinámico y granular sobre el entorno de ejecución paralelo. Funciones como omp_get_thread_num() permiten a un hilo identificar su propio índice dentro del equipo, mientras que omp_get_num_threads() devuelve el número total de hilos activos. Otras, como omp_set_num_threads(), permiten al programa ajustar el número de hilos que se usarán en las siguientes regiones paralelas, facilitando la creación de algoritmos que se adaptan a las condiciones del sistema en tiempo de ejecución. En conjunto, estos tres componentes (directivas, cláusulas y librería de runtime) conforman un ecosistema robusto que ha consolidado a OpenMP como el estándar para el paralelismo de memoria compartida.


=== 2.2.2. Fundamentos Conceptuales de la Problemática


==== 2.2.2.1. Fricción Semántico-Sintáctica

Este es el concepto central que formaliza el problema. Se define como el choque cognitivo y técnico que ocurre cuando se intenta implementar un patrón del paradigma declarativo (OpenMP) usando las primitivas del paradigma concurrente idiomático (Go). El resultado es un código más verboso y difícil de mantener, donde la lógica del algoritmo queda fuertemente acoplada con los mecanismos de concurrencia de bajo nivel (Madridejos Zamorano, 2015).


==== 2.2.2.2. Carga Cognitiva

Es la cantidad de esfuerzo mental que un desarrollador necesita para escribir, entender y razonar sobre un algoritmo. La gestión manual del paralelismo, consecuencia de la fricción, incrementa la carga cognitiva al forzar al programador a manejar detalles complejos de sincronización, lo que aumenta la probabilidad de errores algorítmicos como condiciones de carrera o interbloqueos (Powers & Alaghband, 2007).


==== 2.2.2.3. Baja Expresividad y Productividad

La expresividad se refiere a la capacidad de un paradigma para formular ideas computacionales complejas de forma clara y concisa. La implementación manual del paralelismo en Go a menudo requiere más líneas de código (LOC) que una solución declarativa para lograr el mismo objetivo (Vikas et al., 2014). Esta verbosidad es un síntoma de baja expresividad y, consecuentemente, de una menor productividad, ya que dificulta la rápida experimentación y el mantenimiento del código.
