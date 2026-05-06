
= Capítulo 1. Generalidades


== 1.1. Problemática

En la era de la computación multinúcleo, el aprovechamiento del paralelismo se ha convertido en un pilar fundamental para el desarrollo de software de alto rendimiento. El lenguaje de programación Go ha ganado una notable popularidad gracias a su innovador modelo de concurrencia, ideal para aplicaciones de red y servicios escalables (Madridejos Zamorano, 2015). No obstante, la aplicación de este modelo al paralelismo de cómputo intensivo, cuyo dominio tradicional de estándares como OpenMP se encuentran en C++ y Fortran, presenta desafíos significativos que limitan la productividad del desarrollador y la eficiencia del código. Este trabajo emplea la técnica de árbol de problemas para sistematizar y abordar esta discordancia de paradigmas, identificando los problemas causa, el problema central y los problemas efecto.


=== 1.1.1. Árbol de problemas

El árbol de problemas es una herramienta analítica que permite elaborar un diagrama de causas y efectos entre los distintos problemas identificados para ofrecer una visión parcial y jerarquizada de la realidad. La construcción de este diagrama se basa en determinar un problema central, para luego ordenar en torno a él sus causas en la parte inferior y sus efectos o consecuencias en la parte superior (Camacho et al., 2001).

En la tabla 1 que se presenta a continuación, se muestran los problemas causa, el problema central y los problemas de efecto planteados para el proyecto.

#figure(
  table(
    columns: 4,
    stroke: 0.5pt,
    table.cell(fill: luma(230))[*PROBLEMAS EFECTOS*], table.cell(fill: luma(230))[*Aumento de barreras para la optimización y escalabilidad del código.*], table.cell(fill: luma(230))[*Incremento de carga cognitiva y tasa de errores del programador.*], table.cell(fill: luma(230))[*Ausencia de herramientas de alto nivel para el paralelismo de memoria compartida en el ecosistema de Go.*],
    [PROBLEMA CENTRAL], [Existe fricción semántico-sintáctica al implementar algoritmos paralelos computacionalmente intensivos basados en el paradigma de memoria compartida en lenguajes multiparadigma modernos como Go.],
    [PROBLEMAS CAUSA], [Existen limitaciones expresivas inherentes a las primitivas de concurrencia nativas de los lenguajes de propósito general como Go.], [Al implementar algoritmos paralelos basados en memoria compartida en Go o lenguajes multiparadigma modernos, existe acoplamiento entre lógicas del algoritmo y mecanismos de concurrencia.], [Incertidumbre sobre la viabilidad de abstracciones de paralelismo de alto nivel para memoria compartida en lenguajes multiparadigma modernos como Go.],
  ),
  caption: [Tabla 1. Esquema del árbol de problemas],
) <tab:tabla-1-esquema-del-árbol-de-p>


=== 1.1.2. Descripción

El lenguaje de programación Go se ha consolidado como una herramienta potente para el desarrollo de sistemas concurrentes, demostrando su utilidad para la programación paralela en arquitecturas multi-núcleo (Madridejos Zamorano, 2015). No obstante, su modelo nativo, basado en goroutines y channels, no se traduce directamente a los patrones de paralelismo de cómputo intensivo tradicionalmente abordados por estándares como OpenMP en C++ o Fortran. Esta divergencia de paradigmas genera una fricción semántico-sintáctica para los desarrolladores que buscan implementar algoritmos de High-Performance Computing (HPC) en Go, obligándolos a reestructurar lógicas y a gestionar manualmente complejidades que en otros ecosistemas son abstraídas.

Una de las causas fundamentales de esta fricción es el acoplamiento entre la lógica del algoritmo y los mecanismos de concurrencia al usar las primitivas nativas de Go. La implementación manual del paralelismo mediante goroutines, WaitGroups y channels obliga al programador a entrelazar los detalles de la sincronización y el reparto de trabajo directamente en el cuerpo del algoritmo (Madridejos Zamorano, 2015). Este enfoque se aleja del principio de separación de incumbencias, resultando en un código más verboso y difícil de mantener en comparación con los modelos basados en directivas, donde el código paralelo se mantiene estructuralmente muy cercano a su versión secuencial (Kambites et al., 2001; Vikas et al., 2014).

Adicionalmente, existen limitaciones expresivas inherentes a las primitivas de concurrencia nativas de los lenguajes de propósito general. Aunque potentes, estas primitivas son de bajo nivel y no ofrecen abstracciones directas para patrones de HPC comunes. Por ejemplo, lenguajes como Java históricamente han carecido de mecanismos de sincronización de alto rendimiento, como las barreras rápidas, que son cruciales para la eficiencia de muchos algoritmos paralelos (Kambites et al., 2001). De manera similar, modelos como Fork/Join, aunque conceptualmente análogos a la división de tareas, son difíciles de aplicar a cómputos que no siguen un patrón estricto de "divide y vencerás", especialmente cuando existen dependencias de datos complejas (Yoshida et al., 2017).

