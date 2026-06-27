
= Capítulo 6. OE3: Evaluación de GompherMP mediante benchmarks

En este capítulo se detalla el trabajo realizado para cumplir con el tercer objetivo específico, orientado a validar empíricamente la herramienta GompherMP. Este objetivo materializa la evaluación de rendimiento y escalabilidad de la solución desarrollada. Los resultados aquí presentados toman como insumo la herramienta funcional producida en el OE2 y la someten a una suite de _benchmarks_ representativos, cuantificando el rendimiento computacional del código transpilado frente a las implementaciones secuencial y paralela manual en Go.


== R10: Suite de _benchmarks_ implementada

Para cumplir con este resultado, se diseñó e implementó una suite de diez _benchmarks_ de cómputo intensivo que ejercita de forma diferenciada los principales patrones de paralelismo soportados por GompherMP. La selección de los algoritmos se guió por dos criterios complementarios. El primero es la representatividad de los dominios canónicos del cómputo paralelo: álgebra lineal densa, búsqueda combinatoria, ordenamiento, simulación numérica estocástica, procesamiento por etapas y operaciones de reducción sobre datos. El segundo es la cobertura de directivas: el conjunto de _benchmarks_ ejercita colectivamente la totalidad de los constructos implementados en GompherMP, desde el paralelismo estructurado de bucles (`parallel for`, `schedule`, `reduction`) hasta el paralelismo de tareas con dependencias explícitas (`task`, `taskloop`, `taskgroup`, cláusulas `depend`), pasando por los mecanismos de sincronización (`barrier`, `single`) y las secciones concurrentes heterogéneas (`parallel sections`).

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


== Discusión de resultados

Los resultados de la evaluación permiten clasificar el comportamiento de GompherMP en tres categorías, cuyas causas raíz apuntan tanto a fortalezas del diseño como a limitaciones concretas de la implementación actual.

*GompherMP supera o iguala al manual (Reduce, Pipeline, N-Queens, MatMul).* El factor común en este grupo es que se trata de cargas de trabajo _compute-bound_ con patrones de acceso regular o tareas heterogéneas en duración. En _Reduce_, la directiva `reduction` genera un acumulador privado por goroutine que se combina al final, eliminando toda contención durante la fase de cómputo; la versión manual utiliza un patrón con mayor sincronización intermedia. En _Pipeline_, el _pool_ persistente de goroutines despacha las secciones de forma más eficiente que las goroutines _ad hoc_ con `sync.WaitGroup` cuando las tareas son heterogéneas en duración, ya que el _pool_ puede absorber el desequilibrio de carga sin el costo de creación y destrucción de hilos en cada región paralela. En _N-Queens_, la granularidad natural del _backtracking_ recursivo distribuye trabajo uniforme entre las goroutines del _pool_, logrando paridad estadística perfecta con una sola directiva donde la versión manual requiere gestión explícita de goroutines y sincronización.

*Rendimiento equivalente estadísticamente (MergeSort, Sections).* En _MergeSort_, ambas variantes paralelas logran un _speedup_ menor que $1 times$ porque la fase de _merge_ final, que consolida los resultados de los subárboles recursivos, es inherentemente secuencial y domina el tiempo total de ejecución conforme a la ley de Amdahl. En _Sections_, la alta varianza de ambas versiones paralelas (CV cercano al 22%) produce intervalos de confianza que se solapan ampliamente, de modo que la diferencia observada no tiene respaldo estadístico: las tres secciones de trabajo son de duración corta y el overhead de goroutines domina sobre el beneficio del paralelismo en ambos casos.

*GompherMP cede rendimiento al manual (PrefixSum, MonteCarlo, QuickSort, Fibonacci).* Las causas raíz en este grupo son cuatro limitaciones implementativas identificadas durante la evaluación:

En primer lugar, _PrefixSum_ expone el costo de la función `getGoroutineID()`, que utiliza `runtime.Stack()` para extraer el identificador de la goroutine actual parseando el _stack trace_. Esta operación, invocada en cada llamada a `Barrier()` y `Single()`, introduce una latencia de entre 1 y 5 µs por llamada que acumula un overhead significativo en _benchmarks_ con sincronización frecuente. Una solución futura consiste en reemplazar esta estrategia por un almacenamiento local de goroutine o por un mapa indexado por canal.

En segundo lugar, _MonteCarlo_ evidencia la ausencia de una variable equivalente a `omp_get_thread_num()` dentro de las directivas de bucle. Para obtener alto rendimiento en una simulación de Monte Carlo, cada goroutine debe inicializar su propio generador de números aleatorios con una semilla distinta, algo que la versión manual logra trivialmente almacenando el RNG localmente. GompherMP, al no exponer el identificador de hilo al cuerpo del bucle, fuerza al usuario a utilizar un RNG por iteración con el overhead asociado.

En tercer lugar, _QuickSort_ ilustra una limitación expresiva del constructo `taskloop`: esta directiva divide el espacio de iteración en bloques de tamaño fijo (_grainsize_), lo que no permite capturar la recursión paralela natural del algoritmo. La versión manual lanza goroutines recursivas que explotan toda la profundidad del árbol de recursión. Para expresar este patrón en GompherMP sería necesaria una combinación de `task` y `taskwait` que el motor de transformación actual no genera automáticamente a partir de `taskloop`.

En cuarto lugar, _Fibonacci_ muestra que el overhead de la función `TaskWithDepend` (que realiza una asignación dinámica de memoria y adquiere el _lock_ del registro de dependencias por cada tarea) domina completamente el tiempo de ejecución cuando el trabajo por tarea es de grano muy fino (submilisegundo). El alto CV del 65% en la variante `taskloop` refleja además contención esporádica en la inicialización del _pool_, cuya magnitud es proporcional al overhead relativo de gestión cuando el trabajo útil es trivial.

En síntesis, los resultados confirman que GompherMP es competitivo con la paralelización manual para cargas de trabajo _compute-bound_ con patrones regulares o tareas de granularidad media-alta, categoría que representa los casos de uso más frecuentes en HPC de propósito general. Las brechas de rendimiento observadas no son atribuibles al modelo de directivas en sí, sino a cuatro limitaciones concretas y acotadas de la implementación actual, cada una de las cuales tiene una solución técnica identificada. Este resultado valida la hipótesis central del proyecto: el paradigma de paralelismo basado en directivas es viable en el ecosistema de Go y puede ofrecer rendimiento comparable e incluso superior al del código paralelo manual idiomático del lenguaje.
