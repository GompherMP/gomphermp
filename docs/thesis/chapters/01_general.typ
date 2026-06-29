
= Capítulo 1. Generalidades

== Problemática

En la era de la computación multinúcleo, el aprovechamiento del paralelismo se ha consolidado como un requisito indispensable para el desarrollo de software con ciertos requerimientos de rendimiento. Para abordar esta necesidad, resulta de gran interés el paradigma de programación basado en directivas, un enfoque que proviene originalmente del área de la Computación de Alto Rendimiento (HPC) y que destaca por simplificar drásticamente la implementación del paralelismo y la sincronización. Al permitir que el desarrollador delegue la gestión de hilos y la coordinación de tareas al compilador, este modelo facilita una "extensión relativamente sencilla de aplicaciones secuenciales en aplicaciones paralelas" (Czarnul et al., 2020), abstrayendo la complejidad de las primitivas de sincronización de la plataforma subyacente (sistema operativo o hardware) y logrando una ejecución eficiente con un esfuerzo de programación significativamente reducido en comparación con métodos manuales.

Sin embargo, el estrecho acoplamiento de este paradigma a lenguajes de sistemas tradicionales como C, C++ y Fortran obliga a los desarrolladores a sacrificar las ventajas de la ingeniería de software moderna en favor del rendimiento. Esta dependencia de herramientas clásicas impide aprovechar la mayor expresividad y la gestión automática de recursos propia de los lenguajes de propósito general modernos; beneficios que, según Nanz y Furia (2015), permiten obtener soluciones más concisas y reducir significativamente la probabilidad de errores derivados de la complejidad técnica.

Entonces, la industria debe elegir entre la efectividad y simplicidad del paralelismo basado en directivas y los beneficios de usar lenguajes de programación modernos. El presente trabajo emplea la técnica de árbol de problemas para sistematizar y abordar esta brecha, identificando los problemas causa, el problema central y los problemas efecto derivados de esta limitación en el ecosistema de desarrollo actual.


=== Árbol de problemas

El árbol de problemas es una herramienta analítica que permite elaborar un diagrama de causas y efectos entre los distintos problemas identificados para ofrecer una visión parcial y jerarquizada de la realidad. La construcción de este diagrama se basa en determinar un problema central, para luego ordenar en torno a él sus causas en la parte inferior y sus efectos o consecuencias en la parte superior (Camacho et al., 2001).

En la Tabla 1 que se presenta a continuación, se muestran los problemas causa, el problema central y los problemas de efecto planteados para el proyecto.

#figure(
  table(
    columns: (0.75fr, ..(1fr,) * 3),
    stroke: 0.5pt,
    fill: (col, row) => if col == 0 { luma(230) },
    align: (col, row) => if col == 0 { center + horizon } else { left + top },
    [PROBLEMAS EFECTOS], [Dependencia de lenguajes tradicionales (C/C++) para garantizar el rendimiento, sacrificando la expresividad, la gestión automática de memoria y la seguridad intrínseca de los lenguajes modernos], [Aumento del esfuerzo de programación al paralelizar algoritmos manualmente en lenguajes modernos, lo que desincentiva su uso y deriva en la subutilización del hardware multinúcleo], [Fragmentación del ecosistema de software, lo que dificulta la adopción y reutilización de algoritmos de alto rendimiento en arquitecturas modernas],
    [PROBLEMA CENTRAL], table.cell(colspan: 3)[Existe una brecha semántica y de abstracción entre el paradigma de paralelismo basado en directivas y los ecosistemas de lenguajes de propósito general modernos.],
    [PROBLEMAS CAUSA], [Divergencia semántica entre el modelo tradicional de las directivas de paralelismo y las abstracciones de concurrencia nativas de los lenguajes modernos], [Alta complejidad inherente en la transformación de código para mapear directivas hacia las primitivas de concurrencia de lenguajes modernos], [Insuficiencia de implementaciones de referencia y de evidencia empírica que validen la viabilidad de este paradigma fuera del ecosistema tradicional],
  ),
  caption: [Esquema del árbol de problemas],
  kind: table,
) <tab:tabla-1-esquema-del-arbol-de-p>


=== Descripción

El problema central radica en la brecha semántica y de abstracción entre el paralelismo basado en directivas y los lenguajes de propósito general modernos. Una causa fundamental de esta brecha es la divergencia semántica: las primitivas de concurrencia modernas no se alinean directamente con los patrones tradicionales de memoria compartida de estándares como OpenMP. En consecuencia, la implementación manual del paralelismo obliga al programador a entrelazar la lógica del algoritmo con los detalles de sincronización, alejándose del principio de separación de incumbencias (Kambites et al., 2001; Vikas et al., 2014).

Asimismo, abordar esta divergencia implica una alta complejidad inherente en la transformación de código. Mapear directivas hacia runtimes modernos gestionados (con recolectores de basura y planificadores propios) requiere el desarrollo de compiladores especializados. Esto se relaciona con la tercera causa: la insuficiencia de implementaciones de referencia. Aunque existen esfuerzos recientes para adaptar el paradigma de directivas a lenguajes como Python (Piñeiro & Pichel, 2026) o Zig (Kacs et al., 2024), aún existe un vacío empírico sobre su viabilidad y rendimiento generalizado fuera de C, C++ y Fortran.

