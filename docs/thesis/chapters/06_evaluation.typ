
= Capítulo 6. OE3: Evaluación de GompherMP mediante benchmarks

En este capítulo se detalla el trabajo realizado para cumplir con el tercer objetivo específico, orientado a validar empíricamente la herramienta GompherMP. La evaluación aborda dos dimensiones complementarias: la eficiencia en tiempo de ejecución del código transpilado y el impacto del paradigma de directivas sobre la expresividad del desarrollador. Ambas dimensiones son necesarias para responder la pregunta central del proyecto: si es posible cerrar la brecha semántica entre el modelo de directivas de paralelismo y el ecosistema de Go, obteniendo rendimiento comparable al código paralelo manual idiomático sin sacrificar la legibilidad ni incrementar la carga cognitiva del programador. Los resultados aquí presentados toman como insumo la herramienta funcional producida en el OE2 y la someten a una suite de _benchmarks_ representativos, cuantificando el rendimiento computacional y la expresividad del código transpilado frente a las implementaciones secuencial y paralela manual en Go.


== R10: Suite de _benchmarks_ implementada

Para cumplir con este resultado, se diseñó e implementó una suite de diez _benchmarks_ de cómputo intensivo que ejercita de forma diferenciada los principales patrones de paralelismo soportados por GompherMP. La selección de los algoritmos se guió por dos criterios complementarios. El primero es la representatividad de los patrones de paralelismo que un desarrollador Go puede encontrar al paralelizar aplicaciones de propósito general: bucles de datos independientes, reducciones, ordenamiento recursivo, búsqueda combinatoria, simulación estocástica, procesamiento por etapas y grafos de dependencias entre tareas. El segundo es la cobertura de directivas: el conjunto de _benchmarks_ ejercita colectivamente la totalidad de los constructos implementados en GompherMP, desde el paralelismo estructurado de bucles (`parallel for`, `schedule`, `reduction`) hasta el paralelismo de tareas con dependencias explícitas (`task`, `taskloop`, `taskgroup`, cláusulas `depend`), pasando por los mecanismos de sincronización (`barrier`, `single`) y las secciones concurrentes heterogéneas (`parallel sections`).

Cada algoritmo fue implementado en tres variantes complementarias:

- *Secuencial (`seq`):* Implementación directa sin primitivas de concurrencia, empleada como línea base para el cálculo del _speedup_ relativo de cada variante paralela.

- *Paralela manual (`manual`):* Paralelización idiomática en Go, utilizando goroutines, `sync.WaitGroup`, canales y operaciones atómicas, siguiendo las mejores prácticas del ecosistema.

- *Paralela con GompherMP (`gompher`):* Código fuente anotado con directivas `//gompher` que, tras ser procesado por la herramienta, genera automáticamente las llamadas al runtime de GompherMP. El código original permanece libre de primitivas de concurrencia explícitas.

El algoritmo _Fibonacci_ cuenta con dos variantes GompherMP que evalúan estrategias distintas de paralelismo de tareas: `taskloop` (partición iterativa del espacio de cómputo en bloques de grano fijo) y `task_depend` (modelado explícito de las dependencias de datos mediante las cláusulas `depend(out/in)`), elevando a once las configuraciones GompherMP totales de la suite.

La @tab:tabla-descripcion-benchmarks resume los diez _benchmarks_ implementados describiendo el problema que cada uno resuelve y el algoritmo base empleado.

