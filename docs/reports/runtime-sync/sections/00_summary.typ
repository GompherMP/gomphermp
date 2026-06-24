#align(center)[
  #text(12pt, weight: "bold")[Resumen]
]

#pad(x: 2em)[
  Este informe documenta la cobertura de pruebas del módulo de mecanismos de sincronización de GompherMP, responsable de coordinar la ejecución de las goroutines dentro de una región paralela y de proteger el acceso a memoria compartida. Se presenta la suite de pruebas ejecutadas, la verificación funcional de cada primitiva de sincronización soportada (exclusión mutua, ejecución única, ejecución maestra y barrera de equipo) y los resultados cuantitativos de cobertura obtenidos mediante la herramienta `go test -cover`.
]
