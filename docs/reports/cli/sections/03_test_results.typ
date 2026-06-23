= Resultados de Ejecución

== Resumen cuantitativo

#figure(
  table(
    columns: (auto, auto),
    align: (left, right),
    [*Métrica*],                                            [*Valor*],
    [Total de pruebas ejecutadas],                          [20],
    [Pruebas exitosas],                                     [20],
    [Pruebas fallidas],                                     [0],
    [Pruebas exitosas con detector de carreras activo],     [20],
    [Cobertura total de instrucciones],                     [90.4%],
  ),
  caption: [Resumen cuantitativo de la ejecución de la suite],
)

== Distribución de pruebas por área

#figure(
  table(
    columns: (auto, auto, 1fr),
    align: (left, right, left),
    table.header([*Área*], [*Cantidad*], [*Propósito*]),
    [Despacho de comandos], [4], [Verifican `--version`, `--help`, la ausencia de argumentos y el comando desconocido.],
    [Validación de argumentos y opciones], [5], [Verifican el archivo ausente, el exceso de argumentos, la extensión no `.go`, el archivo inexistente y la opción inválida.],
    [Modos de error de las fases], [4], [Verifican los fallos de análisis, validación, transformación y compilación.],
    [Flujo de extremo a extremo], [5], [Verifican el flujo completo con y sin directivas, los modos detallado y de conservación del temporal, y el nombre de salida por defecto, invocando al compilador de Go.],
    [Información de directivas], [2], [Verifican la identificación de cada tipo de directiva y el caso no reconocido.],
  ),
  caption: [Distribución de pruebas por área de la interfaz],
)

== Cobertura detallada por función

#figure(
  table(
    columns: (auto, auto),
    align: (left, right),
    [*Función*],        [*Cobertura*],
    [`hasArg`],         [100.0%],
    [`run`],            [100.0%],
    [`runBuild`],       [87.8%],
    [`directiveInfo`],  [100.0%],
    [`main`],           [0.0%],
    [*Total del módulo*], [*90.4%*],
  ),
  caption: [Cobertura agregada por función del módulo],
)

== Instrucciones no cubiertas

La cobertura restante corresponde a dos clases de código que el flujo de
ejecución normal no alcanza:

- La función `main` es únicamente un envoltorio que delega en el punto de entrada
  testeable y termina el proceso con el código de salida resultante. Las pruebas
  ejercitan ese punto de entrada directamente, por lo que el envoltorio no se
  contabiliza.
- Dentro de `runBuild`, unas pocas ramas manejan fallos de entrada y salida que
  no se reproducen de forma portable en una prueba (por ejemplo, una escritura
  que falla después de haberse creado el archivo, o la resolución de una ruta que
  no falla en la práctica). Se conservan por robustez.