Esta situación genera tres efectos adversos críticos en el desarrollo de software. Primero, produce una fragmentación arquitectónica del software al forzar la separación estructural entre los módulos enfocados en la productividad y lógica general (desarrollados en lenguajes modernos) y aquellos relegados estrictamente al alto rendimiento (escritos en lenguajes tradicionales). Esta dicotomía, conocida en la literatura como el problema de los dos lenguajes, introduce una compleja y costosa necesidad de integración mediante interfaces de funciones foráneas, lo que aumenta la complejidad operativa y dificulta el mantenimiento del sistema (Bezanson et al., 2017). Segundo, fuerza la dependencia de lenguajes tradicionales para garantizar el rendimiento, sacrificando la expresividad, la gestión automática de memoria y la seguridad que previenen errores técnicos complejos (Nanz y Furia, 2015). Tercero, si se opta por paralelizar manualmente en lenguajes modernos, el aumento del esfuerzo de programación y el código repetitivo elevan el riesgo de errores de concurrencia como condiciones de carrera o _deadlocks_ (Powers & Alaghband, 2007), lo que frecuentemente desincentiva su uso y deriva en la subutilización del hardware multinúcleo.


=== Problema Seleccionado

El problema central que aborda esta investigación es la brecha semántica y de abstracción que impide integrar eficientemente el paradigma de paralelismo basado en directivas en los ecosistemas de lenguajes de propósito general modernos.

Para acotar y abordar empíricamente este problema, se ha seleccionado a Go como lenguaje de implementación y caso de estudio. La elección de Go se fundamenta en su notable simplicidad sintáctica y sus potentes primitivas nativas de concurrencia (goroutines y channels), diseñadas inherentemente para facilitar el desarrollo de sistemas escalables y maximizar el aprovechamiento de arquitecturas multinúcleo (Pike, 2012; Donovan & Kernighan, 2015). Sin embargo, al aplicar Go en tareas que involucran concurrencia, sus primitivas obligan al programador a gestionar manualmente la sincronización y el reparto de trabajo. Esta gestión manual no solo introduce un fuerte acoplamiento y eleva la carga cognitiva, sino que incrementa significativamente la probabilidad de introducir errores complejos de concurrencia, como condiciones de carrera y _deadlocks_, un problema empíricamente documentado en el ecosistema del lenguaje (Tu et al., 2019). Entonces, es posible incluir a Go en el grupo de lenguajes modernos a los que se hace referencia.

Para superar esta limitación, la presente investigación propone diseñar, desarrollar y evaluar la herramienta GompherMP. El resultado principal será un compilador _source-to-source_ y una librería de runtime funcional que automatice la paralelización en Go mediante directivas. La efectividad de esta propuesta se validará a través de _benchmarks_, comparando su rendimiento computacional y la expresividad del código resultante frente a las implementaciones paralelas manuales tradicionales, demostrando así la viabilidad del paradigma en un lenguaje moderno.


== Objetivos

En esta sección, se presentan tanto el objetivo general como los objetivos específicos. El objetivo general se enfoca a resolver el problema seleccionado, mientras que los objetivos específicos guardan relación con los problemas causa mencionados en el árbol de problemas diagramado anteriormente.


=== Objetivo General

Diseñar, desarrollar y evaluar la herramienta GompherMP que implemente un modelo de paralelismo de alto nivel basado en directivas para el lenguaje Go, inspirado en un subconjunto del estándar OpenMP, a fin de cerrar la brecha semántica y de abstracción entre el paradigma de paralelismo basado en directivas y los ecosistemas de lenguajes de propósito general modernos.


=== Objetivos Específicos

*OE1.* Diseñar y especificar la arquitectura de la herramienta GompherMP, incluyendo la sintaxis de sus directivas y cláusulas, los algoritmos de transformación de código y la interfaz de línea de comandos (CLI).

*OE2.* Implementar la infraestructura central de GompherMP, integrando un compilador _source-to-source_ para la transformación de directivas y una librería de runtime para la orquestación de la concurrencia y sincronización en Go.

*OE3.* Evaluar la herramienta GompherMP mediante la ejecución de _benchmarks_, comparando sus resultados en términos de rendimiento y complejidad sintáctica del código frente a las implementaciones secuenciales y paralelas manuales en Go.


=== Resultados Esperados

*OE1. Diseñar y especificar la arquitectura de la herramienta GompherMP, incluyendo la sintaxis de sus directivas y cláusulas, los algoritmos de transformación de código y la interfaz de línea de comandos (CLI).*

#pad(left: 2em)[
  R1. Documento de especificación de directivas y cláusulas.

  R2. Diseño de la arquitectura de la herramienta.

  R3. Especificación funcional de la Interfaz de Línea de Comandos (CLI).
]

