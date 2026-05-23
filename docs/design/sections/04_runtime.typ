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

Para dar soporte al paralelismo no estructurado (mediante directivas como task y taskwait), el runtime incluye un planificador asíncrono especializado. Este componente mantiene una cola de tareas concurrente y un motor de resolución de dependencias. Cuando el desarrollador utiliza la cláusula depend(in, out), el planificador construye un grafo de dependencias acíclico dirigido (DAG) en tiempo de ejecución. Este motor asegura que una tarea no sea extraída de la cola ni asignada a una goroutine ociosa del pool hasta que todos sus pre-requisitos de memoria hayan sido satisfechos.

A continuación se muestra un ejemplo de un grafo acíclico dirigido (DAG) que representa las dependencias entre tareas para un proceso de transformación y agregación de datos genérico:

#figure(
  image("../assets/tasks_dag.png", width: 100%),
  caption: [
    Grafo acíclico dirigido (DAG) de tareas
  ],
)