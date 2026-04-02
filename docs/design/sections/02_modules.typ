= Módulos del software

La herramienta GompherMP cuenta con los siguientes módulos:

+ CLI, el cual mapea argumentos de línea de comandos.
+ Parser, el cual transforma código fuente con directivas de GompherMP en representación intermedia.
+ Transformer, el cual inspecciona una representación intermedia, revisa su sintaxis GompherMP y realiza las transformaciones correspondientes a cada directiva o cláusula.
+ Printer, el cual transforma representación intermedia en código fuente del lenguaje Go.

Naturalmente, el módulo más relevante es el de transformer, el cual será descrito a detalle en una sección posterior.

== Interacciones entre módulos

A continuación se detallan las interacciones entre módulos del programa mediante el siguiente diagrama que representa el caso de uso principal del programa, transformar un archivo en lenguaje go con sintaxis enriquecida a un archivo en lenguaje go estándar:

#figure(
  image("../assets/modules_diagram.png", width: 100%),
  caption: [
    Diagrama de módulos e interacciones de GompherMP
  ],
)

Donde:

- main representa el punto de entrada del programa.
- m_cli representa al módulo CLI.
- m_parser representa al módulo parser.
- m_transformer representa al módulo transformer.
- m_printer representa al módulo printer.
- f_input representa el archivo de entrada en lenguaje Go con sintaxis enriquecida.
- f_output representa el archivo de salida en lenguaje Go con las directivas aplicadas.

Se observan las siguientes interacciones:

- main llama a m_cli para decidir que acción debe realizar el programa.
- main llama a m_parser para transformar f_input a una representación intermedia a ser manipulada.
- main llama a m_transformer y le entrega dicha representación intermedia para recibir una representación intermedia con las directivas y cláusulas GompherMP ya aplicados.
- main llama a m_printer y le entrega la representación intermedia transformada previamente y genera el código fuente correspondiente en lenguaje Go estándar en f_output.