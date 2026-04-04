= Estructura de Comandos y Opciones

== Sintaxis Base
La interacción con la herramienta se realiza a través del comando raíz `gompher`, seguido del subcomando de construcción `build`. La sintaxis general se define de la siguiente manera:


#figure(
  ```bash
  gompher build [opciones] <ruta_del_archivo.go>
  ```,
  caption: [Sintaxis de invocación del comando `build`]
)


El argumento principal es la ruta al archivo fuente que contiene las directivas `//gompher`. Si no se especifican opciones adicionales, la herramienta generará un ejecutable con el nombre predeterminado del archivo fuente en el directorio actual.

== Banderas (Flags) Soportadas
Para ofrecer un mayor control sobre el proceso de transpilación y compilación, se han definido las siguientes banderas:

#table(
  columns: (auto, auto, 1fr),
  inset: 10pt,
  align: horizon,
  [*Bandera*], [*Alias*], [*Descripción*],
  [`--output`], [`-o`], [Especifica el nombre y la ruta del binario ejecutable resultante.],
  [`--verbose`], [`-v`], [Activa el modo detallado. Imprime en la terminal las fases del pipeline y las directivas detectadas en el AST.],
  [`--keep-temp`], [`-k`], [Conserva los archivos fuente `.go` intermedios generados tras la inyección del AST. Útil para depuración.],
  [`--help`], [`-h`], [Despliega el menú de ayuda con la descripción de comandos y sintaxis de las directivas.],
  [`--version`], [], [Muestra la versión actual de la herramienta GompherMP.],
)

== Ejemplos de Uso

En lugar de invocar cada opción por separado, GompherMP permite combinar banderas para ajustarse a las necesidades del desarrollador. A continuación se presentan los escenarios de uso más comunes:

*1. Compilación estándar con salida personalizada:*
Genera el ejecutable con un nombre específico, silenciando los logs de los módulos internos.

#figure(
  ```sh
  gompher build -o mi_programa_paralelo algoritmo.go
  ```,
  caption: [Compilación estándar con flag de salida]
)

*2. Modo depuración (Debug del Transpilador):*
Ideal para analizar qué directivas detectó el AST y conservar el código Go intermedio generado (útil para verificar la inyección de k-funciones).

#figure(
  ```bash
  gompher build --verbose --keep-temp algoritmo.go
  ```,
  caption: [Ejecución en modo detallado conservando temporales]
)

*3. Consulta del Menú de Ayuda:*

#figure(
  ```text
  $ gompher build --help
  GompherMP CLI - Transpilador de paralelismo estructurado para Go
  Uso: gompher build [opciones] <archivo.go>
  ```,
  caption: [Salida estándar del comando --help]
)