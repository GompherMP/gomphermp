= Cobertura Funcional

Esta sección demuestra que cada comando, opción y modo de error de la interfaz
cuenta con pruebas dedicadas.

== Comandos y opciones

#figure(
  table(
    columns: (auto, 1fr),
    align: (left, left),
    [*Invocación*], [*Comportamiento verificado*],
    [`gompher build <archivo.go>`], [Transpila el archivo y lo compila, generando un binario con el nombre del archivo de entrada.],
    [`-o`, `--output <ruta>`], [Especifica la ruta del binario resultante.],
    [`-v`, `--verbose`], [Imprime las fases del flujo y las directivas detectadas.],
    [`-k`, `--keep-temp`], [Conserva el archivo Go intermedio generado.],
    [`-h`, `--help`], [Muestra el uso y las opciones disponibles.],
    [`--version`], [Imprime la versión de la herramienta.],
  ),
  caption: [Comandos y opciones de la interfaz, con el comportamiento verificado],
)

== Manejo de errores

La interfaz distingue cada modo de fallo y termina con un código de salida
distinto de cero y un mensaje específico. Las pruebas cubren los siguientes
casos:

#figure(
  table(
    columns: (auto, 1fr),
    align: (left, left),
    [*Caso*], [*Resultado esperado*],
    [Sin argumentos o comando desconocido], [Mensaje de uso y código de salida distinto de cero.],
    [Sin archivo, demasiados argumentos o extensión no `.go`], [Mensaje de error de argumentos.],
    [Archivo inexistente], [Mensaje de acceso al archivo.],
    [Opción inválida], [Mensaje de uso de la opción.],
    [Error de análisis sintáctico], [Mensaje de fallo de análisis.],
    [Error de validación contextual], [Mensaje de fallo de validación.],
    [Error de transformación], [Mensaje de fallo de transformación.],
    [Error de compilación], [Mensaje de fallo del compilador de Go.],
  ),
  caption: [Modos de error verificados por la suite],
)
