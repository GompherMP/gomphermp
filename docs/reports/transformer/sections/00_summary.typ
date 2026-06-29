#align(center)[
  #text(12pt, weight: "bold")[Resumen]
]

#pad(x: 2em)[
  Este informe documenta la cobertura de pruebas del módulo transformador de
  GompherMP, responsable de reescribir el Árbol Sintáctico Abstracto (AST) de
  un programa Go anotado con directivas `//gompher` en un programaequivalente
  que invoca a la librería de runtime. Se presenta la descripción del móduloy
  su flujo de transformación, la verificación funcional de cada directiva y
  cláusula soportada, la evidencia de la transformación de la directiva
  `atomic` y de las cláusulas de gestión de datos, y los resultados
  cuantitativos de cobertura obtenidos mediante la herramienta
  `go test -cover`.
]