#figure(
  table(
    columns: (0.85fr, 2.9fr),
    stroke: 0.5pt,
    fill: (col, row) => if row == 0 { luma(230) },
    align: (col, row) => if row == 0 { center + horizon } else { left + top },
    [*Benchmark*], [*Descripción del problema y algoritmo base*],
    [MatMul],
    [Multiplicación de dos matrices cuadradas densas de punto flotante de doble precisión. Operación $O(n^3)$ sin dependencias entre filas, caso arquetípico del paralelismo de datos regular con carga homogénea entre hilos.],
    [PrefixSum],
    [Suma de prefijo (scan inclusivo) sobre un arreglo de enteros. Implementado en dos fases: reducción paralela ascendente y distribución descendente. Ejercita `reduction`, `barrier` y `single` de forma combinada.],
    [MergeSort],
    [Ordenamiento recursivo por fusión sobre un arreglo de enteros. El paralelismo se aplica en los subárboles de recursión; la fase de fusión final es inherentemente secuencial, lo que acota el _speedup_ teórico por la ley de Amdahl.],
    [Fibonacci (taskloop)],
    [Cálculo de la suma de los primeros $n$ términos de Fibonacci mediante un árbol de tareas generado con `taskloop`. Diseñado para estresar el planificador de tareas con trabajo de grano fino y alta tasa de creación.],
    [Fibonacci (depend)],
    [Misma computación que Fibonacci (taskloop) pero expresada como grafo explícito de dependencias entre tareas con cláusulas `depend(out/in)`. Evalúa el overhead del registro de dependencias para grano de tarea muy fino.],
    [Sections],
    [Tres reducciones aritméticas de distinta duración ejecutadas como secciones paralelas independientes. Representa cargas de trabajo heterogéneas sin estructura iterativa, donde el overhead de coordinación es significativo frente al trabajo útil de cada sección.],
    [QuickSort],
    [Ordenamiento recursivo por partición in-place sobre un arreglo de enteros. El árbol de recursión irregular del algoritmo permite comparar `taskloop` con `grainsize` fijo frente a goroutines recursivas que explotan toda la profundidad del árbol.],
    [Pipeline],
    [Procesamiento de datos en tres etapas independientes (filtrado, transformación y acumulación) modeladas como secciones paralelas. Representa el patrón de concurrencia estructural donde las etapas difieren en duración y no se comunican entre sí.],
    [N-Queens],
    [Conteo de todas las soluciones al problema de las $N$-reinas en un tablero $N times N$ mediante _backtracking_ recursivo. La primera dimensión del espacio de búsqueda se distribuye entre goroutines con `parallel for`, generando subproblemas de cómputo intenso y uniforme.],
    [MonteCarlo],
    [Estimación de $pi$ por el método de Monte Carlo: generación de puntos aleatorios en el cuadrado unitario y conteo de los que caen dentro del círculo inscrito. Ejercita la cláusula `reduction` sobre un cómputo estocástico con inicialización de estado local por hilo.],
    [Reduce],
    [Suma de los resultados de una función de cómputo costoso aplicada sobre un arreglo de gran tamaño. Diseñado para saturar el tiempo de CPU y minimizar el peso relativo del overhead del runtime, ofreciendo la medición más limpia del _speedup_ de la directiva `reduction`.],
  ),
  caption: [Descripción del problema y algoritmo base de cada _benchmark_ de la suite],
  kind: table,
) <tab:tabla-descripcion-benchmarks>

El _harness_ de medición fue implementado utilizando el paquete `testing` de la librería estándar de Go. Cada configuración (combinación de algoritmo, número de procesadores y variante) fue ejecutada con $n = 10$ repeticiones independientes. El grado de paralelismo se controló mediante `runtime.GOMAXPROCS(P)` al inicio de cada corrida, evaluando cinco configuraciones de escalabilidad: $P in {1, 2, 4, 8, 16}$. El _hardware_ utilizado para todas las mediciones fue un procesador AMD Ryzen 7 3700X de 8 núcleos físicos (16 hilos lógicos) bajo Go 1.22. El código fuente completo de la suite, incluyendo las tres versiones de cada algoritmo, se encuentra disponible en el repositorio en línea del proyecto referenciado en el Anexo G.


== R11: Informe de evaluación de rendimiento y escalabilidad

Este resultado consolida el análisis cuantitativo del rendimiento de GompherMP frente a las implementaciones secuencial y paralela manual. La evaluación se organiza en dos ejes complementarios: el análisis del rendimiento absoluto a la configuración máxima de $P = 16$ procesadores y el análisis de la escalabilidad conforme aumenta el grado de paralelismo de $P = 1$ a $P = 16$.

*Metodología estadística.* Dado que los tiempos de ejecución medidos bajo un sistema operativo de propósito general presentan variabilidad estocástica inherente, se adoptó el siguiente protocolo de análisis para cada configuración. La media aritmética ($overline(t)$) se empleó como estimador puntual del tiempo central. La desviación estándar muestral ($s$) cuantifica la dispersión de las $n = 10$ mediciones, y el coeficiente de variación ($"CV" = s \/ overline(t) times 100\%$) provee un indicador relativo de estabilidad que permite comparar la volatilidad entre _benchmarks_ de magnitudes distintas. El intervalo de confianza al 95% se construyó mediante la distribución $t$-Student con $nu = 9$ grados de libertad ($t_{0.025,\, 9} = 2.262$), según la fórmula:

$ "IC"_(95%) = t_(0.025,\, 9) dot.op s / sqrt(n) $