Finalmente, esta situación genera una incertidumbre sobre la viabilidad y el rendimiento de abstracciones de paralelismo de alto nivel en el ecosistema de Go. La falta de una implementación de referencia de un estándar como OpenMP en Go crea un vacío en la literatura y en las herramientas disponibles, dejando a los desarrolladores sin una alternativa probada para el desarrollo de aplicaciones de HPC. Trabajos pioneros como la tesis de Madridejos Zamorano (2015) comenzaron a explorar esta área, y la investigación activa en otros lenguajes modernos como Zig (Kacs et al., 2024) demuestra que la adaptación de estos paradigmas sigue siendo un campo de estudio relevante y no resuelto. Como consecuencia directa de esta fricción, surgen efectos adversos que impactan tanto al desarrollador como al software. Principalmente, se incrementa la carga cognitiva y la tasa de errores del programador, ya que la gestión manual del paralelismo es un proceso difícil que a menudo requiere reestructuraciones considerables del código (Vikas et al., 2014). Esto eleva el riesgo de introducir errores sutiles como condiciones de carrera o interbloqueos (deadlocks) (Powers & Alaghband, 2007). A su vez, el acoplamiento entre la lógica del algoritmo y los mecanismos de concurrencia genera barreras para la optimización y escalabilidad, pues el código resultante es más difícil de mantener y adaptar (Vikas et al., 2014). Esta ineficiencia en la implementación tiene implicaciones directas en los límites teóricos del rendimiento: según la Ley de Amdahl, la aceleración máxima (speedup) de un sistema está acotada por la fracción del código que debe ejecutarse secuencialmente (Hennessy & Patterson, 2012). La fricción semántico-sintáctica descrita propicia que el programador, por complejidad, deje secciones sin paralelizar o utilice sincronizaciones ineficientes, aumentando artificialmente dicha fracción secuencial y alejando el rendimiento real del potencial teórico del hardware. En conjunto, estos efectos justifican la necesidad de una investigación rigurosa para establecer una solución robusta y eficiente para Go


=== 1.1.3. Problema Seleccionado

El problema central que aborda esta investigación es la fricción semántico-sintáctica que enfrentan los desarrolladores al implementar algoritmos de cómputo intensivo basados en el paradigma de memoria compartida en Go. Esta fricción surge porque el modelo de concurrencia nativo de Go, aunque excelente para servicios de red, no ofrece las abstracciones de alto nivel necesarias para el dominio de HPC. Como resultado, los programadores se ven forzados a implementar la paralelización de forma manual, acoplando la lógica del algoritmo con mecanismos de concurrencia de bajo nivel, lo que incrementa la carga cognitiva, aumenta la probabilidad de errores (como condiciones de carrera o deadlocks) y dificulta la optimización y el mantenimiento del código. La ausencia de un estándar declarativo como OpenMP en el ecosistema de Go crea una barrera para la productividad y la adopción del lenguaje en la comunidad de computación de alto rendimiento, un vacío que este proyecto busca resolver mediante la propuesta de la herramienta GompherMP. Para abordar este problema, se espera como resultado principal la herramienta en estado funcional, un compilador source to source con su correspondiente librería de runtime. El método a emplear consistirá en la transformación de código basada en directivas para automatizar la paralelización. La efectividad de la herramienta se validará mediante benchmarks, comparando su rendimiento y la expresividad del código resultante frente a implementaciones paralelas manuales en Go.


== 1.2. Objetivos

En esta sección, se presentan tanto el objetivo general como los objetivos específicos. El objetivo general se enfoca a resolver el problema seleccionado, mientras que los objetivos específicos guardan relación con los problemas causa mencionados en el árbol de problemas diagramado anteriormente.


=== 1.2.1. Objetivo General

Diseñar, desarrollar y evaluar la herramienta GompherMP que implemente un modelo de paralelismo de alto nivel basado en directivas para el lenguaje Go y que está inspirado en un subconjunto del estándar OpenMP, a fin de cerrar la brecha entre la productividad del desarrollador y el rendimiento del código en la implementación de algoritmos de cómputo intensivo.


=== 1.2.2. Objetivos Específicos

OE1. Diseñar y especificar la arquitectura de la herramienta GompherMP, incluyendo la sintaxis de sus directivas y cláusulas, los algoritmos de transformación de código y la interfaz de línea de comandos (CLI).

OE2. Implementar la infraestructura central de GompherMP, integrando un compilador source-to-source para la transformación de directivas y una librería de runtime para la orquestación de la concurrencia y sincronización en Go.

OE3. Evaluar la herramienta GompherMP mediante la ejecución de benchmarks, comparando sus resultados en términos de rendimiento y expresividad del código frente a las implementaciones secuenciales y paralelas manuales en Go.


=== 1.2.3. Resultados Esperados