*OE2. Implementar la infraestructura central de GompherMP, integrando un compilador source-to-source para la transformación de directivas y una librería de runtime para la orquestación de la concurrencia y sincronización en Go.*
#pad(left: 2em)[
  R4. Módulo de gestión de goroutines y reparto de trabajo implementado y probado.

  R5. Módulo de mecanismos de sincronización implementado y probado.

  R6. Módulo de soporte para paralelismo de tareas (_tasking_) implementado y probado.

  R7. Analizador sintáctico (Parser) de directivas GompherMP.

  R8. Motor de transformación del AST implementado.

  R9. Herramienta GompherMP (CLI) funcional.

]
*OE3. Evaluar la herramienta GompherMP mediante la ejecución de benchmarks, comparando sus resultados en términos de rendimiento y complejidad sintáctica del código frente a las implementaciones secuenciales y paralelas manuales en Go.*

#pad(left: 2em)[
  R10. Suite de _benchmarks_ implementada.

  R11. Informe de evaluación de rendimiento y escalabilidad.

  R12. Reporte de análisis comparativo sobre expresividad y productividad.
]

=== Mapeo de problemática con objetivos

A través de la siguiente tabla, se pretende mostrar cómo cada objetivo específico se relaciona con cada uno de los problemas causa del árbol de problemas.

#figure(
  table(
    columns: 2,
    stroke: 0.5pt,
    fill: (col, row) => if row == 0 { luma(230) },
    align: (col, row) => if row == 0 { center + horizon } else { left + top },
    [*Problema causa*], [*Objetivo Específico*],
    [Divergencia semántica entre el modelo tradicional de las directivas de paralelismo y las abstracciones de concurrencia nativas de los lenguajes modernos], [OE1. Diseñar y especificar la arquitectura de la herramienta GompherMP, incluyendo la sintaxis de sus directivas y cláusulas, los algoritmos de transformación de código y la interfaz de línea de comandos (CLI).],
    [Alta complejidad inherente en la transformación de código para mapear directivas hacia las primitivas de concurrencia de lenguajes modernos], [OE2. Implementar la infraestructura central de GompherMP, integrando un compilador source-to-source para la transformación de directivas y una librería de runtime para la orquestación de la concurrencia y sincronización en Go.],
    [Insuficiencia de implementaciones de referencia y de evidencia empírica que validen la viabilidad de este paradigma fuera del ecosistema tradicional], [OE3. Evaluar la herramienta GompherMP mediante la ejecución de benchmarks, comparando sus resultados en términos de rendimiento y complejidad sintáctica del código frente a las implementaciones secuenciales y paralelas manuales en Go.],
  ),
  caption: [Relación entre problemas causa y objetivos específicos],
  kind: table,
) <tab:tabla-2-relacion-entre-problem>


=== Mapeo de objetivos, resultados y medios de verificación

En esta sección, se pretende mostrar cómo cada objetivo específico se traduce en resultados esperados juntos con sus medios de verificación e indicadores objetivamente verificables. Esto permitirá que se realice una evaluación adecuada y precisa en el transcurso del proyecto.

#figure(
  caption: [Resultados esperados, medios de verificación e indicadores objetivamente verificables (IOV) para el objetivo específico OE1],
  kind: table,
  table(
    columns: 3,
    stroke: 0.5pt,
    fill: (col, row) => if row < 2 { luma(230) },
    align: (col, row) => if row == 1 { center + top } else { left + top },
    table.header(
      repeat: false,
      table.cell(colspan: 3)[*Objetivo 1: Diseñar y especificar la arquitectura de la herramienta GompherMP, incluyendo la sintaxis de sus directivas y cláusulas, los algoritmos de transformación de código y la interfaz de línea de comandos (CLI).*],
      [*Resultado*], [*Medio de verificación*], [*Indicador objetivamente verificable (IOV)*],
    ),

  [R1. Documento de especificación de directivas y cláusulas.],
  [- Informe técnico de especificación de GompherMP.],
  [- El informe define la sintaxis, semántica y presenta al menos un ejemplo de uso para cada directiva y cláusula del alcance del proyecto (parallel, for, task, depend, critical, etc.).
  - Se obtienela aprobación escrita de un experto en programación concurrente.],
  [R2. Diseño de la arquitectura de la herramienta.],
  [- Documento de diseño de arquitectura.],
  [- El documento contiene diagramas que ilustran la interacción entre el compilador _source-to-source_ y la librería de runtime.
  - Existe una descripción del flujo general del proceso de transformación del AST.
  - Se obtiene laaprobación escrita de la arquitectura por parte de un experto en programación concurrente.],
  [R3. Especificación funcional de la Interfaz de Línea de Comandos (CLI).],
  [- Manual de usuario de la CLI.],
  [- El manual especifica todos los comandos, argumentos y opciones disponibles.
  - Incluye ejemplos claros deuso para la transpilación de código.
  - Se obtiene la aprobación escrita de un experto en programación concurrente.],
  ),
) <tab:tabla-3-resultados-esperados-m>