El _speedup_ relativo ($S_P$) se define como el cociente entre el tiempo medio de la versión secuencial y el tiempo medio de la variante paralela evaluada a $P$ procesadores:

$ S_P = overline(t)_"seq" / overline(t)_P $

La incertidumbre del _speedup_ fue estimada mediante propagación cuadrática de errores relativos, dado que el _speedup_ es un cociente de dos estimadores con su propia varianza:

$ sigma_(S_P) / S_P = sqrt("CV"_"seq"^2 + "CV"_P^2) $

donde $"CV"_"seq"$ y $"CV"_P$ son los coeficientes de variación de la versión secuencial y de la variante paralela, respectivamente.

*Resultados a $P = 16$ procesadores.* La @tab:tabla-resultados-p16 presenta los tiempos de ejecución medios con sus intervalos de confianza al 95% para las tres variantes de cada _benchmark_ a $P = 16$, junto con el _speedup_ calculado para la versión manual y la versión GompherMP respecto a la línea base secuencial.

#figure(
  table(
    columns: (1fr, 1.15fr, 1.15fr, 1.15fr, 1fr, 1fr),
    stroke: 0.5pt,
    fill: (col, row) => if row == 0 { luma(230) },
    align: (col, row) => if row == 0 { center + horizon } else { center + top },
    [*Benchmark*],
    [$overline(t)_"seq" plus.minus "IC"$ (ms)],
    [$overline(t)_"man" plus.minus "IC"$ (ms)],
    [$overline(t)_"gmp" plus.minus "IC"$ (ms)],
    [*Speedup \ manual*],
    [*Speedup \ GompherMP*],
    [MatMul],              [484.96 ± 9.34],  [42.08 ± 3.52],  [39.89 ± 4.39],  [11.53 ± 3.13], [12.16 ± 4.30],
    [PrefixSum],           [3.21 ± 0.21],    [1.97 ± 0.14],   [2.32 ± 0.33],   [1.63 ± 0.50],  [1.38 ± 0.68],
    [MergeSort],           [16.92 ± 1.48],   [26.12 ± 1.03],  [26.84 ± 1.59],  [0.65 ± 0.20],  [0.63 ± 0.21],
    [Fibonacci (taskloop)],[2.44 ± 0.37],    [0.55 ± 0.05],   [0.99 ± 0.46],   [4.42 ± 2.44],  [2.45 ± 3.80],
    [Fibonacci (depend)],  [2.44 ± 0.37],    [0.55 ± 0.05],   [1.38 ± 0.31],   [4.42 ± 2.44],  [1.77 ± 1.52],
    [Sections],            [13.57 ± 0.17],   [16.89 ± 2.87],  [19.80 ± 3.01],  [0.80 ± 0.43],  [0.69 ± 0.33],
    [QuickSort],           [49.30 ± 1.17],   [20.40 ± 2.40],  [32.22 ± 1.92],  [2.42 ± 0.92],  [1.53 ± 0.31],
    [Pipeline],            [52.65 ± 1.49],   [22.24 ± 2.66],  [19.92 ± 1.65],  [2.37 ± 0.92],  [2.64 ± 0.73],
    [N-Queens],            [1111.34 ± 4.73], [118.36 ± 5.16], [117.77 ± 6.16], [9.39 ± 1.30],  [9.44 ± 1.57],
    [MonteCarlo],          [24.50 ± 0.21],   [4.35 ± 0.67],   [6.93 ± 0.80],   [5.64 ± 2.76],  [3.54 ± 1.29],
    [Reduce],              [292.39 ± 1.17],  [38.19 ± 0.84],  [21.90 ± 1.06],  [7.66 ± 0.54],  [13.35 ± 2.05],
  ),
  caption: [Tiempos medios e intervalos de confianza al 95% por variante a $P = 16$ ($n = 10$). Fibonacci (depend) comparte el tiempo secuencial con Fibonacci (taskloop) por ser el mismo algoritmo base.],
  kind: table,
) <tab:tabla-resultados-p16>

#figure(
  image("../figures/speedup_p16.png"),
  caption: [Speedup relativo al secuencial a $P = 16$ procesadores para las once configuraciones evaluadas, con barras de error correspondientes al IC 95%. Los _benchmarks_ se presentan agrupados por clasificación de rendimiento de GompherMP.],
) <fig:speedup-p16>