OE1. Diseñar y especificar la arquitectura de la herramienta GompherMP, incluyendo la sintaxis de sus directivas y cláusulas, los algoritmos de transformación de código y la interfaz de línea de comandos (CLI).

R1. Documento de especificación de directivas y cláusulas.

R2. Diseño de la arquitectura de la herramienta.

R3. Especificación funcional de la Interfaz de Línea de Comandos (CLI).

OE2. Implementar la infraestructura central de GompherMP, integrando un compilador source-to-source para la transformación de directivas y una librería de runtime para la orquestación de la concurrencia y sincronización en Go.

R4. Módulo de gestión de goroutines y reparto de trabajo implementado y probado.

R5. Módulo de mecanismos de sincronización implementado y probado.

R6. Módulo de soporte para paralelismo de tareas (tasking) implementado y probado.

R7. Analizador sintáctico (Parser) de directivas GompherMP.

R8. Motor de transformación del AST implementado.

R9. Herramienta GompherMP (CLI) funcional.

OE3. Evaluar la herramienta GompherMP mediante la ejecución de benchmarks, comparando sus resultados en términos de rendimiento y expresividad del código frente a las implementaciones secuenciales y paralelas manuales en Go.

R10. Suite de benchmarks implementada.

R11. Informe de evaluación de rendimiento y escalabilidad.

R12. Reporte de análisis comparativo sobre expresividad y productividad.


=== 1.2.4. Mapeo de problemática con objetivos

A través de la siguiente tabla, se pretende mostrar cómo cada objetivo específico se relaciona con cada uno de los problemas causa del árbol de problemas.

#figure(
  table(
    columns: 2,
    stroke: 0.5pt,
    table.cell(fill: luma(230))[*Problema causa*], table.cell(fill: luma(230))[*Objetivo Específico*],
    [1. Existen limitaciones expresivas inherentes a las primitivas de concurrencia en lenguajes de propósito general modernos como Go], [OE1. Diseñar y especificar la arquitectura de la herramienta GompherMP,],
    [2. Existe acoplamiento entre lógica y mecanismos de concurrencia al momento de implementar algoritmos paralelos basados en memoria compartida en lenguajes de propósito general modernos como Go], [OE2. Implementar la librería de runtime de GompherMP, que proveerá las funciones de soporte para el código paralelo generado.],
    [3. Es incierto qué tan viables son las abstracciones de alto nivel para memoria compartida en lenguajes de propósito general modernos como Go], [OE3. Evaluar la herramienta GompherMP mediante la ejecución de benchmarks.],
  ),
  caption: [Tabla 2. Relación entre problemas causa y objetivos específicos],
) <tab:tabla-2-relación-entre-problem>


=== 1.2.5. Mapeo de objetivos, resultados y medios de verificación

En esta sección, se pretende mostrar cómo cada objetivo específico se traduce en resultados esperados juntos con sus medios de verificación e indicadores objetivamente verificables. Esto permitirá que se realice una evaluación adecuada y precisa en el transcurso del proyecto.

#figure(
  table(
    columns: 3,
    stroke: 0.5pt,
    table.cell(fill: luma(230))[*Objetivo 1: Diseñar y especificar la arquitectura de la herramienta GompherMP, incluyendo la sintaxis de sus directivas y cláusulas, los algoritmos de transformación de código y la interfaz de línea de comandos (CLI).*],
    [Resultado], [Medio de verificación], [Indicador objetivamente verificable (IOV)],
    [R1. Documento de especificación de directivas y cláusulas.], [- Informe técnico de especificación de GompherMP.], [- El informe define la sintaxis, semántica y presenta al menos un ejemplo de uso para cada directiva y cláusula del alcance del proyecto (parallel, for, task, depend, critical, etc.). \\ - Se obtiene la aprobación escrita de un experto en programación concurrente.],
    [R2. Diseño de la arquitectura de la herramienta.], [- Documento de diseño de arquitectura.], [- El documento contiene diagramas que ilustran la interacción entre el compilador source-to-source y la librería de runtime. \\ - Existe una descripción del flujo general del proceso de transformación del AST. \\ - Se obtiene la aprobación escrita de la arquitectura por parte de un experto en programación concurrente.],
    [R3. Especificación funcional de la Interfaz de Línea de Comandos (CLI).], [- Manual de usuario de la CLI.], [- El manual especifica todos los comandos, argumentos y opciones disponibles.  \\ - Incluye ejemplos claros de uso para la transpilación de código. \\ - Se obtiene la aprobación escrita de un experto en programación concurrente.],
  ),
  caption: [Tabla 3. Resultados esperados, medios de verificación e indicadores objetivamente verificables (IOV) para el objetivo específico OE1],
) <tab:tabla-3-resultados-esperados,>

