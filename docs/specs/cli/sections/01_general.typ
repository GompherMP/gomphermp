= Aspectos Generales

== Propósito de la CLI
La interfaz de línea de comandos (CLI) de GompherMP ha sido diseñada como la capa de abstracción principal entre el desarrollador y el complejo ecosistema de transformación de código. Su función primordial es reducir la fricción sintáctica, permitiendo que la invocación de directivas de paralelismo tipo OpenMP sea tan sencilla como ejecutar un comando de compilación estándar. 

El CLI orquesta el flujo de trabajo completo: desde el análisis léxico del código fuente original hasta la generación de binarios optimizados, ocultando los pasos intermedios de manipulación del Árbol Sintáctico Abstracto (AST) y la inyección de dependencias del Runtime.

== Prerrequisitos del Sistema
Para garantizar la correcta ejecución del pipeline de GompherMP, el sistema anfitrión debe cumplir con las siguientes especificaciones técnicas:

- *Compilador de Go:* Versión 1.20 o superior instalada y accesible en el `PATH` del sistema.
- *Variables de Entorno:* Configuración correcta de `$GOPATH` y `$GOROOT` para permitir el enlazado dinámico de las librerías de soporte de GompherMP.
- *Permisos de Archivos:* El usuario debe contar con permisos de lectura sobre los archivos fuente (`.go`) y permisos de escritura en el directorio de destino para la generación de archivos temporales y el binario final.
- *Dependencias:* Instalación previa de la librería de Runtime de GompherMP en el entorno local de Go para asegurar el enlace durante la fase de `go build`.