Los resultados a $P = 16$ permiten identificar tres patrones de comportamiento diferenciados. En primer lugar, _Reduce_ constituye el resultado más destacado de la evaluación: GompherMP alcanza un _speedup_ de $13.35 times$ frente a los $7.66 times$ de la versión manual, una mejora del 74% cuyos intervalos de confianza no se solapan, confirmando su significancia estadística. Este resultado supera la eficiencia paralela teórica del hardware ($S_{16}^"gmp" / 16 = 0.83$), lo que indica que la transformación automática de la directiva `reduction` genera un patrón de acumulación privada por goroutine que evita la contención que introduce el patrón manual. En segundo lugar, _N-Queens_ logra paridad estadística prácticamente perfecta: $9.44 times$ para GompherMP frente a $9.39 times$ para el manual, con intervalos de confianza solapados, resultado que se obtiene con una sola anotación `//gompher parallel for` sobre el bucle externo de _backtracking_. En tercer lugar, _Pipeline_ muestra una ventaja de GompherMP del 11% ($2.64 times$ vs. $2.37 times$), y _MatMul_ presenta una leve ventaja del 5% dentro del margen de error.

En el extremo opuesto, _PrefixSum_, _QuickSort_ y las variantes de _Fibonacci_ exhiben un rendimiento inferior de GompherMP respecto al manual, con diferencias que oscilan entre el 18% (_PrefixSum_) y el 56% (_Fibonacci_ con `task_depend`). _MergeSort_ y _Sections_ muestran _speedup_ efectivo por debajo de $1 times$ en ambas variantes paralelas: los IC solapados confirman que la diferencia entre manual y GompherMP no es estadísticamente significativa, siendo la secuencialidad del trabajo no paralelizable el factor determinante.

*Análisis de escalabilidad.* El Anexo L presenta las curvas de _speedup_ individuales por benchmark para $P in {1, 2, 4, 8, 16}$, generadas desde los datos crudos del repositorio. Los valores a $P = 1$ en los _benchmarks_ de ordenamiento (_MergeSort_ y _QuickSort_) reflejan una ganancia algorítmica y no de paralelismo, dado que las variantes paralelas emplean `sort.Ints` de la librería estándar mientras la versión secuencial usa el algoritmo recursivo base.

A partir de las curvas, es posible identificar cuatro perfiles de escalabilidad cualitativamente distintos. _N-Queens_ y _MatMul_ presentan escalabilidad casi lineal (eficiencias de $0.59$ y $0.76$ a $P = 16$, respectivamente) con GompherMP trazando la misma curva que el manual en ambos casos, lo que confirma que la transformación automática no introduce penalizaciones de escalabilidad para cargas regularmente paralelizables. _Reduce_ exhibe el caso más llamativo: su curva de GompherMP se separa progresivamente de la del manual a partir de $P = 4$, alcanzando $13.35 times$ frente a $7.66 times$ a $P = 16$, lo que sugiere que el patrón de reducción generado automáticamente elimina un cuello de botella de sincronización presente en la versión manual desde las primeras etapas del escalado.

_PrefixSum_ exhibe el perfil más irregular: la variante manual escala hasta $P = 8$ ($2.73 times$) pero retrocede a $1.63 times$ a $P = 16$, mientras GompherMP nunca supera $1.38 times$. Este comportamiento regresivo es consistente con la ley de Amdahl aplicada a la fase secuencial forzada por la suma prefija. _MergeSort_ muestra degradación progresiva en ambas variantes conforme aumenta $P$, alcanzando valores por debajo de $1 times$ a $P = 16$: la fase de _merge_ final domina el tiempo de ejecución independientemente de la estrategia de paralelización. Finalmente, _Fibonacci_ presenta el patrón de escalabilidad más errático de GompherMP, con un retroceso entre $P = 4$ y $P = 8$ atribuible a contención esporádica en el _pool_ para tareas de grano muy fino.


== R12: Reporte de análisis comparativo sobre expresividad y productividad

Para cumplir con este resultado, se elaboró un análisis comparativo de la expresividad y la productividad del desarrollador entre las versiones paralela manual y paralela con GompherMP de cada _benchmark_ de la suite. La evaluación se estructura en dos ejes complementarios: un análisis cuantitativo basado en el conteo de líneas de código (LoC) de las funciones de paralelización, y un análisis cualitativo orientado a caracterizar la legibilidad, la separación de incumbencias y la propensión a errores de concurrencia de cada enfoque.