#figure(
  table(
    columns: 3,
    stroke: 0.5pt,
    table.cell(fill: luma(230))[*Objetivo 2: Implementar la infraestructura central de GompherMP, integrando un compilador source-to-source para la transformación de directivas y una librería de runtime para la orquestación eficiente de la concurrencia y sincronización en Go.*],
    [Resultado], [Medio de verificación], [Indicador objetivamente verificable (IOV)],
    [R4. Módulo de gestión de goroutines y reparto de trabajo implementado y probado.], [- Código fuente en repositorio en línea.  \\ - Informe de pruebas unitarias.], [- Repositorio accesible con código fuente documentado. \\ - El código implementa la gestión del pool de goroutines y el reparto de iteraciones para la directiva for.  \\ - El informe evidencia una cobertura de pruebas del 100% para el módulo.],
    [R5. Módulo de mecanismos de sincronización implementado y probado.], [- Código fuente en repositorio en línea.  \\ - Informe de pruebas unitarias.], [- El código fuente implementa la funcionalidad para las directivas critical y single. \\ - El informe evidencia una cobertura de pruebas del 100% para el módulo.],
    [R6. Módulo de soporte para paralelismo de tareas (tasking) implementado y probado.], [- Código fuente en repositorio en línea.  \\ - Informe de pruebas de integración.], [- El código implementa la creación de tareas y la gestión de su grafo de dependencias (in, out, inout).  \\ - Las pruebas validan la correcta ejecución de tareas según las dependencias definidas. \\ - El informe evidencia una cobertura de pruebas del 100% para el módulo.],
    [R7. Analizador sintáctico (Parser) de directivas GompherMP.], [- Código fuente del compilador en repositorio en línea. \\ - Informe de pruebas de integración.], [- El módulo de código es capaz de identificar y extraer correctamente todas las directivas y cláusulas definidas en la especificación (O1.R1) a partir de los comentarios //gompher. \\ - El informe evidencia una cobertura de pruebas del 100% para el módulo.],
    [R8. Motor de transformación del AST implementado.], [- Código fuente del compilador en repositorio en línea. \\ - Informe de pruebas de integración.], [- El código modifica el AST de un programa de entrada, inyectando las llamadas a la librería de runtime correspondientes a las directivas encontradas. \\ - El informe evidencia una cobertura de pruebas del 100% para el módulo.],
    [R9. Herramienta GompherMP (CLI) funcional.], [- Ejecutable de la herramienta CLI.  \\ - Guía de instalación y uso.], [- La herramienta transpila correctamente archivos .go con directivas GompherMP, generando como salida un archivo .go con código concurrente nativo, el cual debe ser compilable con el compilador estándar de Go. \\ - Un experto en programación concurrente valida el funcionamiento respecto a su especificación funcional],
  ),
  caption: [Tabla 4. Resultados esperados, medios de verificación e indicadores objetivamente verificables (IOV) para el objetivo específico OE2],
) <tab:tabla-4-resultados-esperados,>

#figure(
  table(
    columns: 3,
    stroke: 0.5pt,
    table.cell(fill: luma(230))[*Objetivo 3: Evaluar la herramienta GompherMP mediante la ejecución de benchmarks, comparando sus resultados en términos de rendimiento y expresividad del código frente a las implementaciones secuenciales y paralelas manuales en Go.*],
    [Resultado], [Medio de verificación], [Indicador objetivamente verificable (IOV)],
    [R10. Suite de benchmarks implementada.], [- Código fuente de los algoritmos de benchmark en repositorio en línea.], [- El repositorio contiene al menos 3 algoritmos de cómputo intensivo.  \\ - Cada algoritmo está implementado en 3 versiones: secuencial, paralela manual y paralela con GompherMP. \\ - Se obtiene la aprobación escrita de un experto en programación concurrente.],
    [R11. Informe de evaluación de rendimiento y escalabilidad.], [- Informe de evaluación de rendimiento.], [- El informe presenta los resultados de tiempo de ejecución, speedup y eficiencia para cada versión de los benchmarks.  \\ - El informe incluye gráficos comparativos que visualizan los resultados. \\ - Se presentan conclusiones respecto a la efectividad de la herramienta en el informe. \\ - El informe debe ser aprobado por un experto en programación concurrente.],
    [R12. Reporte de análisis comparativo sobre expresividad y productividad.], [- Reporte de análisis comparativo.], [- El reporte incluye un análisis cuantitativo (ej. líneas de código) y cualitativo de la complejidad y legibilidad del código entre la versión paralela manual y la versión con GompherMP. \\ - Se presentan conclusiones sobre el impacto de GompherMP en la productividad del desarrollador.],
  ),
  caption: [Tabla 5. Resultados esperados, medios de verificación e indicadores objetivamente verificables (IOV) para el objetivo específico OE3],
) <tab:tabla-5-resultados-esperados,>


== 1.3. Metodología

A continuación, se detallan las herramientas, métodos y procedimientos necesarios para construir los resultados esperados de esta tesis. Para cada objetivo específico, se presenta una tabla que conecta los resultados con las técnicas y recursos que permitirán alcanzarlos, junto con una descripción de cada uno de ellos.

