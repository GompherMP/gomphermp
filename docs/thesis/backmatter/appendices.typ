
// Los encabezados de anexos no se numeran como capítulos (llevan "Anexo A." manual).
#set heading(numbering: none)

= Anexos

== Anexo A. Formulario de Extracción Aplicado

En este anexo se proporciona acceso al conjunto completo de datos recopilados durante la revisión sistemática de la literatura. El enlace adjunto dirige a una hoja de cálculo que contiene la aplicación detallada del formulario de extracción, definido en el Capítulo 2, para cada uno de los artículos incluidos en este estudio.

#link("https://docs.google.com/spreadsheets/d/1esh1d59-I08rL1KxT_cNNeMCng-wOmas4HU-_uyKId0/edit?usp=sharing")

== Anexo B. Plan de Proyecto

En este anexo se proporciona el plan de proyecto a seguir para cumplir con todos los resultados esperados del proyecto de tesis.

*Justificación*

La realización del presente trabajo se fundamenta en los siguientes criterios:

*Conveniencia:* Se busca cerrar la brecha existente entre la productividad del desarrollador y el rendimiento del código en algoritmos de cómputo intensivo, mitigando la fricción semántico-sintáctica actual en lenguajes como Go.

*Valor teórico:* La investigación se sitúa en la intersección entre la teoría de compiladores y la Computación de Alto Rendimiento (HPC), contribuyendo conceptualmente a la integración de paradigmas declarativos de concurrencia en lenguajes de alto nivel.

*Utilidad metodológica:* Se propone una metodología híbrida de evaluación que combina métricas cuantitativas de performance con un análisis cualitativo de expresividad y productividad.

*Implicaciones prácticas:* La herramienta reducirá la carga cognitiva y la tasa de errores asociados a la sincronización manual, disminuyendo las barreras para la optimización y escalabilidad del código.

*Viabilidad*

*Viabilidad Técnica:* Se sustenta en una arquitectura componetizada y desacoplada que favorece el desarrollo guiado por pruebas (TDD). El alcance técnico se limita a un subconjunto de directivas, apoyándose en herramientas nativas del ecosistema Go (ast, testing, pprof) .

*Viabilidad Económica:* El proyecto se basa en software de código abierto (licencias MIT/GNU) y hardware personal, eliminando la necesidad de adquirir GPUs. Se contempla el uso de Cloud Computing para pruebas estandarizadas.

*Viabilidad Temporal:* Las fases están delimitadas para su ejecución durante el verano de 2026 y el ciclo académico 2026-1, alrededor de 6 meses.

*Alcance*

El proyecto comprende el desarrollo de una herramienta de paralelismo estructurado y basado en tareas para Go.

*Incluye:* Implementación de un subset de directivas y cláusulas OpenMP; desarrollo de una CLI; construcción de un compilador source-to-source (transformación de AST); implementación de un runtime (gestión de goroutine pool, scheduling, sincronización); y evaluación mediante benchmarking de algoritmos de cómputo intensivo .

*Excluye:* Especificación completa del estándar OpenMP, soporte para CUDA/GPUs y paralelización automática sin directivas explícitas .

*Limitaciones*

*Mismatch de modelos:* Dificultad en la alineación entre los modelos de memoria de OpenMP y la concurrencia nativa de Go.

*Overhead:* Posible impacto en el rendimiento debido a la sobrecarga del código transpilado frente a implementaciones manuales.

*Recursos:* Restricciones de tiempo y experiencia del equipo investigador.

*Identificación de los riesgos del proyecto*

#figure(
  table(
    columns: 2,
    stroke: 0.5pt,
    fill: (col, row) => if row == 0 { luma(230) },
    align: left,
    [*Riesgo identificado*], [*Estrategia de mitigación*],
    [Coordinación de equipo: Desalineamiento en tareas], [Implementación de metodología Kanban, división clara de trabajo según capacidades y reuniones continuas.],
    [Reproducibilidad experimental: Variabilidad en benchmarks.], [Estandarización del entorno de pruebas y uso de herramientas estándar de benchmarking.],
    [Integración entre componentes: Fallos entre Runtime y Compilador.], [Definición estricta de APIs y contratos desde el inicio; mantenimiento de estructura modular.],
    [Errores de sincronización: Deadlocks o condiciones de carrera.], [Aplicación de metodologías de testing que soporten errores de concurrencia.],
  ),
  numbering: none,
  outlined: false,
)

*Estructura de Descomposición del Trabajo (EDT)*

#figure(
  image("../figures/figB_edt.png"),
  caption: [Estructura de Descomposición del Trabajo (EDT)],
  numbering: none,
  outlined: false,
) <fig:edt-estructura>

*Lista de Tareas*