*Metodología del análisis cuantitativo.* La unidad de análisis son las funciones que implementan la lógica de paralelización en cada variante. Se excluyen los auxiliares de medición de tiempo, las funciones secuenciales, el `main` y las rutinas de utilidad compartidas entre variantes. Para los _benchmarks_ con múltiples funciones paralelas (PrefixSum, QuickSort, Fibonacci), el conteo agrega todas las funciones que corresponden a cada variante. La variación relativa de LoC se define como:

$ Delta"LoC" = (L_"gmp" - L_"man") / L_"man" times 100\% $

donde $L_"man"$ y $L_"gmp"$ son los conteos de líneas de la versión manual y GompherMP, respectivamente. Un valor negativo indica que GompherMP reduce el código; un valor positivo, que lo incrementa. Adicionalmente, se contabiliza el número de directivas GompherMP requeridas como indicador del nivel de anotación.

La elección de LoC como proxy cuantitativo de la expresividad se fundamenta en la relación directa entre el número de líneas dedicadas a la infraestructura de concurrencia y la carga cognitiva del desarrollo: cada línea de _boilerplate_ (goroutines, canales, mutexes, grupos de espera) es un concepto adicional que el programador debe gestionar y verificar manualmente, independientemente de la lógica algorítmica que pretende paralelizar. Nanz y Furia (2015) documentan que los lenguajes que producen soluciones más concisas para tareas equivalentes tienden a reducir la probabilidad de errores derivados de la complejidad técnica, observación que motiva el uso de LoC como indicador de la eficiencia expresiva. No obstante, dado que el recuento de líneas no captura la claridad estructural del código, el análisis cuantitativo se complementa con un análisis cualitativo orientado a examinar la separación entre lógica algorítmica y gestión de concurrencia en cada variante.

La @tab:tabla-loc-comparativo presenta los resultados del análisis cuantitativo para las once configuraciones de la suite, ordenadas de mayor reducción de LoC a mayor incremento.

#figure(
  table(
    columns: (1.6fr, 0.7fr, 0.7fr, 0.9fr, 0.7fr),
    stroke: 0.5pt,
    fill: (col, row) => if row == 0 { luma(230) },
    align: (col, row) => if row == 0 { center + horizon } else { center + top },
    [*Benchmark*], [$L_"man"$], [$L_"gmp"$], [*Directivas*], [$Delta$*LoC*],
    [Reduce],               [34], [10], [1],  [-71%],
    [MonteCarlo],           [26], [11], [1],  [-58%],
    [PrefixSum],            [55], [27], [6],  [-51%],
    [MatMul],               [24], [12], [1],  [-50%],
    [Sections],             [26], [26], [4],  [0%],
    [MergeSort],            [24], [26], [2],  [+8%],
    [QuickSort],            [21], [23], [4],  [+10%],
    [Fibonacci (taskloop)], [16], [21], [2],  [+31%],
    [N-Queens],             [14], [19], [2],  [+36%],
    [Pipeline],             [27], [54], [9],  [+100%],
    [Fibonacci (depend)],   [16], [39], [8],  [+144%],
  ),
  caption: [Conteo de líneas de código (LoC) de las funciones paralelas por variante, número de directivas GompherMP utilizadas y variación relativa respecto al paralelo manual ($Delta$LoC)],
  kind: table,
) <tab:tabla-loc-comparativo>

#figure(
  image("../figures/loc_comparison.png"),
  caption: [Comparación de LoC entre las versiones manual y GompherMP para cada _benchmark_ de la suite. Las etiquetas indican el $Delta$LoC respecto al paralelo manual; valores negativos (verde) indican reducción de código con GompherMP.],
) <fig:loc-comparativo>

Los resultados de la @tab:tabla-loc-comparativo permiten identificar tres patrones de comportamiento diferenciados en cuanto a la expresividad del código.

*GompherMP reduce significativamente el LoC ($Delta$LoC $lt.eq -50%$).* Los _benchmarks_ _Reduce_ (−71%), _MonteCarlo_ (−58%), _PrefixSum_ (−51%) y _MatMul_ (−50%) concentran el mayor beneficio de productividad. En los tres primeros, una sola directiva sobre el bucle externo reemplaza una estructura manual de entre 24 y 34 líneas que combina arreglos de resultados parciales, instancias de `sync.WaitGroup`, bucles de despacho de goroutines y bucles de consolidación. El caso más ilustrativo es _Reduce_: la versión manual requiere 34 líneas para paralelizar un cuerpo de bucle de 4 líneas de algoritmo, mientras que la versión GompherMP con la directiva `//gompher parallel for schedule(dynamic, 64) reduction(max:m)` expresa el mismo cómputo en 10 líneas, de las cuales 9 corresponden a la lógica algorítmica y una sola es la anotación de paralelismo. En _PrefixSum_, aunque el ahorro también es del 51%, se distribuye entre dos funciones y requiere 6 directivas para modelar la sincronización explícita con `barrier` y `single` dentro de una región paralela estructurada, ejercitando colectivamente un subconjunto más amplio de la especificación de GompherMP.