#figure(
  caption: [Resultados esperados, medios de verificación e indicadores objetivamente verificables (IOV) para el objetivo específico OE2],
  kind: table,
  table(
    columns: 3,
    stroke: 0.5pt,
    fill: (col, row) => if row < 2 { luma(230) },
    align: (col, row) => if row == 1 { center + top } else { left + top },
    table.header(
      repeat: false,
      table.cell(colspan: 3)[*Objetivo 2: Implementar la infraestructura central de GompherMP, integrando un compilador source-to-source para la transformación de directivas y una librería de runtime para la orquestación eficiente de la concurrencia y sincronización en Go.*],
      [*Resultado*], [*Medio de verificación*], [*Indicador objetivamente verificable (IOV)*],
    ),

  [R4. Módulo de gestión de goroutines y reparto de trabajo implementado y probado.], 
  [- Código fuente en repositorio en línea. 
  - Informe de pruebas unitarias.], 
  [- Repositorio accesible con código fuente documentado.
  - El código implementa la gestión del _pool_ de goroutines y el reparto de iteraciones para las directivas parallel, for, parallel for sections y schedule.
  - El informe evidencia una cobertura de pruebas de al menos el 80% del código para el módulo.],
  [R5. Módulo de mecanismos de sincronización implementado y probado.],
  [- Código fuente en repositorio en línea.
  - Informe de pruebas unitarias.], 
  [- El código fuente implementa la funcionalidad para las directivas critical, single, master y barrier.
  - El informe evidencia una cobertura de pruebas de al menos el 80% del código para el módulo.],
  [R6. Módulo de soporte para paralelismo de tareas (_tasking_) implementado y probado.],
  [- Código fuente en repositorio en línea.
  - Informe de pruebas de integración.],
  [- El código implementa la creación de tareas y la gestión de su grafo dedependencias (in, out, inout).
  - Las pruebas validan la correcta ejecución de tareas según las dependencias definidas. Incluye las siguientes directivas: task, `taskwait`, `taskgroup` y `taskloop`.
  - El informe evidencia una cobertura de pruebas deal menos el 80% del código para el módulo.],
  [R7. Analizador sintáctico (Parser) de directivas GompherMP.],
  [- Código fuente del compilador en repositorio en línea.
  - Informe de pruebas de integración.],
  [- El módulo de código es capaz de identificar y extraer correctamente todas lasdirectivas y cláusulas definidas en la especificación (O1.R1) a partir de los comentarios \/\/gompher.
  - El informe evidencia una cobertura de pruebas de al menos el 80% del código para el módulo.],
  [R8. Motor de transformación del AST implementado.],
  [- Código fuente del compilador en repositorio en línea.
  - Informe de pruebas de integración.],
  [- El código modifica el AST de un programa de entrada, inyectando las llamadas a la librería deruntime correspondientes a las directivas encontradas.
  - Implementa la transformación de la directiva atomic y lascláusulas de gestión de datos como private, `firstprivate`, `lastprivate`, shared y reduction.
  - El informe evidencia unacobertura de pruebas de al menos el 80% del código para el módulo.],
  [R9. Herramienta GompherMP (CLI) funcional.],
  [- Ejecutable de la herramienta CLI.
  - Guía de instalación y uso.],
  [- La herramienta transpila correctamente archivos .go con directivas GompherMP, generando como salida un archivo .go con código concurrente nativo, el cual debe ser compilable con el compilador estándar de Go.
  - Un experto en programación concurrente valida el funcionamiento respecto a su especificación funcional],
  ),
) <tab:tabla-4-resultados-esperados-m>

#figure(
  caption: [Resultados esperados, medios de verificación e indicadores objetivamente verificables (IOV) para el objetivo específico OE3],
  kind: table,
  table(
    columns: 3,
    stroke: 0.5pt,
    fill: (col, row) => if row < 2 { luma(230) },
    align: (col, row) => if row == 1 { center + top } else { left + top },
    table.header(
      repeat: false,
      table.cell(colspan: 3)[*Objetivo 3: Evaluar la herramienta GompherMP mediante la ejecución de benchmarks, comparando sus resultados en términos de rendimiento y expresividad del código frente a las implementaciones secuenciales y paralelas manuales en Go.*],
      [*Resultado*], [*Medio de verificación*], [*Indicador objetivamente verificable (IOV)*],
    ),

  [R10. Suite de _benchmarks_ implementada.],
  [- Código fuente de los algoritmos de _benchmark_ en repositorio en línea.],
  [- El repositorio contiene al menos 3 algoritmos de cómputo intensivo.
  - Cada algoritmo está implementado en 3 versiones: secuencial, paralela manual y paralela con GompherMP.
  - Se obtiene la aprobación escrita de un experto en programación concurrente.],
  [R11. Informe de evaluación de rendimiento y escalabilidad.],
  [- Informe de evaluación de rendimiento.],
  [- El informe presenta los resultados de tiempo de ejecución, _speedup_ y eficiencia para cada versión de los _benchmarks_.
  - El informe incluye gráficos comparativos que visualizan los resultados.
  - Se presentan conclusiones respecto a la efectividad de la herramienta en el informe.
  - El informe debe ser aprobado por un experto en programación concurrente.],
  [R12. Reporte de análisis comparativo sobre expresividad y productividad.],
  [- Reporte de análisis comparativo.],
  [- El reporte incluye un análisis cuantitativo (ej. líneas de código) y cualitativo de la complejidad y legibilidad del código entre la versión paralela manual y la versión con GompherMP. 
  - Se presentan conclusiones sobre el impacto de GompherMP en la productividad del desarrollador.],
  ),
) <tab:tabla-5-resultados-esperados-m>