#figure(
  table(
    columns: 5,
    stroke: 0.5pt,
    fill: (col, row) => if row == 0 { luma(230) },
    align: left,
    [*Tarea / Resultado*], [*Responsable*], [*Duración (días)*], [*Esfuerzo (Horas)*], [*Costo (S/.)*],
    [Definición y planificación del proyecto (Tesis 1)], table.cell(colspan: 4)[],
    [EP1.1: Ficha de registro de idea de tesis y asesor], [Ambos], [1], [4h tesista (Total)], [40],
    [EP1.2: Protocolo de revisión. Diseño de Formulario de extracción.], [Ambos], [6], [18h tesista (Total)], [180],
    [EP1.3: Reporte de ejecución de la revisión. Formulario de extracción.], [Ambos], [14], [42h tesista (Total)], [420],
    [E1: Problemática. Marco conceptual/Marco teórico. Estado del Arte.], [Ambos], [5], [15h tesista (Total)], [150],
    [EP2.1: Árbol de objetivos. Objetivo general. Objetivos específicos.], [Ambos], [6], [18h tesista (Total)], [180],
    [E2: Objetivo general. Objetivos específicos. Resultados Esperados. Medios de verificación.], [Ambos], [5], [15h tesista (Total)], [150],
    [E3: Resultados esperados. Herramientas, métodos y procedimientos. Alcance y Limitaciones.], [Ambos], [5], [15h tesista (Total)], [150],
    [E4: Proyecto de fin de Carrera completo incluyendo: todas las correcciones y el Anexo de Plan de Proyectos.], [Ambos], [7], [21h tesista (Total)], [210],
    [Reuniones con el asesor (Tesis 1)], [Ambos], [14], [28h tesista (Total) y 28h asesor], [2800],
    [Sesiones de clases de Proyecto de Tesis 1], [Ambos], [14], [10h tesista (Total)], [100],
    [Levantamiento de observaciones de los entregables (Tesis 1)], [Ambos], [12], [56h tesista (Total)], [560],
    [Exposición final de Proyecto de Tesis 1], [Ambos], [7], [28h tesista (Total)], [280],
    table.cell(colspan: 5)[Ejecución del Proyecto],
    [(OE1) Diseño y Especificación de la Arquitectura], [Ambos], table.cell(colspan: 3)[],
    [OE1.R1: Documento de especificación de directivas y cláusulas], [Ambos], [3], [30h tesista (Total)], [300],
    [OE1.R2: Diseño de la arquitectura de la herramienta], [Ambos], [4], [40h tesista (Total)], [400],
    [OE1.R3: Especificación funcional de la CLI], [Ambos], [2], [20h tesista (Total)], [200],
    [(OE2) Implementación de Infraestructura (Compilador y Runtime)], [Dividido], table.cell(colspan: 3)[],
    [Componente 1 - Paralelismo Estructurado], [Patricia], table.cell(colspan: 3)[],
    [OE2.R4/R5: Implementación de Runtime (Pool, for, barrier, critical)], [Patricia], [9], [85h tesista (Total)], [850],
    [OE2.R7/R8: Implementación de Compilador (AST para parallel, for, reduction)], [Patricia], [12], [120h tesista (Total)], [1200],
    [Componente 2 - Paralelismo Basado en Tareas], [Jorge], table.cell(colspan: 3)[],
    [OE2.R6: Implementación de Runtime (Tasking, taskgroup, Dependencias)], [Jorge], [10], [50h tesista (Total)], [500],
    [OE2.R7/R8: Implementación de Compilador (AST para task, depend, cláusulas)], [Jorge], [10], [85h tesista (Total)], [850],
    [(OE3) Evaluación y Análisis de Resultados], [Ambos], table.cell(colspan: 3)[],
    [OE3.R10: Suite de benchmarks implementada], [Ambos], [4], [40h tesista (Total)], [400],
    [OE3.R11: Informe de evaluación de rendimiento y escalabilidad], [Ambos], [7], [55h tesista (Total)], [550],
    [OE3.R12: Reporte de análisis comparativo (expresividad/productividad)], [Ambos], [4], [25h tesista (Total)], [250],
    [Actividades Generales (Tesis 2)], [Ambos], table.cell(colspan: 3)[],
    [Reuniones con asesor (Tesis 2)], [Ambos], [14], [28h tesista (Total) 28h asesor], [2800],
    [Levantamiento de observaciones (Tesis 2)], [Ambos], [7], [40h tesista (Total)], [400],
    [Sesiones de clase y exposiciones], [Ambos], [12], [10h tesista (Total)], [100],
    [Exposición final de tesis 2 (Preparación)], [Ambos], [7], [28h tesista (Total)], [280],
  ),
  numbering: none,
  outlined: false,
)

*Cronograma del proyecto*

Para facilitar la visualización detallada de la planificación, las dependencias entre tareas y la asignación de recursos, el cronograma relacionado a los avances de resultados se ha desarrollado en formato de tabla y diagrama de Gantt.

A continuación, se proporciona el enlace permanente al cronograma maestro, donde se detallan las fechas de inicio y fin, la duración y los responsables de cada actividad correspondiente al desarrollo de los objetivos específicos definidos:

#link("https://docs.google.com/spreadsheets/d/1D10xEJDti1dd1OBpyHO2SXlBPY3YX12pQ4uDEEwk7J4/edit?usp=sharing")

*Lista de recursos*

*Humanos:* Tesistas (Patricia Cántaro, Jorge Alejandro), Asesor (Prof. Viktor Khlebnikov), Especialista de Computer Systems.

