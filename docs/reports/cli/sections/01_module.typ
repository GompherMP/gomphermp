= Descripción del Módulo

La interfaz de línea de comandos es el ejecutable que el usuario invoca para
transpilar y compilar un programa. Coordina, en orden, a los módulos de la
herramienta: lee el código fuente, lo entrega al analizador sintáctico, valida
las reglas contextuales, aplica la transformación del AST, serializa el resultado
a un archivo Go y finalmente invoca al compilador estándar de Go para producir el
binario. La interfaz oculta estos pasos intermedios y expone una única operación
de construcción con un conjunto reducido de opciones.

== Ubicación

El módulo reside en el paquete `cmd/gompher/` del repositorio.

#figure(
  table(
    columns: (auto, 1fr),
    align: (left, left),
    [*Archivo*],            [*Responsabilidad*],
    [`main.go`],            [Punto de entrada, despacho de argumentos, orquestación del flujo de construcción y manejo de errores de cada fase.],
    [`main_test.go`],       [Suite de pruebas de la interfaz: despacho de comandos, validación de argumentos, modos de error y flujo de extremo a extremo.],
  ),
  caption: [Archivos que componen la interfaz de línea de comandos],
)

La interfaz se apoya en el paquete `flag` de la biblioteca estándar de Go para el
análisis de las opciones, sin dependencias externas.

== Flujo de procesamiento

El comando de construcción ejecuta una secuencia de fases. Cualquier fase que
falle detiene el flujo, emite un mensaje de error descriptivo y termina con un
código de salida distinto de cero.

#figure(
  table(
    columns: (auto, 1fr),
    align: (left, left),
    [*Fase*], [*Descripción*],
    [Lectura], [Valida la extensión y la existencia del archivo de entrada y lee su contenido.],
    [Análisis], [El analizador sintáctico extrae las directivas y cláusulas del código fuente.],
    [Validación], [Se comprueban las reglas contextuales antes de mutar el árbol.],
    [Transformación], [El transformador reescribe el AST inyectando las llamadas al runtime.],
    [Serialización], [El AST transformado se escribe a un archivo Go temporal.],
    [Compilación], [Se invoca al compilador estándar de Go sobre el archivo temporal para producir el binario.],
  ),
  caption: [Fases del flujo de construcción de la interfaz],
)

El archivo temporal se elimina al finalizar, salvo que se solicite conservarlo
con la opción correspondiente.

== Metodología de pruebas

La suite de pruebas ejercita la interfaz a través de su punto de entrada
testeable, capturando el código de salida y la salida de texto en lugar de
escribir a los flujos del proceso. Esto permite verificar tanto el resultado
como los mensajes emitidos. Las pruebas cubren el despacho de comandos, la
validación de argumentos y opciones, cada modo de error de las fases, y un
conjunto de pruebas de extremo a extremo que ejecutan el flujo completo,
incluida la invocación real al compilador de Go, para confirmar que un programa
anotado se transpila y compila correctamente.