== Metodología

A continuación, se detallan las herramientas, métodos y procedimientos necesarios para construir los resultados esperados de esta tesis. Para cada objetivo específico, se presenta una tabla que conecta los resultados con las técnicas y recursos que permitirán alcanzarlos, junto con una descripción de cada uno de ellos.

#figure(
  table(
    columns: (0.75fr, 1fr),
    stroke: 0.5pt,
    fill: (col, row) => if row < 2 { luma(230) },
    align: (col, row) => if row == 1 { center + top } else { left + top },
    table.header(
      repeat: false,
      table.cell(colspan: 2)[*Objetivo 1: Diseñar y especificar la arquitectura de la herramienta GompherMP, incluyendo la sintaxis de sus directivas y cláusulas, los algoritmos de transformación de código y la interfaz de línea de comandos (CLI).*],
      [*Resultado*], [*Herramientas, métodos y procedimientos*],
    ),
    [R1. Documento de especificación de directivas y cláusulas.], [*Herramienta:* Typst],
    [R2. Diseño de la arquitectura de la herramienta.], [*Herramientas:* Excalidraw, Typst \ Método: Arquitectura centrada en componentes],
    [R3. Especificación funcional de la Interfaz de Línea de Comandos (CLI).], [*Herramienta:* Typst],
  ),
  caption: [Herramientas, métodos y procedimientos relacionados al objetivo específico 1 y sus resultados esperados],
  kind: table,
) <tab:tabla-6-herramientas-metodos-y>

#figure(
  table(
    columns: (0.75fr, 1fr),
    stroke: 0.5pt,
    fill: (col, row) => if row < 2 { luma(230) },
    align: (col, row) => if row == 1 { center + top } else { left + top },
    table.header(
      repeat: false,
      table.cell(colspan: 2)[*Objetivo 2: Implementar la infraestructura central de GompherMP, integrando un compilador source-to-source para la transformación de directivas y una librería de runtime para la orquestación eficiente de la concurrencia y sincronización en Go.*],
      [*Resultado*], [*Herramientas, métodos y procedimientos*],
    ),
    [R4. Módulo de gestión de goroutines y reparto de trabajo implementado y probado.], [*Herramientas:* Go, Neovim, VScode, Github \ *Métodos:* Metodología Kanban, TDD \ *Procedimientos:* Pruebas unitarias],
    [R5. Módulo de mecanismos de sincronización implementado y probado.], [*Herramientas:* Go, Neovim, VScode, Github \ *Métodos:* Metodología Kanban, TDD \ *Procedimientos:* Pruebas unitarias],
    [R6. Módulo de soporte para paralelismo de tareas (tasking) implementado y probado.], [*Herramientas:* Go, Neovim, VScode, Github \ *Métodos:* Metodología Kanban, TDD \ *Procedimientos:* Pruebas unitarias],
    [R7. Analizador sintáctico (Parser) de directivas GompherMP.], [*Herramientas:* Go, Neovim, VScode, Github, Go/AST, Go/test \ *Métodos:* Metodología kanban, TDD, análisis sintáctico \ *Procedimientos:* Pruebas unitarias],
    [R8. Motor de transformación del AST implementado.], [*Herramientas:* Go, Neovim, VScode, Github, Go/AST, Go/test \ *Métodos:* Metodología kanban, TDD, transformación de abstract syntax trees \ *Procedimientos:* Pruebas unitarias],
    [R9. Herramienta GompherMP (CLI) funcional.], [*Herramientas:* Go, Neovim, VScode, Github, Go/flag \ *Métodos:* Metodología kanban, TDD \ *Procedimientos:* Pruebas unitarias, pruebas de integración],
  ),
  caption: [Herramientas, métodos y procedimientos relacionados al objetivo específico 2 y sus resultados esperados],
  kind: table,
) <tab:tabla-7-herramientas-metodos-y>

#figure(
  table(
    columns: (0.75fr, 1fr),
    stroke: 0.5pt,
    fill: (col, row) => if row < 2 { luma(230) },
    align: (col, row) => if row == 1 { center + top } else { left + top },
    table.header(
      repeat: false,
      table.cell(colspan: 2)[*Objetivo 3: Evaluar la herramienta GompherMP mediante la ejecución de benchmarks, comparando sus resultados en términos de rendimiento y expresividad del código frente a las implementaciones secuenciales y paralelas manuales en Go.*],
      [*Resultado*], [*Herramientas, métodos y procedimientos*],
    ),
    [R10. Suite de benchmarks implementada.], [*Herramientas:* Go, C, OpenMP \ *Métodos:* Benchmarking de algoritmos de computación intensiva \ *Procedimiento:* Performance profiling],
    [R11. Informe de evaluación de rendimiento y escalabilidad.], [*Herramientas:* Python, Matplotlib, Typst \ *Métodos:* Análisis comparativo con pruebas estadísticas, análisis cualitativo \ *Procedimiento:* ANOVA, t-test],
    [R12. Reporte de análisis comparativo sobre expresividad y productividad.], [*Herramientas:* Python, Scipy, Typst \ *Métodos:* Análisis comparativo con pruebas estadísticas, análisis cualitativo \ *Procedimiento:* ANOVA, t-test],
  ),
  caption: [Herramientas, métodos y procedimientos relacionados al objetivo específico 3 y sus resultados esperados],
  kind: table,
) <tab:tabla-8-herramientas-metodos-y>