#figure(
  table(
    columns: 2,
    stroke: 0.5pt,
    table.cell(fill: luma(230))[*Objetivo 1: Diseñar y especificar la arquitectura de la herramienta GompherMP, incluyendo la sintaxis de sus directivas y cláusulas, los algoritmos de transformación de código y la interfaz de línea de comandos (CLI).*],
    [Resultado], [Herramientas, métodos y procedimientos],
    [R1. Documento de especificación de directivas y cláusulas.], [Herramienta: LaTeX],
    [R2. Diseño de la arquitectura de la herramienta.], [Herramientas: Excalidraw, LaTeX \\ Método: Arquitectura centrada en componentes],
    [R3. Especificación funcional de la Interfaz de Línea de Comandos (CLI).], [Herramienta: LaTeX],
  ),
  caption: [Tabla 6. Herramientas , métodos y procedimientos relacionados al objetivo específico 1 y sus resultados esperados],
) <tab:tabla-6-herramientas-,-métodos>

#figure(
  table(
    columns: 2,
    stroke: 0.5pt,
    table.cell(fill: luma(230))[*Objetivo 2: Implementar la infraestructura central de GompherMP, integrando un compilador source-to-source para la transformación de directivas y una librería de runtime para la orquestación eficiente de la concurrencia y sincronización en Go.*],
    [Resultado], [Herramientas, métodos y procedimientos],
    [R4. Módulo de gestión de goroutines y reparto de trabajo implementado y probado.], [Herramientas: Go, Neovim, VScode, Github \\ Métodos: Metodología Kanban, TDD \\ Procedimientos: Pruebas unitarias],
    [R5. Módulo de mecanismos de sincronización implementado y probado.], [Herramientas: Go, Neovim, VScode, Github \\ Métodos: Metodología Kanban, TDD \\ Procedimientos: Pruebas unitarias],
    [R6. Módulo de soporte para paralelismo de tareas (tasking) implementado y probado.], [Herramientas: Go, Neovim, VScode, Github \\ Métodos: Metodología Kanban, TDD \\ Procedimientos: Pruebas unitarias],
    [R7. Analizador sintáctico (Parser) de directivas GompherMP.], [Herramientas: Go, Neovim, VScode, Github, Go/AST, Go/test \\ Métodos: Metodología kanban, TDD, análisis sintáctico \\ Procedimientos: Pruebas unitarias],
    [R8. Motor de transformación del AST implementado.], [Herramientas: Go, Neovim, VScode, Github, Go/AST, Go/test \\ Métodos: Metodología kanban, TDD, transformación de abstract syntax trees \\ Procedimientos: Pruebas unitarias],
    [R9. Herramienta GompherMP (CLI) funcional.], [Herramientas: Go, Neovim, VScode, Github, Cobra \\ Métodos: Metodología kanban, TDD \\ Procedimientos: Pruebas unitarias, pruebas de integración],
  ),
  caption: [Tabla 7. Herramientas , métodos y procedimientos relacionados al objetivo específico 2 y sus resultados esperados],
) <tab:tabla-7-herramientas-,-métodos>

#figure(
  table(
    columns: 2,
    stroke: 0.5pt,
    table.cell(fill: luma(230))[*Objetivo 3: Evaluar la herramienta GompherMP mediante la ejecución de benchmarks, comparando sus resultados en términos de rendimiento y expresividad del código frente a las implementaciones secuenciales y paralelas manuales en Go.*],
    [Resultado], [Herramientas, métodos y procedimientos],
    [R10. Suite de benchmarks implementada.], [Herramientas: Go, C, OpenMP \\ Métodos: Benchmarking de algoritmos de computación intensiva \\ Procedimiento: Performance profiling],
    [R11. Informe de evaluación de rendimiento y escalabilidad.], [Herramientas: Python, Matplotlib, LaTeX \\ Métodos: Análisis comparativo con pruebas estadísticas, análisis cualitativo \\ Procedimiento: ANOVA, t-test],
    [R12. Reporte de análisis comparativo sobre expresividad y productividad.], [Herramientas: Python, Scipy, LaTeX \\ Métodos: Análisis comparativo con pruebas estadísticas, análisis cualitativo \\ Procedimiento: ANOVA, t-test],
  ),
  caption: [Tabla 8. Herramientas , métodos y procedimientos relacionados al objetivo específico 3 y sus resultados esperados],
) <tab:tabla-8-herramientas-,-métodos>


=== 1.3.1. Herramientas


==== 1.3.1.1 LaTeX

Según The LaTeX Project (s.f.), LaTeX es un sistema de preparación de documentos de alta calidad tipográfica diseñado para la publicación científica y técnica, que permite a los autores enfocarse en el contenido en lugar de la apariencia. Para este proyecto, resulta ideal en la elaboración de documentos como el informe de evaluación de rendimiento y el diseño de la arquitectura, ya que su capacidad para gestionar automáticamente referencias, figuras y tablas asegura la consistencia y profesionalismo requeridos para presentar los resultados.