*Paridad de LoC ($|Delta"LoC"| lt.eq 10%$).* _MergeSort_ (+8%), _QuickSort_ (+10%) y _Sections_ (0%) presentan conteos prácticamente equivalentes entre versiones. En _MergeSort_ y _QuickSort_, la versión GompherMP introduce 2 a 4 líneas adicionales atribuibles a los bloques explícitos `{}` requeridos por la sintaxis de directivas de tarea y a las variables auxiliares necesarias para capturar el índice de iteración por tarea. En _Sections_, la estructura de bloques anidada impuesta por `parallel sections` / `section` es casi idéntica en LoC a los cierres anónimos de goroutines de la versión manual, siendo la diferencia estructural y no cuantitativa.

*GompherMP incrementa el LoC ($Delta"LoC" > 10%$).* _Fibonacci (taskloop)_ (+31%), _N-Queens_ (+36%), _Pipeline_ (+100%) y _Fibonacci (depend)_ (+144%) son los _benchmarks_ donde la versión GompherMP requiere más líneas que el manual. Este incremento responde a dos causas diferenciadas. En _N-Queens_ y _Fibonacci (taskloop)_, el incremento es menor (5 a 7 líneas) y se debe a que las tareas GompherMP requieren variables auxiliares explícitas para capturar el índice de columna por tarea (e.g., `c := col`) y un arreglo de resultados indexado por posición, mientras que la versión manual puede acumular resultados directamente en un canal con semántica de cola. En _Pipeline_ y _Fibonacci (depend)_, el incremento es estructural: expresar un grafo de dependencias de datos explícito con las cláusulas `depend(out/in)` requiere declarar variables centinela para cada arco del DAG (las variables `s0..s3` y `f0..f3` en _Pipeline_, las parciales `r0..r3`, `m0`, `m1` en _Fibonacci (depend)_) y anotar cada tarea con sus roles de entrada y salida, elevando el LoC al doble o más respecto al manual.

*Análisis cualitativo.* Más allá del conteo de líneas, la diferencia estructural entre ambos enfoques se aprecia con mayor claridad sobre un ejemplo concreto. La @fig:fibonacci-comparativo muestra las funciones paralelas del _benchmark_ Fibonacci en sus dos variantes, elegido por ser representativo del caso intermedio: GompherMP introduce líneas adicionales (+31%) pero reestructura el código de forma significativa.

#figure(
  grid(
    columns: (1fr, 1fr),
    gutter: 1.2em,
    align(top)[
      *Versión manual* ($L_"man" = 16$ líneas)
      ```go
      func sumManual(data []float64) float64 {
          p, chunk := numProcs(),
                      len(data)/numProcs()
          ch := make(chan float64, p)
          for t := 0; t < p; t++ {
              lo, hi := t*chunk, (t+1)*chunk
              if t == p-1 {
                  hi = len(data)
              }
              go func(lo, hi int) {
                  ch <- heavySqrtSum(data, lo, hi)
              }(lo, hi)
          }
          total := 0.0
          for range p {
              total += <-ch
          }
          return total
      }
      ```
    ],
    align(top)[
      *Versión GompherMP (taskloop)* ($L_"gmp" = 21$ líneas)
      ```go
      func sumTaskloop(data []float64) float64 {
          p := numProcs()
          results := make([]float64, p)
          chunkSize := len(data) / p
          //gompher taskgroup
          {
              //gompher taskloop grainsize(1)
              for c := 0; c < p; c++ {
                  lo, hi := c*chunkSize,
                            (c+1)*chunkSize
                  if c == p-1 {
                      hi = len(data)
                  }
                  results[c] = heavySqrtSum(
                      data, lo, hi)
              }
          }
          total := 0.0
          for _, v := range results {
              total += v
          }
          return total
      }
      ```
    ]
  ),
  caption: [Comparación de implementaciones paralelas del _benchmark_ Fibonacci. Izquierda: versión manual con goroutines y canal. Derecha: versión GompherMP con directivas `taskgroup` y `taskloop`. Los dos comentarios `//gompher` son las únicas anotaciones de paralelismo.],
  kind: image,
) <fig:fibonacci-comparativo>

