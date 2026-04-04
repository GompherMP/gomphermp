= Manejo de Errores y Excepciones

El CLI de GompherMP está diseñado para capturar errores en diferentes etapas del pipeline de arquitectura, proporcionando retroalimentación clara al desarrollador para facilitar la depuración sin exponer directamente el código temporal autogenerado. Los errores se clasifican según el módulo donde ocurren:

== Errores de Análisis (Módulo Parser)
Ocurren cuando el archivo fuente contiene sintaxis nativa de Go inválida, lo que impide construir el Árbol Sintáctico Abstracto (AST). En este escenario, el proceso se aborta inmediatamente.
- *Causas comunes:* Falta de llaves `{}`, errores tipográficos en funciones o variables no declaradas en el código original.

== Errores de Transformación (Módulo Transformer)
Son los errores específicos del dominio de GompherMP. Se detectan cuando el módulo *Transformer* inspecciona el AST y encuentra discrepancias en la sintaxis de los comentarios especiales.
- *Causas comunes:* Uso de una directiva inexistente (ej. `//gompher paralel` con una sola 'l'), o variables no declaradas dentro de cláusulas como `reduction` o `private`.

== Errores de Compilación (Invocación Nativa)
Ocurren en la fase final, cuando el CLI invoca a `go build` utilizando el código emitido por el módulo *Printer*. Generalmente se deben a dependencias faltantes en el entorno.
- *Causas comunes:* El usuario no tiene instalada la librería del runtime de GompherMP (`gomphermp_runtime`) en su `go.mod`.

== Ejemplos de Interacción en Terminal

*Escenario 1: Error de Transformación (Sintaxis Inválida)*
#figure(
  ```text
  $ gompher build algoritmo.go
  [Error] Transformación fallida en algoritmo.go:14
  > //gompher task depend(in: )
  Detalle: La cláusula 'depend' requiere al menos una variable válida.
  Compilación abortada.
  ```,
  caption: [Salida de terminal ante error en el Transformer]
)

*Escenario 2: Compilación Exitosa en Modo Detallado (Verbose)*

#figure(
  ```text
  $ gompher build -v algoritmo.go
  [INFO] Analizando AST de algoritmo.go...` \
  [INFO] Directiva 'parallel for' detectada en línea 22.
  [INFO] Transformando bloque a función encapsulada (Hash: dab070c2...).
  [INFO] Generando código temporal...
  [INFO] Ejecutando go build...
  [EXITO] Binario generado: ./algoritmo
  ```,
  caption: [Salida de terminal para una compilación exitosa en modo detallado]
)