==== 1.3.1.2 Excalidraw

Según Excalidraw (s.f.), Excalidraw es una pizarra virtual para esbozar diagramas con aspecto de dibujo a mano. Para este proyecto, será una herramienta principal en la fase de diseño, específicamente para crear los diagramas de arquitectura requeridos. Su enfoque en la simplicidad facilitará la ilustración de la interacción entre el compilador source-to-source y la librería de runtime, así como el flujo de transformación del AST.


==== 1.3.1.3 Go

Según el equipo de Go (s.f.), Go es un lenguaje de programación de código abierto, apoyado por Google, que facilita la construcción de software simple, seguro y escalable, destacando por su concurrencia nativa y una robusta librería estándar. En el contexto de este proyecto, Go es el lenguaje objetivo sobre el cual se construye la herramienta GompherMP; el compilador source-to-source propuesto analizará código Go con directivas y lo transformará para generar código concurrente nativo, utilizando las primitivas del lenguaje como goroutines y channels para implementar el paralelismo.


==== 1.3.1.4 Neovim

Según Neovim (s.f.), Neovim es un editor de texto de la familia Vim, enfocado en la extensibilidad y la usabilidad, que puede ser utilizado desde un terminal o como una aplicación gráfica independiente. Para el desarrollo de este proyecto, Neovim será un entorno de codificación donde se implementará tanto el compilador source-to-source como la librería de runtime de GompherMP. Su naturaleza ligera y configurable permitirá un flujo de trabajo eficiente para escribir, probar y depurar el código Go que compone la herramienta.


==== 1.3.1.5 VSCode

Según Microsoft (s.f.), Visual Studio Code es un editor de código fuente ligero pero potente que se ejecuta en el escritorio y está disponible para Windows, macOS y Linux. Para el desarrollo de este proyecto, Visual Studio Code será otro entorno de codificación utilizado para colaborar en la implementación de la herramienta. Sus robustas capacidades de depuración, la integración con Git y el amplio ecosistema de extensiones facilitarán su contribución en el desarrollo tanto del compilador como de la librería de runtime.


==== 1.3.1.6 Github

Según GitHub (s.f.), GitHub es la plataforma de desarrollo completa para construir, escalar y entregar software de forma segura. Para este proyecto, servirá como el repositorio de código fuente central para la herramienta GompherMP, facilitando el control de versiones y la colaboración durante la implementación del compilador y la librería de runtime. Además, el repositorio en línea actuará como el medio de verificación principal para los entregables de código, tal como se especifica en los objetivos específicos 2 y 3 del proyecto.


==== 1.3.1.7 Go/AST

Según el equipo de Go (s.f.), el paquete ast declara los tipos utilizados para representar árboles de sintaxis abstracta para archivos de código Go. En este proyecto, este paquete es fundamental para el compilador source-to-source, ya que su función principal es analizar el código fuente, interpretar las directivas y transformar el Árbol de Sintaxis Abstracta (AST) para generar código Go nativo concurrente. La capacidad de manipular directamente la estructura del código a través del AST es, por lo tanto, esencial para cumplir con el objetivo específico 3 del proyecto.


==== 1.3.1.8 Go/test

Según el equipo de Go (s.f.), el paquete testing provee el soporte para pruebas automatizadas de paquetes de Go, siendo la base para la creación de tests unitarios y benchmarks. Dentro de este proyecto, será una herramienta crucial para cumplir con los objetivos de evaluación, ya que se utilizará para implementar las pruebas unitarias y de integración que validarán la correctitud de los módulos de la librería de runtime y del compilador (Objetivo específico 2). Asimismo, será fundamental para desarrollar la suite de benchmarks necesaria para medir el rendimiento y la escalabilidad de GompherMP, tal como lo exige el objetivo específico 3.


==== 1.3.1.9 Cobra

Según Steve Francia (s.f.), Cobra es tanto una librería para crear potentes aplicaciones CLI modernas como un programa para generar aplicaciones y archivos de comandos. En el contexto de este proyecto, Cobra será la base para construir la Interfaz de Línea de Comandos (CLI) de la herramienta GompherMP, tal como se especifica en los objetivos 1 y 2. Su uso simplificará la implementación de los comandos, argumentos y opciones necesarios para que el usuario pueda transpilar su código Go de manera sencilla.


==== 1.3.1.10 C

Según C-Language.org (s.f.), C es un lenguaje de programación de propósito general conocido por su rendimiento y su capacidad para operar a bajo nivel. En el marco de esta investigación, se utilizará para desarrollar implementaciones de referencia para los algoritmos de la suite de benchmarks, como se detalla en el objetivo específico 3. Dado que los estándares de paralelismo como OpenMP tienen su dominio tradicional en lenguajes de sistemas, estas versiones en C servirán como una línea base de rendimiento contra la cual se podrán contrastar los resultados de las implementaciones en Go, enriqueciendo así el informe de evaluación de rendimiento y escalabilidad.


