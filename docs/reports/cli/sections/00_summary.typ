#align(center)[
  #text(12pt, weight: "bold")[Resumen]
]

#pad(x: 2em)[
  Este informe documenta la cobertura de pruebas de la interfaz de línea de
  comandos de GompherMP, el punto de entrada que orquesta el flujo completode
  la herramienta: desde la lectura de un archivo Go anotado con directivas
  `//gompher` hasta la generación del binario final. Se presenta ladescripción
  del módulo y su flujo de procesamiento, la verificación funcional de los
  comandos, opciones y modos de error, y los resultados cuantitativos de
  cobertura obtenidos mediante la herramienta `go test -cover`.
]