En la versión manual, el algoritmo (la llamada a `heavySqrtSum` con los índices de segmento) queda enterrado dentro de un cierre anónimo de goroutine: el programador debe razonar simultáneamente sobre la aritmética de partición del arreglo y sobre el protocolo del canal: su capacidad, el orden de envío y la recolección exacta de `p` valores. En la versión GompherMP, el bucle de partición es el cuerpo visible del código, precedido por dos anotaciones que declaran la intención de paralelismo; el cuerpo puede leerse como código secuencial con paralelismo implícito. El coste de esta reestructuración es la necesidad de un arreglo auxiliar `results[]` para acumular los valores por tarea, lo que explica las 5 líneas adicionales: la directiva `taskloop` no dispone de un mecanismo de reducción implícita (a diferencia de `parallel for reduction`), de modo que el programador gestiona la recolección de resultados de forma explícita.

Este ejemplo ilustra un límite de expresividad del diseño actual: cuando la acumulación de resultados por tarea es necesaria, `taskloop` es menos conciso que `parallel for reduction`. Sin embargo, incluso en este escenario de mayor LoC, la versión GompherMP elimina los conceptos más propensos a errores, como goroutines explícitas, un canal con capacidad fija y el cierre con captura de variables, que son la combinación más habitual de errores de concurrencia documentados en el ecosistema de Go (Tu et al., 2019).

La distinción estructural se agudiza en los extremos del espectro. En _Reduce_, el mismo patrón de distribución de trabajo se expresa con `parallel for schedule(dynamic, 64) reduction(max:m)`: una sola línea de directiva reemplaza las 24 líneas de boilerplate de la versión manual (arreglo de parciales, `sync.WaitGroup`, bucle de despacho, bucle de consolidación), dejando solo las 9 líneas del cuerpo algorítmico. En el otro extremo, _Pipeline_ y _Fibonacci (depend)_ requieren variables centinela explícitas por cada arco del grafo de dependencias, como `s0..s3` y `f0..f3` en _Pipeline_, lo que incrementa el LoC pero deja el flujo de datos completamente auditable en el código fuente, lo que puede facilitar la revisión y el mantenimiento.

La variante manual exige dominar goroutines, canales, `sync.WaitGroup`, `sync.Mutex` y operaciones atómicas del paquete `sync/atomic`, aplicándolos correctamente bajo el modelo de memoria de Go. La variante GompherMP desplaza este conocimiento hacia la especificación declarativa de directivas, reduciendo la barrera de entrada para paralelizar código existente. En síntesis, el análisis confirma que el impacto de GompherMP en la expresividad del código es fuertemente dependiente del patrón de paralelismo: para bucles _data-parallel_ con reducción o distribución estática, GompherMP elimina entre el 50% y el 71% del código eliminando por completo el _boilerplate_ de concurrencia; para grafos de dependencias de tareas, el impacto en LoC es inverso pero se gana legibilidad declarativa del flujo de datos. Para los patrones de paralelismo más habituales en aplicaciones Go de propósito general, que son también los _benchmarks_ con mejores métricas de rendimiento, GompherMP resulta más expresivo y más productivo que la paralelización manual idiomática.


== Discusión de resultados

La evaluación conjunta de rendimiento (R11) y expresividad (R12) dibuja un cuadro coherente sobre las condiciones bajo las cuales GompherMP es eficaz. El patrón más claro emerge del grupo de _benchmarks_ que concentran los mejores resultados en ambas dimensiones: _Reduce_, _MatMul_ y _N-Queens_ logran simultáneamente las mayores reducciones de LoC (−71%, −50% y una paridad con ventaja de rendimiento, respectivamente) y los _speedups_ más altos o parejos de GompherMP. El factor unificador no es el tipo de algoritmo sino la estructura del paralelismo: en los tres casos, el trabajo por goroutine es suficientemente pesado como para amortizar el _overhead_ del _pool_, la distribución de carga es regular y una sola directiva captura completamente la intención paralela. Esta convergencia entre simplicidad de expresión y eficiencia de ejecución es el argumento más fuerte a favor del modelo: la directiva `reduction` no solo elimina el _boilerplate_ del acumulador manual, sino que genera un patrón de acumulación privada por goroutine que evita la contención que introduce el patrón de canales de la versión manual en _Reduce_, resultando en un _speedup_ de $13.35 times$ frente a $7.66 times$.