==== 1.3.1.11 OpenMP

Según la OpenMP Architecture Review Board (s.f.), OpenMP es una API que soporta la programación paralela de memoria compartida en C, C++ y Fortran mediante un conjunto de directivas de compilador. Este estándar es la inspiración fundamental para el presente proyecto, ya que la herramienta GompherMP busca adaptar un subconjunto de su modelo de paralelismo basado en directivas al ecosistema de Go. Por lo tanto, la especificación de OpenMP proporciona el marco conceptual para el diseño de las directivas y cláusulas de GompherMP, sirviendo como referente para evaluar la expresividad y funcionalidad de la solución propuesta.


==== 1.3.1.12 Python

Según la Python Software Foundation (s.f.), Python es un lenguaje de programación que permite trabajar rápidamente e integrar sistemas de manera más efectiva. En el contexto de esta investigación, se utilizará para desarrollar versiones de los algoritmos de la suite de benchmarks, permitiendo una comparación de rendimiento y expresividad frente a la solución propuesta en Go, similar a como se plantea para C. Adicionalmente, se empleará para la automatización de la ejecución de pruebas y la generación de gráficos para el informe de evaluación, facilitando el análisis comparativo requerido en el objetivo específico 3.


==== 1.3.1.13 Matplotlib

Según The Matplotlib Development Team (s.f.), Matplotlib es una librería completa para crear visualizaciones estáticas, animadas e interactivas en Python. En el marco de este proyecto, esta librería será fundamental para cumplir con los requisitos del objetivo específico 3, ya que se empleará para generar los gráficos comparativos que visualizarán los resultados de rendimiento y escalabilidad. Estos gráficos, que mostrarán métricas como el tiempo de ejecución y el speedup de las versiones secuencial, manual y con GompherMP, son un componente esencial del informe de evaluación de rendimiento.


==== 1.3.1.14 Scipy

Según la comunidad de SciPy (s.f.), SciPy es una librería de Python que proporciona algoritmos y rutinas numéricas eficientes para optimización, álgebra lineal y estadística. En el marco de esta investigación, se utilizará para realizar las pruebas estadísticas sobre los datos de rendimiento obtenidos de la suite de benchmarks. Esto permitirá validar con rigor si las diferencias en métricas como el speedup y el tiempo de ejecución son significativas, fortaleciendo las conclusiones del informe de evaluación de rendimiento.


=== 1.3.2. Métodos


==== 1.3.2.1 Arquitectura centrada en componentes

Según Bass et al. (2012), una arquitectura centrada en componentes se enfoca en la descomposición del sistema en unidades funcionales o lógicas con interfaces bien definidas. Este enfoque será aplicado en el diseño de GompherMP, separando el compilador source-to-source y la librería de runtime como componentes distintos. Esta separación modular facilitará el desarrollo, las pruebas y el mantenimiento independiente de cada parte, asegurando que la interacción entre ellos para la transformación y ejecución del código sea clara y cohesiva.


==== 1.3.2.2 Metodología Kanban

De acuerdo con Anderson (2010), Kanban es una metodología para gestionar el flujo de trabajo que se centra en la visualización del trabajo, la limitación del trabajo en progreso y la maximización de la eficiencia. Para la gestión de este proyecto, se utilizará un tablero Kanban para organizar y priorizar las tareas asociadas a los cuatro objetivos específicos, desde el diseño de la arquitectura hasta la evaluación final, permitiendo un seguimiento transparente del avance y una adaptación continua a los desafíos que surjan durante el desarrollo.


==== 1.3.2.3 Desarrollo guiado por pruebas (TDD)

Según Beck (2003), el Desarrollo Guiado por Pruebas es una práctica de desarrollo de software donde los tests se escriben antes que el código que los debe pasar. Esta metodología se aplicará en la implementación de la librería de runtime de GompherMP, asegurando que cada función de sincronización y gestión de goroutines sea robusta y correcta desde su concepción. Este enfoque es fundamental para cumplir con los indicadores de cobertura de pruebas superiores al 80% estipulados en el objetivo específico 2.


==== 1.3.2.4 Análisis Léxico y Sintáctico

Este método constituye la fase inicial o frontend de la solución propuesta, siguiendo el modelo de fases secuenciales descrito por Aho et al. (2007). El procedimiento comenzará con un análisis léxico (scanning), encargado de leer el flujo de caracteres del código fuente Go y agruparlos en tokens significativos (lexemas). Posteriormente, se ejecutará el análisis sintáctico, que utilizará estos tokens para construir una representación jerárquica de la estructura gramatical. Este método es fundamental para el objetivo específico 2, ya que permitirá identificar y extraer correctamente las directivas //gompher y las estructuras de control asociadas, generando el insumo necesario para la etapa de transformación.


==== 1.3.2.5 Compilación Source-to-Source y Transformación de AST