=== Herramientas


==== Typst

Según Typst (s.f.), Typst es un sistema de preparación de documentos de tipografía profesional, diseñado como una alternativa moderna a LaTeX que combina la calidad tipográfica de los sistemas tradicionales con una sintaxis simplificada y una compilación significativamente más rápida. Para este proyecto, resulta adecuado en la elaboración de documentos como el informe de evaluación de rendimiento, el diseño de la arquitectura y los reportes de cobertura de pruebas, dado que su capacidad para gestionar automáticamente referencias cruzadas, figuras y tablas garantiza la consistencia.


==== Excalidraw

Según Excalidraw (s.f.), Excalidraw es una pizarra virtual para esbozar diagramas con aspecto de dibujo a mano. Para este proyecto, será una herramienta principal en la fase de diseño, específicamente para crear los diagramas de arquitectura requeridos. Su enfoque en la simplicidad facilitará la ilustración de la interacción entre el compilador _source-to-source_ y la librería de runtime, así como el flujo de transformación del AST.


==== Go

Según el equipo de Go (s.f.), Go es un lenguaje de programación de código abierto, apoyado por Google, que facilita la construcción de software simple, seguro y escalable, destacando por su concurrencia nativa y una robusta librería estándar. En el contexto de este proyecto, Go es el lenguaje objetivo sobre el cual se construye la herramienta GompherMP; el compilador _source-to-source_ propuesto analizará código Go con directivas y lo transformará para generar código concurrente nativo, utilizando las primitivas del lenguaje como goroutines y channels para implementar el paralelismo.


==== Neovim

Según Neovim (s.f.), Neovim es un editor de texto de la familia Vim, enfocado en la extensibilidad y la usabilidad, que puede ser utilizado desde un terminal o como una aplicación gráfica independiente. Para el desarrollo de este proyecto, Neovim será un entorno de codificación donde se implementará tanto el compilador _source-to-source_ como la librería de runtime de GompherMP. Su naturaleza ligera y configurable permitirá un flujo de trabajo eficiente para escribir, probar y depurar el código Go que compone la herramienta.


==== VSCode

Según Microsoft (s.f.), Visual Studio Code es un editor de código fuente ligero pero potente que se ejecuta en el escritorio y está disponible para Windows, macOS y Linux. Para el desarrollo de este proyecto, Visual Studio Code será otro entorno de codificación utilizado para colaborar en la implementación de la herramienta. Sus robustas capacidades de depuración, la integración con Git y el amplio ecosistema de extensiones facilitarán su contribución en el desarrollo tanto del compilador como de la librería de runtime.


==== Github

Según GitHub (s.f.), GitHub es la plataforma de desarrollo completa para construir, escalar y entregar software de forma segura. Para este proyecto, servirá como el repositorio de código fuente central para la herramienta GompherMP, facilitando el control de versiones y la colaboración durante la implementación del compilador y la librería de runtime. Además, el repositorio en línea actuará como el medio de verificación principal para los entregables de código, tal como se especifica en los objetivos específicos 2 y 3 del proyecto.


==== Go/AST

Según el equipo de Go (s.f.), el paquete ast declara los tipos utilizados para representar árboles de sintaxis abstracta para archivos de código Go. En este proyecto, este paquete es fundamental para el compilador _source-to-source_, ya que su función principal es analizar el código fuente, interpretar las directivas y transformar el Árbol de Sintaxis Abstracta (AST) para generar código Go nativo concurrente. La capacidad de manipular directamente la estructura del código a través del AST es, por lo tanto, esencial para cumplir con el objetivo específico 3 del proyecto.


==== Go/test

Según el equipo de Go (s.f.), el paquete testing provee el soporte para pruebas automatizadas de paquetes de Go, siendo la base para la creación de tests unitarios y _benchmarks_. Dentro de este proyecto, será una herramienta crucial para cumplir con los objetivos de evaluación, ya que se utilizará para implementar las pruebas unitarias y de integración que validarán la correctitud de los módulos de la librería de runtime y del compilador (objetivo específico 2). Asimismo, será fundamental para desarrollar la suite de _benchmarks_ necesaria para medir el rendimiento y la escalabilidad de GompherMP, tal como lo exige el objetivo específico 3.


==== Go/flag

Según el equipo de Go (s.f.), el paquete `flag` de la librería estándar implementa el análisis de las opciones de la línea de comandos. En el contexto de este proyecto, el paquete `flag` será la base para construir la Interfaz de Línea de Comandos (CLI) de la herramienta GompherMP, tal como se especifica en los objetivos 1 y 2. Al pertenecer a la librería estándar del lenguaje, su uso simplifica la definición de los comandos, argumentos y opciones necesarios para que el usuario pueda transpilar su código Go de manera sencilla, sin introducir dependencias externas a la herramienta.