*Equipamiento:* PCs de desarrollo, instancias en la nube para benchmarking .

*Herramientas:* Go toolchain (go, ast, testing, pprof), Python, Matplotlib, SciPy, Github, Typst, VSCode/Neovim .

*Costeo del proyecto*

#figure(
  table(
    columns: 4,
    stroke: 0.5pt,
    fill: (col, row) => if row == 0 { luma(230) },
    align: left,
    [*Descripción*], [*Cantidad*], [*Valor Unitario (S/.)*], [*Monto Total (S/.)*],
    [Horas de tesistas], [\~1000 horas], [15.00], [15’000.00],
    [Horas de asesor], [\~60 horas], [100.00], [6’000.00],
    [Especialista de Computer Systems], [\~20 horas], [100.00], [2’000.00],
    [Computadoras personales], [2], [5’000.00], [10’000.00],
    [Instancias Cloud], [1], table.cell(colspan: 2)[100.00],
    table.cell(colspan: 3)[Total], [33’100.00],
  ),
  numbering: none,
  outlined: false,
)

== Anexo C. Especificación Técnica de Directivas y Cláusulas

Documento técnico que contiene la definición formal de la gramática y semántica de las directivas de paralelismo adaptadas para GompherMP. Incluye el detalle de los constructos estructurados, de tareas y las cláusulas de gestión de datos necesarias para la transformación del código.

#link("https://drive.google.com/drive/folders/1v6BKVVQJG4hcHQkgWt4tcRvVjhmtbXFP?usp=drive_link")

== Anexo D. Diseño y Arquitectura de la Herramienta GompherMP

Documento de arquitectura donde se especifica cómo se compone la herramienta GompherMP, sus módulos, el flujo general y las características principales del módulo principal.

#link("https://drive.google.com/drive/folders/14wSehD3LK5fQpDDOM7J97Ka4pu70he1X?usp=sharing")

== Anexo E. Especificación Funcional de la Interfaz de GompherMP

Documento de referencia para la interacción con la herramienta a través de la terminal.

#link("https://drive.google.com/drive/folders/1SyDSrzh93VLD5Ia9fNRnz3Zbuhabm4s5?usp=drive_link")

== Anexo F. Validaciones de Experto

Carpeta de validaciones del experto en programación paralela para los resultados específicos R1, R2, R3, R9, R10, R11 y R12.

#link("https://drive.google.com/drive/folders/18CR_kgMkpq5EWXXH_LhFh6ei-MOzmptT?usp=drive_link")

== Anexo G. Repositorio del Código Fuente

Este anexo proporciona el acceso al repositorio oficial de control de versiones de la herramienta GompherMP.

#link("https://github.com/zet79/gomphermp")

== Anexo H. Cobertura de Pruebas del Módulo de Gestión de Goroutines

Documento técnico de cobertura de pruebas del módulo de gestión de goroutines y reparto de trabajo correspondiente al resultado R4. Incluye las  pruebas unitarias ejecutadas, las primitivas verificadas (Parallel, For, ParallelFor, ForDynamic y Sections) y los resultados cuantitativos obtenidos mediante las herramientas nativas de cobertura de Go.

#link("https://drive.google.com/drive/folders/1sSyETXlrd0ZIiSJ--a7X05ysg5Jmyu0d?usp=drive_link")

== Anexo I. Cobertura de Pruebas de los Mecanismos de Sincronización

Documento técnico de cobertura de pruebas del módulo de mecanismos de sincronización correspondiente al resultado R5. Incluye las pruebas unitarias ejecutadas, las primitivas verificadas (Critical, Single, Master y Barrier) y los resultados cuantitativos obtenidos mediante las herramientas nativas de cobertura de Go.

#link("https://drive.google.com/drive/folders/1mFvz1WowOUsMPsgsaAGyQKANYIiQthYt?usp=drive_link")

== Anexo J. Cobertura de Pruebas del módulo de Tareas

Documento técnico de cobertura de pruebas del módulo de soporte para paralelismo de tareas correspondiente al resultado R6. Incluye las pruebas unitarias ejecutadas, las primitivas verificadas (Task, Taskwait, Taskgroup, Taskloop y TaskWithDepend) y los resultados cuantitativos obtenidos mediante las herramientas nativas de cobertura de Go.

#link("https://drive.google.com/file/d/1x756HAXtZuinyMadj8NQ5XmUdUpHCqCD/view?usp=sharing")

== Anexo K. Informe de cobertura del módulo Parser

Este anexo contiene el informe técnico de cobertura de pruebas del módulo Parser correspondiente al resultado R7. El documento detalla la suite completa de pruebas ejecutada, la trazabilidad entre cada directiva y cláusula de la especificación R1 y las pruebas que verifican su análisis sintáctico, las validaciones semánticas implementadas y los resultados cuantitativos obtenidos mediante las herramientas nativas de cobertura de Go.

#link("https://drive.google.com/drive/folders/1nwgdgIVw349f4FEeI03hwK2Inl35lclJ?usp=sharing")
