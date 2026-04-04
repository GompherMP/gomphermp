= Flujo de Ejecución (Pipeline)

El CLI de GompherMP opera como un transpilador (source-to-source) que orquesta la transformación del código fuente antes de la compilación final. Basado en el diseño arquitectónico de la herramienta, el pipeline se divide en las siguientes etapas consecutivas:

== Validación y Mapeo (Módulo CLI)
El punto de entrada recibe los argumentos del usuario y valida la integridad de los archivos fuente (`.go`). En esta etapa se comprueban las rutas de acceso y se preparan los flags de configuración (como el modo *verbose* o la ruta de *output*) que afectarán el comportamiento de los módulos posteriores.

== Análisis Sintáctico (Módulo Parser)
El código fuente es procesado por el *Parser*, el cual genera una representación intermedia basada en un Árbol Sintáctico Abstracto (AST). Esta fase es crucial para identificar la estructura del programa y localizar los nodos que contienen comentarios de directivas de GompherMP.

== Transformación Semántica (Módulo Transformer)
El *Transformer* inspecciona el AST en busca de directivas y cláusulas. Al detectar un constructo (como `parallel for`), realiza las siguientes acciones:
- Valida la sintaxis específica de la directiva y sus cláusulas asociadas.
- Extrae el bloque de código y lo encapsula en una *k-función* con un identificador único (hash) para evitar colisiones de nombres.
- Inyecta las llamadas al Runtime de GompherMP, pasando como argumento la función encapsulada y los parámetros de gestión de datos (`private`, `shared`, `reduction`).

== Generación de Código (Módulo Printer)
Una vez transformado el AST, el *Printer* convierte la representación intermedia de nuevo en código fuente Go legible. Este módulo emite un archivo temporal que incluye tanto el código original del usuario como las nuevas estructuras de concurrencia inyectadas.

== Compilación Nativa y Enlazado
El CLI invoca internamente el comando nativo `go build`. En esta etapa, el compilador de Go enlaza el código generado por el *Printer* con la librería de soporte del Runtime de GompherMP, produciendo el binario ejecutable final.

== Finalización y Limpieza
Tras una compilación exitosa, el sistema procede a eliminar los archivos temporales generados durante las fases de transformación y generación, entregando al usuario únicamente el ejecutable solicitado y garantizando la limpieza del directorio de trabajo.