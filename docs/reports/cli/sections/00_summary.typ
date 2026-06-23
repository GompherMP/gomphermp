#align(center)[
  #block(fill: luma(240), inset: 12pt, radius: 4pt, width: 100%)[
    #text(weight: "bold")[Resumen]

    #set par(justify: true)
    #set align(left)
    Este informe documenta la cobertura de pruebas de la interfaz de línea de
    comandos de GompherMP, el punto de entrada que orquesta el flujo completo de
    la herramienta: desde la lectura de un archivo Go anotado con directivas
    `//gompher` hasta la generación del binario final. Se presenta la descripción
    del módulo y su flujo de procesamiento, la verificación funcional de los
    comandos, opciones y modos de error, y los resultados cuantitativos de
    cobertura obtenidos mediante la herramienta `go test -cover`.
  ]
]
