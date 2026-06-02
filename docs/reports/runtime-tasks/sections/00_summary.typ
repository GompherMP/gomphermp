#align(center)[
  #text(12pt, weight: "bold")[Resumen]
]

#pad(x: 2em)[
  Este informe documenta la cobertura de pruebas del módulo de tareas y dependencias de datos de GompherMP, responsable de materializar las directivas de paralelismo basado en tareas del lenguaje sobre las primitivas nativas de concurrencia de Go. Se presenta la suite de pruebas ejecutadas, la verificación funcional de cada primitiva soportada (creación asíncrona de tareas, sincronización de subtareas, grupos de tareas con barrera profunda, distribución de iteraciones como tareas y ordenamiento por dependencias de datos) y los resultados cuantitativos de cobertura obtenidos mediante la herramienta `go test -cover`. La suite comprende 35 pruebas que cubren el comportamiento correcto, casos límite, garantías de ordenamiento bajo dependencias y ausencia de condiciones de carrera.
]
