#align(center)[
  #text(12pt, weight: "bold")[Resumen]
]

#pad(x: 2em)[
  Este informe documenta la cobertura de pruebas del módulo de gestión de goroutines y reparto de trabajo de GompherMP, responsable de materializar las directivas de paralelismo estructurado del lenguaje sobre las primitivas nativas de concurrencia de Go. Se presenta la suite de pruebas ejecutadas, la verificación funcional de cada primitiva de reparto de trabajo soportada (creación de regiones paralelas, reparto estático y dinámico de iteraciones, y reparto de bloques independientes) y los resultados cuantitativos de cobertura obtenidos mediante la herramienta `go test -cover`.
]
