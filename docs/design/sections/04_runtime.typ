= Arquitectura Interna del Runtime

Si bien el ejemplo de la sección anterior ilustra el mecanismo básico de delegación de trabajo mediante una implementación "naive", la arquitectura real del runtime de GompherMP no se basa en un modelo de creación dinámica de goroutines por cada bucle. Dicho enfoque de "lanzar y olvidar" introduciría un alto overhead en operaciones de cómputo intensivo. Para solucionar esto, el runtime se estructura en tres subsistemas internos diseñados para maximizar el rendimiento y el control sobre la CPU:

== Gestor de Pool de Goroutines

Para gestionar directivas estructuradas como parallel, el runtime implementa un patrón de Worker Pool. Al iniciar la ejecución del programa, el sistema pre-instancia un equipo de goroutines persistentes (cuyo tamaño se ajusta dinámicamente o mediante la variable de entorno GOMAXPROCS). Las funciones extraídas por el módulo transformador se envían a este pool a través de canales no bloqueantes, eliminando por completo el costo de creación y destrucción de hilos nativos en regiones paralelas que se ejecutan de manera recurrente.

A continuación se muestra un diagrama explicativo sobre el funcionamiento de este threadpool:

#figure(
  image("../assets/threadpool.png", width: 100%),
  caption: [
    Threadpool
  ],
)

== Planificador de Tareas

Para dar soporte al paralelismo no estructurado (mediante directivas como `task`, `taskwait` y `taskgroup`), el runtime incluye un motor de resolución de dependencias. Cuando el desarrollador utiliza la cláusula `depend(in, out, inout)`, este motor garantiza que las tareas respeten las relaciones de orden sobre los datos compartidos, modeladas conceptualmente como un grafo de dependencias acíclico dirigido (DAG): cada tarea es un nodo y cada relación de precedencia entre accesos a memoria es una arista.

Sin embargo, la implementación no construye dicho grafo de forma explícita. En su lugar, se emplea un esquema de *seguimiento por frontera* (_frontier-based tracking_): por cada token de variable (la dirección de memoria utilizada como clave de correlación), el motor mantiene únicamente dos piezas de información: la señal de completitud de la última tarea escritora y el conjunto de señales de las tareas lectoras activas en ese momento. Esta frontera es el mínimo de información necesaria para tomar una decisión de despacho correcta, sin necesidad de almacenar el historial completo de dependencias.

Cuando se registra una nueva tarea con cláusulas `depend`, el motor realiza atómicamente las siguientes dos acciones:

+ *Recopila las señales de espera:* según el rol de la tarea para cada token (`in`, `out` o `inout`), determina qué tareas predecesoras deben haber concluido antes de que esta pueda ejecutar su cuerpo.
+ *Actualiza la frontera:* registra la señal de completitud de la nueva tarea como la nueva referencia de escritora o lectora activa para ese token.

La espera efectiva sobre estas señales ocurre dentro de la propia goroutine de la tarea, inmediatamente antes de ejecutar su cuerpo. Esto garantiza el orden _happens-before_ correcto para todas las combinaciones de dependencias sin requerir un grafo en memoria.

A continuación se muestra un ejemplo de las relaciones de orden entre tareas que este motor garantiza, representadas como el DAG conceptual que el esquema de frontera reproduce implícitamente:

#figure(
  image("../assets/tasks_dag.png", width: 100%),
  caption: [
    Grafo acíclico dirigido (DAG) de dependencias entre tareas
  ],
)