==== C

Según C-Language.org (s.f.), C es un lenguaje de programación de propósito general conocido por su rendimiento y su capacidad para operar a bajo nivel. En el marco de esta investigación, se utilizará para desarrollar implementaciones de referencia para los algoritmos de la suite de _benchmarks_, como se detalla en el objetivo específico 3. Dado que los estándares de paralelismo como OpenMP tienen su dominio tradicional en lenguajes de sistemas, estas versiones en C servirán como una línea base de rendimiento contra la cual se podrán contrastar los resultados de las implementaciones en Go, enriqueciendo así el informe de evaluación de rendimiento y escalabilidad.


==== OpenMP

Según la OpenMP Architecture Review Board (s.f.), OpenMP es una API que soporta la programación paralela de memoria compartida en C, C++ y Fortran mediante un conjunto de directivas de compilador. Este estándar es la inspiración fundamental para el presente proyecto, ya que la herramienta GompherMP busca adaptar un subconjunto de su modelo de paralelismo basado en directivas al ecosistema de Go. Por lo tanto, la especificación de OpenMP proporciona el marco conceptual para el diseño de las directivas y cláusulas de GompherMP, sirviendo como referente para evaluar la expresividad y funcionalidad de la solución propuesta.


==== Python

Según la Python Software Foundation (s.f.), Python es un lenguaje de programación que permite trabajar rápidamente e integrar sistemas de manera más efectiva. En el contexto de esta investigación, se utilizará para desarrollar versiones de los algoritmos de la suite de _benchmarks_, permitiendo una comparación de rendimiento y expresividad frente a la solución propuesta en Go, similar a como se plantea para C. Adicionalmente, se empleará para la automatización de la ejecución de pruebas y la generación de gráficos para el informe de evaluación, facilitando el análisis comparativo requerido en el objetivo específico 3.


==== Matplotlib

Según The Matplotlib Development Team (s.f.), Matplotlib es una librería completa para crear visualizaciones estáticas, animadas e interactivas en Python. En el marco de este proyecto, esta librería será fundamental para cumplir con los requisitos del objetivo específico 3, ya que se empleará para generar los gráficos comparativos que visualizarán los resultados de rendimiento y escalabilidad. Estos gráficos, que mostrarán métricas como el tiempo de ejecución y el _speedup_ de las versiones secuencial, manual y con GompherMP, son un componente esencial del informe de evaluación de rendimiento.


==== Scipy

Según la comunidad de SciPy (s.f.), SciPy es una librería de Python que proporciona algoritmos y rutinas numéricas eficientes para optimización, álgebra lineal y estadística. En el marco de esta investigación, se utilizará para realizar las pruebas estadísticas sobre los datos de rendimiento obtenidos de la suite de _benchmarks_. Esto permitirá validar con rigor si las diferencias en métricas como el _speedup_ y el tiempo de ejecución son significativas, fortaleciendo las conclusiones del informe de evaluación de rendimiento.


=== Métodos


==== Arquitectura centrada en componentes

Según Bass et al. (2012), una arquitectura centrada en componentes se enfoca en la descomposición del sistema en unidades funcionales o lógicas con interfaces bien definidas. Este enfoque será aplicado en el diseño de GompherMP, separando el compilador _source-to-source_ y la librería de runtime como componentes distintos. Esta separación modular facilitará el desarrollo, las pruebas y el mantenimiento independiente de cada parte, asegurando que la interacción entre ellos para la transformación y ejecución del código sea clara y cohesiva.


==== Metodología Kanban

De acuerdo con Anderson (2010), Kanban es una metodología para gestionar el flujo de trabajo que se centra en la visualización del trabajo, la limitación del trabajo en progreso y la maximización de la eficiencia. Para la gestión de este proyecto, se utilizará un tablero Kanban para organizar y priorizar las tareas asociadas a los cuatro objetivos específicos, desde el diseño de la arquitectura hasta la evaluación final, permitiendo un seguimiento transparente del avance y una adaptación continua a los desafíos que surjan durante el desarrollo.


==== Desarrollo guiado por pruebas (TDD)

Según Beck (2003), el Desarrollo Guiado por Pruebas es una práctica de desarrollo de software donde los tests se escriben antes que el código que los debe pasar. Esta metodología se aplicará en la implementación de la librería de runtime de GompherMP, asegurando que cada función de sincronización y gestión de goroutines sea robusta y correcta desde su concepción. Este enfoque es fundamental para cumplir con los indicadores de cobertura de pruebas superiores al 80% estipulados en el objetivo específico 2.


==== Análisis Léxico y Sintáctico

Este método constituye la fase inicial o frontend de la solución propuesta, siguiendo el modelo de fases secuenciales descrito por Aho et al. (2007). El procedimiento comenzará con un análisis léxico (_scanning_), encargado de leer el flujo de caracteres del código fuente Go y agruparlos en tokens significativos (lexemas). Posteriormente, se ejecutará el análisis sintáctico, que utilizará estos tokens para construir una representación jerárquica de la estructura gramatical. Este método es fundamental para el objetivo específico 2, ya que permitirá identificar y extraer correctamente las directivas \/\/gompher y las estructuras de control asociadas, generando el insumo necesario para la etapa de transformación.