De acuerdo con Aho et al. (2007) y Cooper & Torczon (2012), la transformación de un Árbol de Sintaxis Abstracta (AST) implica modificar la representación estructural del código fuente preservando su semántica. Este método constituye el núcleo de la arquitectura source-to-source de GompherMP para cumplir con el objetivo específico 2.

El procedimiento consiste en recorrer el AST generado en la fase previa (parsing), identificar los nodos marcados por las directivas y reescribir su estructura inyectando las llamadas a la librería de runtime (goroutines y canales). A diferencia de una compilación tradicional a código máquina, este enfoque finaliza con una etapa de síntesis de código, donde el AST transformado se serializa nuevamente en archivos de texto .go. Esto garantiza que el código resultante sea nativo, legible y portable, pudiendo ser compilado por el toolchain estándar de Go sin modificaciones.


==== 1.3.2.6 Benchmarking de algoritmos de computación intensiva

Según Jain (1991), el benchmarking es el proceso de ejecutar un programa para evaluar su rendimiento relativo de forma cuantitativa. Este método será la base para la evaluación de GompherMP (objetivo específico 3), donde se implementará una suite de benchmarks con al menos tres algoritmos de cómputo intensivo. La ejecución de estos algoritmos en sus versiones secuencial, paralela manual y con GompherMP permitirá medir y comparar métricas clave como el tiempo de ejecución y el speedup.


==== 1.3.2.7 Análisis comparativo con pruebas estadísticas

De acuerdo con Thiel (2014), un análisis comparativo utiliza métodos estadísticos para determinar si las diferencias observadas entre los resultados de dos o más grupos son significativas o si podrían haber ocurrido por azar. Este método se empleará para analizar los datos de rendimiento obtenidos del benchmarking. Mediante pruebas estadísticas, se validará si la mejora de rendimiento de GompherMP es estadísticamente significativa en comparación con las versiones secuencial y manual, fortaleciendo las conclusiones del informe de evaluación.


==== 1.3.2.8 Análisis cualitativo

Según Miles et al. (2014), el análisis cualitativo se enfoca en datos no numéricos para comprender conceptos y experiencias. Este método se utilizará para evaluar la expresividad y la productividad, como se exige en el resultado R12 del objetivo 3. Se realizará un análisis comparativo de la complejidad y legibilidad del código entre la versión paralela manual y la versión con GompherMP, apoyado en métricas como las líneas de código (LoC), para concluir sobre el impacto de la herramienta en la productividad del desarrollador.


=== 1.3.3. Procedimientos


==== 1.3.3.1 Pruebas unitarias

Según Sommerville (2011), las pruebas unitarias se centran en verificar la funcionalidad de componentes individuales de un programa de forma aislada. Este procedimiento será aplicado sistemáticamente sobre los módulos de la librería de runtime, como el gestor del pool de goroutines y los mecanismos de sincronización. El objetivo es asegurar que cada función se comporte como se espera antes de integrarla en el sistema completo, cumpliendo con los indicadores de cobertura de pruebas.


==== 1.3.3.2 Pruebas de integración

De acuerdo con Sommerville (2011), las pruebas de integración tienen como objetivo descubrir defectos en las interfaces y las interacciones entre componentes integrados. Este procedimiento se utilizará para validar que el compilador source-to-source y la librería de runtime de GompherMP funcionan correctamente en conjunto. Se verificarán casos de uso completos, desde la transpilación de un archivo .go con directivas hasta la correcta ejecución del código concurrente generado.


==== 1.3.3.3 Performance profiling

Según el equipo de Go (s.f.), el performance profiling es el análisis del software para identificar cuellos de botella y optimizar su rendimiento. Este procedimiento se aplicará utilizando herramientas del ecosistema de Go, como pprof, sobre el código generado por GompherMP. El objetivo es analizar el comportamiento de la librería de runtime y los algoritmos de la suite de benchmarks para detectar posibles sobrecargas (overhead) y asegurar que la implementación sea lo más eficiente posible.


==== 1.3.3.4 ANOVA

Según Snedecor y Cochran (1989), el análisis de varianza (ANOVA) es una prueba estadística utilizada para determinar si existen diferencias estadísticamente significativas entre las medias de tres o más grupos independientes. Este procedimiento se empleará en el análisis de los resultados de los benchmarks para comparar simultáneamente los tiempos de ejecución de las versiones secuencial, paralela manual y con GompherMP de cada algoritmo, y así determinar si el método de paralelización tiene un efecto significativo en el rendimiento.


==== 1.3.3.5 Prueba t (t-test)

De acuerdo con Snedecor y Cochran (1989), la prueba t es un método estadístico que se utiliza para comparar las medias de dos grupos. En el marco de esta investigación, este procedimiento se usará para realizar comparaciones específicas después de un análisis ANOVA, por ejemplo, para determinar si la diferencia de rendimiento entre la versión paralela con GompherMP y la versión paralela manual es estadísticamente significativa, ofreciendo una validación más granular de la eficiencia de la herramienta.