_Pipeline_ matiza este cuadro y aporta un hallazgo relevante: GompherMP supera al manual en _speedup_ ($2.64 times$ vs. $2.37 times$) a pesar de un incremento de LoC del +100%. El grafo de dependencias explícito que obliga a declarar variables centinela por arco del DAG permite al _pool_ persistente despachar las etapas con una eficiencia que el patrón de canales no alcanza, porque el _pool_ puede absorber el desequilibrio de duración entre etapas sin el costo de creación y destrucción de goroutines en cada región paralela. Esto demuestra que la verbosidad adicional impuesta por `depend` en escenarios de _pipeline_ es un costo de expresividad con retorno en rendimiento, no un defecto del modelo.

Las limitaciones de rendimiento identificadas en _PrefixSum_, _MonteCarlo_, _QuickSort_ y _Fibonacci_ tienen causas implementativas concretas que vale la pena examinar en conjunto. En _PrefixSum_, la función `getGoroutineID()` parsea el _stack trace_ de la goroutine en cada llamada a `Barrier()` y `Single()`, acumulando una latencia de 1 a 5 µs por invocación que se vuelve significativa con sincronización frecuente; la solución es sustituir la introspección por un identificador pasado explícitamente al contexto del equipo. En _MonteCarlo_, la ausencia de un equivalente a `omp_get_thread_num()` impide que cada goroutine inicialice su propio generador de números aleatorios, forzando un RNG por iteración con el _overhead_ asociado; la versión manual resuelve esto con una variable local al cierre. En _QuickSort_, la directiva `taskloop` particiona el espacio de iteración en bloques de tamaño fijo y no puede capturar la recursión paralela de profundidad variable que la versión manual explota con goroutines recursivas adaptativas. En _Fibonacci_, el _overhead_ por tarea de `TaskWithDepend`, compuesto por asignación dinámica y adquisición del _lock_ del registro de dependencias, domina el tiempo total cuando el trabajo útil por tarea es submilisegundo. Ninguna de estas limitaciones es intrínseca al paradigma de directivas; todas tienen soluciones técnicas identificadas y trazan con precisión la agenda de mejoras de la implementación.

Los _benchmarks_ de paridad estadística, _MergeSort_ y _Sections_, aportan una evidencia complementaria igualmente importante: cuando el cuello de botella es estructural, GompherMP no introduce penalizaciones artificiales. En _MergeSort_, el _speedup_ efectivo por debajo de $1 times$ en ambas variantes paralelas se explica íntegramente por la fase de _merge_ secuencial final que domina el tiempo de ejecución conforme a la ley de Amdahl; en _Sections_, la alta varianza y los intervalos de confianza solapados reflejan que el overhead de coordinación domina sobre el trabajo útil de cada sección corta. En ambos casos, GompherMP reproduce el comportamiento del manual: no empeora lo que ya es difícil de paralelizar.

Finalmente, _Fibonacci (depend)_ cierra el espectro como el único caso donde el paradigma no ofrece beneficio neto sobre el enfoque manual para las granularidades evaluadas. Su mayor incremento de LoC (+144%) y su _speedup_ más bajo ($1.77 times$) son expresión del mismo fenómeno: el _overhead_ del registro de dependencias perjudica tanto al desarrollador, que debe especificar el grafo explícitamente con variables centinela, como al runtime, que debe resolverlo con una estructura sincronizada en tiempo de ejecución. Este resultado delimita el umbral de granularidad por debajo del cual `taskloop` y `task depend` no son las herramientas adecuadas, información valiosa para guiar al usuario en la elección del patrón de directiva correcto.

En conjunto, la evaluación valida la hipótesis central del proyecto: el paradigma de paralelismo basado en directivas es técnicamente viable en el ecosistema de Go y permite cerrar la brecha semántica entre el modelo de directivas y las primitivas de concurrencia nativas del lenguaje, ofreciendo rendimiento comparable o superior al del código paralelo manual idiomático con un esfuerzo de programación significativamente menor para los patrones de paralelismo más comunes en aplicaciones Go de propósito general. Las brechas de rendimiento observadas no cuestionan la viabilidad del modelo; identifican con precisión los cuatro puntos de la implementación actual donde existe margen de mejora y configuran la agenda de desarrollo de una versión más madura de la herramienta.