==== Compilación Source-to-Source y Transformación de AST

De acuerdo con Aho et al. (2007) y Cooper & Torczon (2012), la transformación de un Árbol de Sintaxis Abstracta (AST) implica modificar la representación estructural del código fuente preservando su semántica. Este método constituye el núcleo de la arquitectura _source-to-source_ de GompherMP para cumplir con el objetivo específico 2.

El procedimiento consiste en recorrer el AST generado en la fase previa (_parsing_), identificar los nodos marcados por las directivas y reescribir su estructura inyectando las llamadas a la librería de runtime (goroutines y canales). A diferencia de una compilación tradicional a código máquina, este enfoque finaliza con una etapa de síntesis de código, donde el AST transformado se serializa nuevamente en archivos de texto .go. Esto garantiza que el código resultante sea nativo, legible y portable, pudiendo ser compilado por el _toolchain_ estándar de Go sin modificaciones.


==== Benchmarking de algoritmos de computación intensiva

Según Jain (1991), el _benchmarking_ es el proceso de ejecutar un programa para evaluar su rendimiento relativo de forma cuantitativa. Este método será la base para la evaluación de GompherMP (objetivo específico 3), donde se implementará una suite de _benchmarks_ con al menos tres algoritmos de cómputo intensivo. La ejecución de estos algoritmos en sus versiones secuencial, paralela manual y con GompherMP permitirá medir y comparar métricas clave como el tiempo de ejecución y el _speedup_.


==== Análisis comparativo con pruebas estadísticas

De acuerdo con Thiel (2014), un análisis comparativo utiliza métodos estadísticos para determinar si las diferencias observadas entre los resultados de dos o más grupos son significativas o si podrían haber ocurrido por azar. Este método se empleará para analizar los datos de rendimiento obtenidos del _benchmarking_. Mediante pruebas estadísticas, se validará si la mejora de rendimiento de GompherMP es estadísticamente significativa en comparación con las versiones secuencial y manual, fortaleciendo las conclusiones del informe de evaluación.


==== Análisis cualitativo

Según Miles et al. (2014), el análisis cualitativo se enfoca en datos no numéricos para comprender conceptos y experiencias. Este método se utilizará para evaluar la expresividad y la productividad, como se exige en el resultado R12 del objetivo 3. Se realizará un análisis comparativo de la complejidad y legibilidad del código entre la versión paralela manual y la versión con GompherMP, apoyado en métricas como las líneas de código (LoC), para concluir sobre el impacto de la herramienta en la productividad del desarrollador.


=== Procedimientos


==== Pruebas unitarias

Según Sommerville (2011), las pruebas unitarias se centran en verificar la funcionalidad de componentes individuales de un programa de forma aislada. Este procedimiento será aplicado sistemáticamente sobre los módulos de la librería de runtime, como el gestor del _pool_ de goroutines y los mecanismos de sincronización. El objetivo es asegurar que cada función se comporte como se espera antes de integrarla en el sistema completo, cumpliendo con los indicadores de cobertura de pruebas.


==== Pruebas de integración

De acuerdo con Sommerville (2011), las pruebas de integración tienen como objetivo descubrir defectos en las interfaces y las interacciones entre componentes integrados. Este procedimiento se utilizará para validar que el compilador _source-to-source_ y la librería de runtime de GompherMP funcionan correctamente en conjunto. Se verificarán casos de uso completos, desde la transpilación de un archivo .go con directivas hasta la correcta ejecución del código concurrente generado.


==== Performance profiling

Según el equipo de Go (s.f.), el performance profiling es el análisis del software para identificar cuellos de botella y optimizar su rendimiento. Este procedimiento se aplicará utilizando herramientas del ecosistema de Go, como pprof, sobre el código generado por GompherMP. El objetivo es analizar el comportamiento de la librería de runtime y los algoritmos de la suite de _benchmarks_ para detectar posibles sobrecargas (_overhead_) y asegurar que la implementación sea lo más eficiente posible.


==== ANOVA

Según Snedecor y Cochran (1989), el análisis de varianza (ANOVA) es una prueba estadística utilizada para determinar si existen diferencias estadísticamente significativas entre las medias de tres o más grupos independientes. Este procedimiento se empleará en el análisis de los resultados de los _benchmarks_ para comparar simultáneamente los tiempos de ejecución de las versiones secuencial, paralela manual y con GompherMP de cada algoritmo, y así determinar si el método de paralelización tiene un efecto significativo en el rendimiento.


==== Prueba t (t-test)

De acuerdo con Snedecor y Cochran (1989), la prueba t es un método estadístico que se utiliza para comparar las medias de dos grupos. En el marco de esta investigación, este procedimiento se usará para realizar comparaciones específicas después de un análisis ANOVA, por ejemplo, para determinar si la diferencia de rendimiento entre la versión paralela con GompherMP y la versión paralela manual es estadísticamente significativa, ofreciendo una validación más granular de la eficiencia de la herramienta.
