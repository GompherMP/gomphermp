
= Capítulo 3. Estado del Arte


== 3.1. Introducción

Para establecer un fundamento sólido para esta investigación, es esencial realizar un análisis riguroso de la literatura existente. La metodología seleccionada para este fin es la revisión sistemática, un enfoque de investigación que utiliza métodos explícitos y sistemáticos para identificar, seleccionar y analizar críticamente la evidencia relevante sobre una pregunta de investigación claramente formulada (Higgins et al., 2024). A diferencia de las revisiones narrativas, este método busca minimizar el sesgo al emplear un proceso transparente y replicable.

La conducción de este estudio se guía por los principios metodológicos establecidos en el Manual Cochrane para Revisiones Sistemáticas de Intervenciones (Higgins et al., 2024). Este manual es considerado el estándar para la ejecución de este tipo de investigaciones, ya que proporciona un marco detallado para cada etapa del proceso, desde la formulación de las preguntas hasta la síntesis e interpretación de los hallazgos. La adhesión a esta guía garantiza el rigor y la solidez de las conclusiones presentadas en este capítulo.


== 3.2. Objetivos de revisión

El objetivo de esta revisión sistemática es realizar un análisis exhaustivo y crítico de la literatura existente sobre la implementación de modelos de paralelismo de alto nivel en lenguajes de programación de propósito general modernos. El propósito principal es comprender la brecha que existe entre los paradigmas de concurrencia nativos de estos lenguajes y las necesidades del cómputo de alto rendimiento, y cómo los enfoques declarativos, como el de OpenMP, buscan cerrarla.

Para cumplir con este objetivo, se llevará a cabo una revisión sistemática de la literatura del estado del arte disponible en bases de datos académicas, conferencias relevantes y repositorios. La revisión se centrará en recopilar información sobre los siguientes objetivos de revisión:

Analizar las limitaciones de los modelos de concurrencia nativos en lenguajes modernos y los principios fundamentales de los paradigmas declarativos (como la separación de incumbencias) que buscan resolverlas.

Examinar las propuestas de implementación existentes, con un enfoque en las arquitecturas de compilador (source-to-source) y librerías de runtime, y sintetizar los retos (sintácticos, semánticos y de rendimiento) reportados en la literatura.

Identificar las metodologías de evaluación (métricas y benchmarks) utilizadas para validar estas implementaciones y consolidar los beneficios reportados en términos de rendimiento, productividad y mantenibilidad del código.


== 3.3. Preguntas de revisión

P1: ¿Cuáles son las limitaciones de los lenguajes de propósito general modernos para problemas de paralelismo de cómputo intensivo?

P2: ¿Qué principios y abstracciones fundamentales proveen los paradigmas de paralelismo declarativo para evitar la fricción semántico-sintáctica en lenguajes de propósito general modernos?

P3: ¿Qué propuestas existen en la literatura para implementar paralelismo declarativo en lenguajes de propósito general modernos?

P4: ¿Qué retos (sintácticos, semánticos, de compilación o de rendimiento) identifica la literatura al intentar implementar paradigmas de paralelismo de alto nivel en lenguajes de propósito general modernos?

P5: ¿Qué conjunto de pruebas y métricas tanto de rendimiento como de expresividad se usan en la literatura para evaluar abstracciones de paralelismo de alto nivel en lenguajes de propósito general modernos?

P6: ¿Qué beneficios y ventajas potenciales se identifican en la literatura al implementar estas abstracciones de paralelismo de alto nivel en lenguajes de propósito general modernos?


== 3.4. Estrategia de búsqueda


=== 3.4.1. Motores de búsqueda a usar

Scopus: Es una base de datos de resúmenes y citas de literatura científica revisada por pares, que abarca revistas, libros y actas de conferencias. Se caracteriza por su cobertura multidisciplinaria y por indexar trabajos de una gran diversidad de editoriales, siendo una herramienta fundamental para el análisis de citas y la investigación científica a gran escala (Elsevier, s.f.).

IEEE Xplore: Es la biblioteca digital del Institute of Electrical and Electronics Engineers (IEEE), una fuente indispensable para el descubrimiento y acceso a literatura técnica en ingeniería eléctrica, ciencias de la computación y electrónica. Ofrece acceso a artículos de revistas, actas de conferencias y estándares técnicos de alta calidad que son fundamentales para la innovación y la investigación en tecnología (IEEE, s.f.).


=== 3.4.2. Cadenas de búsqueda a usar

Para la construcción de las cadenas de búsqueda se utilizaron los siguientes keywords, agrupados por pilar conceptual, que se derivan de las preguntas de investigación especificadas:

Pilar 1 - Lenguajes de Programación: Se enfoca en los lenguajes de propósito general modernos que son objeto del estudio. Se usa OR porque un artículo puede tratar sobre cualquiera de estos lenguajes.

Python

Java

Go / Golang

Rust

Zig

Modern programming language

Pilar 2 - Paradigma de Programación (Paradigm): Describe el enfoque declarativo o basado en directivas para el paralelismo. Son términos intercambiables en este contexto, por lo que se unen con OR.

Pragma

Directive

Declarative

Pilar 3 - Dominio de Aplicación: Define el campo de la computación de alto rendimiento. Se usa OR para incluir términos relacionados.

Parallelism / Parallel

HPC (High-Performance Computing)

Pilar 4 - Implementación: Agrupa los términos que describen la naturaleza de la solución propuesta (una librería, un compilador, etc.), así como estándares de referencia. Se unen con OR para capturar la diversidad de enfoques.

OpenMP

MPI

Compiler / Compiler extension

Library

Framework

Prototype

Pure implementation

Language support

Búsquedas Específicas: Adicionalmente, se incluyeron nombres de implementaciones conocidas como OMP4Py y JOMP para asegurar la recuperación de trabajos altamente relevantes.

A partir de los keywords definidos, se construyó una cadena de búsqueda unificada y adaptada a la sintaxis de cada motor de búsqueda seleccionado. En la tabla que se presenta a continuación, se muestran las cadenas finales utilizadas en cada plataforma.

#figure(
  table(
    columns: 3,
    stroke: 0.5pt,
    table.cell(fill: luma(230))[*Motor de búsqueda*], table.cell(fill: luma(230))[*Cadena de búsqueda*], table.cell(fill: luma(230))[*Cantidad de documentos*],
    [Scopus], [TITLE-ABS-KEY ( ( "OMP4Py" OR "JOMP" OR ( "Zig" AND "OpenMP" ) OR ( "Rust" AND "OpenMP" ) ) OR ( ( "Python" OR "Java" OR "Zig" OR "Rust" OR "Go" OR "Golang" OR "modern programming language" ) AND ( "pragma" OR "directive" OR "declarative" ) AND ( "parallelism" OR "parallel" OR "hpc" ) AND ( "pure implementation" OR "library" OR "compiler" OR "compiler extension" OR "framework" OR "prototype" OR "language support" OR "OpenMP" OR "MPI" ) AND NOT ( "GPU" OR "CUDA" OR "IO" OR "I/O" ) ) )], [112],
    [IEEE Xplore], [(("OMP4Py" OR "JOMP" OR ("Zig" AND "OpenMP") OR ("Rust" AND "OpenMP")) OR (("Python" OR "Java" OR "Zig" OR "Rust" OR "Go" OR "Golang" OR "modern programming language") AND ("pragma" OR "directive" OR "declarative") AND ("parallelism" OR "parallel" OR "hpc") AND ("pure implementation" OR "library" OR "compiler" OR "compiler extension" OR "framework" OR "prototype" OR "language support" OR "OpenMP" OR "MPI") AND NOT ("GPU" OR "CUDA" OR "IO" OR "I/O")))], [27],
  ),
  caption: [Tabla 9. Cadenas de búsqueda usadas],
) <tab:tabla-9-cadenas-de-búsqueda-us>


=== 3.4.3. Criterios de inclusión y exclusión

Criterios de inclusión:

Especificidad Temática: Se incluirán estudios que propongan, implementen o analicen sistemas de paralelismo declarativo o basado en directivas, con un enfoque principal en modelos tipo OpenMP. Se considerarán trabajos aplicados a lenguajes de propósito general modernos (ej. Go, Rust, Zig, etc.) así como trabajos fundacionales en lenguajes de sistemas como C, C++ y Java, que son cruciales para entender la evolución del paradigma.

Enfoque de la Investigación: Se priorizará investigaciones que detallan la arquitectura de compiladores (especialmente source-to-source), librerías de runtime, metodologías de implementación, o herramientas para el análisis de rendimiento de dichos paradigmas.

Ventana Temporal: No se aplicará una ventana temporal estricta para permitir la inclusión de trabajos pioneros y seminales que establecieron las bases del área. Sin embargo, se dará prioridad a trabajos publicados en la última década que aborden los desafíos en lenguajes más modernos.

Idioma: Se incluirán estudios escritos en inglés y español, dado que son los idiomas predominantes en la literatura científica de esta área.

Criterios de exclusión:

Enfoque Temático No Relacionado: Se excluirán los estudios cuyo enfoque principal sea de un área diferente al de las ciencias de la computación o sistemas computacionales.

Enfoque centrado en GPUs y CUDA: Se excluirán los artículos enfocados en el paralelismo masivo de datos de GPUs y CUDA por tratarse de un paradigma significativamente diferente al investigado.

Accesibilidad: Se excluirán los trabajos cuyo texto completo no sea accesible a través de las bases de datos académicas o por medios públicos.

Para refinar la selección de documentos, se aplicó un proceso de filtrado en dos fases sobre los resultados brutos obtenidos de cada motor de búsqueda.

Fase 1: Eliminación de duplicados

Antes de aplicar cualquier criterio de inclusión o exclusión, se realizó una depuración inicial para eliminar duplicados entre Scopus e IEEE Xplore. Para ello se emplearon gestores de referencias, garantizando que cada estudio apareciera una sola vez en el corpus preliminar.

Fase 2: Revisión por título y resumen

Después de eliminar duplicados, se examinaron el título y el resumen de cada artículo recuperado. En esta fase se aplicaron los criterios de inclusión y exclusión para descartar rápidamente los trabajos evidentemente irrelevantes, como aquellos con enfoques temáticos no relacionados o centrados exclusivamente en GPUs y CUDA.

Fase 3: Lectura de texto completo

Los artículos que superaron la revisión inicial fueron sometidos a una lectura exhaustiva del texto completo. En esta etapa se verificó si cada estudio cumplía rigurosamente con los criterios establecidos, prestando especial atención al enfoque de la investigación (p. ej., descripción de arquitectura del compilador, metodologías de evaluación) y la accesibilidad del documento.

El resultado de este proceso de filtrado es el conjunto final de documentos sin duplicados para la extracción de datos, cuya cantidad se detalla en la siguiente tabla.

#figure(
  table(
    columns: 3,
    stroke: 0.5pt,
    table.cell(fill: luma(230))[*Motor de búsqueda*], table.cell(fill: luma(230))[*Cadena de búsqueda*], table.cell(fill: luma(230))[*Cantidad de documentos*],
    [Scopus], [TITLE-ABS-KEY ( ( "OMP4Py" OR "JOMP" OR ( "Zig" AND "OpenMP" ) OR ( "Rust" AND "OpenMP" ) ) OR ( ( "Python" OR "Java" OR "Zig" OR "Rust" OR "Go" OR "Golang" OR "modern programming language" ) AND ( "pragma" OR "directive" OR "declarative" ) AND ( "parallelism" OR "parallel" OR "hpc" ) AND ( "pure implementation" OR "library" OR "compiler" OR "compiler extension" OR "framework" OR "prototype" OR "language support" OR "OpenMP" OR "MPI" ) AND NOT ( "GPU" OR "CUDA" OR "IO" OR "I/O" ) ) )], [14],
    [IEEE Xplore], [(("OMP4Py" OR "JOMP" OR ("Zig" AND "OpenMP") OR ("Rust" AND "OpenMP")) OR (("Python" OR "Java" OR "Zig" OR "Rust" OR "Go" OR "Golang" OR "modern programming language") AND ("pragma" OR "directive" OR "declarative") AND ("parallelism" OR "parallel" OR "hpc") AND ("pure implementation" OR "library" OR "compiler" OR "compiler extension" OR "framework" OR "prototype" OR "language support" OR "OpenMP" OR "MPI") AND NOT ("GPU" OR "CUDA" OR "IO" OR "I/O")))], [7],
  ),
  caption: [Tabla 10. Cadenas de búsqueda usadas],
) <tab:tabla-10-cadenas-de-búsqueda-u>

Cabe resaltar que, según la documentación oficial de IEEE Xplore, la modalidad Command Search permite el uso de operadores booleanos (AND, OR, NOT) para construir consultas personalizadas (IEEE Xplore, s. f.-a). Asimismo, en los ejemplos oficiales proporcionados por la plataforma se evidencia que la búsqueda puede ejecutarse sin especificar un campo único, lo cual sugiere que la expresión booleana se aplica sobre los campos disponibles en la base de datos, incluyendo metadatos y texto completo cuando corresponda (IEEE Xplore, s. f.-b).

En consecuencia, dado que IEEE Xplore no admite un campo combinado equivalente a TITLE-ABS-KEY() de Scopus, en este estudio se empleó una estrategia de búsqueda libre, aplicando la expresión booleana sobre todos los campos relevantes disponibles para garantizar una recuperación integral de documentos


=== 3.4.4. Documentos encontrados

Después de aplicar los criterios de inclusión y exclusión, se seleccionaron 19 documentos cuyas referencias en APA se encuentran a continuación.

#figure(
  table(
    columns: 2,
    stroke: 0.5pt,
    table.cell(fill: luma(230))[*Nro. de Artículo*], table.cell(fill: luma(230))[*Referencia*],
    [1], [Piñeiro, C., & Pichel, J. C. (2026). OMP4Py: A pure Python implementation of OpenMP. Future Generation Computer Systems, 175, Article 108035. https://doi.org/10.1016/j.future.2025.108035],
    [2], [Perugini, A., & Kosmidis, L. (2025). Evaluation of the parallel features of Rust for space systems. Open Access Series in Informatics (OASIcs), 127, 5:1-5:20. https://doi.org/10.4230/OASIcs.PARMA-DITAM.2025.5],
    [3], [Fan, W., He, T., Lai, L., Li, X., Li, Y., Li, Z., Qian, Z., Tian, C., Wang, L., Xu, J., Yao, Y., Yin, Q., Yu, W., Zhou, J., Zhu, D., & Zhu, R. (2021). GraphScope: A unified engine for big graph processing. Proceedings of the VLDB Endowment, 14(12), 2879–2892. https://doi.org/10.14778/3476311.3476369],
    [4], [Fan, X., Mehrabi, M., Sinnen, O., et al. (2017). Supporting enhanced exception handling with OpenMP in object-oriented languages. International Journal of Parallel Programming, 45, 1366–1389. https://doi.org/10.1007/s10766-016-0474-x],
    [5], [Yoshida, A., Kamiyama, A., & Oka, H. (2017). A task-driven parallel code generation scheme for coarse grain parallelization on Android platform. Journal of Information Processing, 25, 426–437. https://doi.org/10.2197/ipsjjip.25.426],
    [6], [Utting, M., Weng, M.-H., & Cleary, J. G. (2014). The JStar language philosophy. Parallel Computing, 40(2), 35–50. https://doi.org/10.1016/j.parco.2013.11.004],
    [7], [Vikas, Giacaman, N., & Sinnen, O. (2014). Multiprocessing with GUI-awareness using OpenMP-like directives in Java. Parallel Computing, 40(2), 69–89. https://doi.org/10.1016/j.parco.2013.11.005],
    [8], [Yoshida, A., Ochi, Y., & Yamanouchi, N. (2014). Parallel Java code generation for layer-unified coarse grain task parallel processing. IPSJ Online Transactions, 7, 168–178. https://doi.org/10.2197/ipsjtrans.7.168],
    [9], [Xiaowen, L. (2014). Research on multi-core PC parallel computation based on OpenMP. International Journal of Multimedia and Ubiquitous Engineering, 9(7), 131–140. https://doi.org/10.14257/ijmue.2014.9.7.12],
    [10], [Senghor, A., & Konate, K. (2012). A Java hybrid compiler for shared memory parallel programming. En 2012 13th International Conference on Parallel and Distributed Computing, Applications and Technologies (pp. 131-136). IEEE. https://doi.org/10.1109/PDCAT.2012.21],
    [11], [Kacs, D., Lee, J., Zarins, J., & Brown, N. (2024). Pragma driven shared memory parallelism in Zig by supporting OpenMP loop directives. En SC24-W: Workshops of the International Conference for High Performance Computing, Networking, Storage and Analysis (pp. 930-938). IEEE. https://doi.org/10.1109/SCW63240.2024.00132],
    [12], [Alexandrov, A., Krastev, G., & Markl, V. (2019). Representations and optimizations for embedded parallel dataflow languages. ACM Transactions on Database Systems, 44(1), Article 4. https://doi.org/10.1145/3281629],
    [13], [Fernandez, R. C., Garefalakis, P., & Pietzuch, P. (2016). Java2SDG: Stateful big data processing for the masses. En 2016 IEEE 32nd International Conference on Data Engineering (ICDE) (pp. 1390-1393). IEEE. https://doi.org/10.1109/ICDE.2016.7498352],
    [14], [Senghor, A., & Konate, K. (2012). Transforming an incorrectly synchronized parallel program into correctly synchronized and well optimized program. En 2012 2nd IEEE International Conference on Parallel, Distributed and Grid Computing (pp. 633-638). IEEE. https://doi.org/10.1109/PDGC.2012.6449894],
    [15], [Powers, F. E., Jr., & Alaghband, G. (2007). The Hydra parallel programming system. Concurrency and Computation: Practice and Experience, 20, 1–27. https://doi.org/10.1002/cpe.1205],
    [16], [Guitart, J., Torres, J., Ayguadé, E., & Bull, J. M. (2001). Performance analysis tools for parallel Java applications on shared-memory systems. En International Conference on Parallel Processing, 2001 (pp. 357-364). IEEE. https://doi.org/10.1109/ICPP.2001.952081],
    [17], [Kambites, M. E., Obdržálek, J., & Bull, J. M. (2001). An OpenMP-like interface for parallel programming in Java. Concurrency and Computation: Practice and Experience, 13, 793–814. https://doi.org/10.1002/cpe.579],
    [18], [Brunschen, C., & Brorsson, M. (2000). OdinMP/CCp—a portable implementation of OpenMP for C. Concurrency: Practice and Experience, 12(12), 1193–1203. https://doi.org/10.1002/1096-9128(200010)12:12\<1193::AID-CPE527\>3.0.CO;2-U],
    [19], [Ramirez, R., & Santosa, A. (2003). A methodology for concurrent and distributed Java applications. En Proceedings International Parallel and Distributed Processing Symposium. IEEE. https://doi.org/10.1109/IPDPS.2003.1213264],
  ),
  caption: [Tabla 11. Lista de documentos encontrados],
) <tab:tabla-11-lista-de-documentos-e>

Aunque no fue recuperado por las cadenas automáticas iniciales, se incluye el trabajo de Madridejos Zamorano (2015), localizado mediante búsqueda manual en el Archivo Digital UPM. Se trata de literatura gris académica (tesis de grado en repositorio institucional), pertinente y de alta relevancia para esta tesis porque documenta explícitamente la integración de funcionalidades OpenMP en el lenguaje Go, proporcionando descripciones de diseño e implementación que ilustran la fricción semántico-sintáctica entre un enfoque declarativo (OpenMP) y el modelo concurrente idiomático de Go. Su disponibilidad en un repositorio universitario con metadatos y enlace persistente asegura trazabilidad y verificabilidad, por lo que se incorpora como evidencia complementaria al corpus principal.


== 3.5. Formulario de extracción de datos

#figure(
  table(
    columns: 3,
    stroke: 0.5pt,
    table.cell(fill: luma(230))[*Campo*], table.cell(fill: luma(230))[*Descripción*], table.cell(fill: luma(230))[*Pregunta*],
    [N° de Artículo], [Identificador del artículo.], [General],
    [Título del artículo], [Título del artículo publicado.], [General],
    [Autor(es)], [Autor(es) o responsable(s) del artículo.], [General],
    [Año de publicación], [Año de publicación del artículo.], [General],
    [Idioma], [Idioma en el que se publicó el artículo], [General],
    [Base de datos], [Base de datos donde se encuentra o indexa el artículo], [General],
    [Abstracto], [Resumen del artículo.], [General],
    [Enlace de Consulta], [URL del artículo.], [General],
    [Limitaciones del Modelo Nativo], [Descripción de las debilidades o insuficiencias del modelo de concurrencia nativo de un lenguaje de propósito general moderno para el paralelismo de cómputo.], [P1],
    [Principios del Paradigma Declarativo], [Explicación de los mecanismos y abstracciones (ej. separación de incumbencias, gestión de datos) que el estudio describe para evitar la fricción semántico-sintáctica.], [P2],
    [Propuesta de Implementación], [Detalles sobre la arquitectura o procedimiento que el estudio propone para implementar un sistema de paralelismo declarativo.], [P3],
    [Lenguaje(s) Objetivo], [Lenguaje(s) de propósito general moderno a los que se aplica la propuesta.], [P3],
    [Retos Identificados], [Dificultades (sintácticas, semánticas, de compilación, de rendimiento) que el estudio identifica al implementar el nuevo paradigma.], [P4],
    [Metodología de Evaluación], [Descripción de los conjuntos de pruebas (benchmarks) y métricas (de rendimiento y/o expresividad) que el estudio utiliza para su evaluación.], [P5],
    [Beneficios Identificados], [Ventajas o mejoras (en productividad, mantenibilidad, escalabilidad, etc.) que el estudio reporta o sugiere al adoptar el nuevo paradigma en su implementación.], [P6],
    [Repositorios de implementación], [Colocar enlace del repositorio donde se encuentra el código de implementación si existe.], [General],
    [Comentarios], [Notas y comentarios en relación al artículo.], [General],
  ),
  caption: [Tabla 12. Formulario de extracción de datos],
) <tab:tabla-12-formulario-de-extracc>

Este formulario se aplicó a cada uno de los artículos seleccionados para la revisión sistemática. El conjunto de datos completo y detallado, producto de esta fase de extracción, se encuentra disponible para su consulta en el Anexo D.


== 3.6. Resultados de revisión


=== 3.6.1. ¿Cuáles son las limitaciones de los lenguajes de propósito general modernos para problemas de paralelismo de cómputo intensivo?

Una limitación recurrente es que las bibliotecas de concurrencia nativas, como los hilos de Java o pThreads, imponen una carga de complejidad significativa sobre el programador (Xiaowen, 2014; Powers & Alaghband, 2007). La implementación manual del paralelismo es un proceso difícil y propenso a errores que a menudo requiere reestructuraciones considerables del código (Vikas et al., 2014). Modelos como Fork/Join en Java, aunque potentes, son difíciles de aplicar a cómputos científicos convencionales que no siguen un patrón de "divide y vencerás", especialmente al considerar las dependencias de datos (Yoshida et al., 2017). Incluso en un lenguaje diseñado para la concurrencia como Go, la implementación manual del paralelismo sigue siendo compleja, pues su modelo de goroutines no equivale directamente a un modelo de ejecución paralela (Madridejos Zamorano, 2015).

Más allá de la complejidad, los lenguajes presentan carencias fundamentales. La limitación más crítica en Python es el Global Interpreter Lock (GIL), que impide la ejecución simultánea de hilos en código intensivo en CPU, neutralizando el beneficio del multihilo (Piñeiro & Pichel, 2026). Java, por su parte, presenta un alto costo en la creación y destrucción de hilos y carece de forma nativa de mecanismos de sincronización de alto rendimiento como las barreras rápidas (Kambites et al., 2001). En otros casos, el lenguaje simplemente carece por completo de soporte para paradigmas clave en HPC, como es el caso de Zig con el paralelismo dirigido por pragmas, lo que constituye una barrera para su adopción en esa comunidad (Kacs et al., 2024).

Finalmente, existen conflictos fundamentales entre los modelos de los lenguajes y los requisitos del paralelismo. Un ejemplo claro es el manejo de excepciones, cuyo diseño para un flujo de control único choca directamente con el modelo de ejecución paralela, llevando a que la especificación OpenMP prohíba construcciones intuitivas como un `try-catch` alrededor de una región paralela (Fan et al., 2016). De manera similar, las plataformas de big data como Spark o Flink imponen un modelo de programación funcional que es un obstáculo para la gran comunidad de desarrolladores de lenguajes imperativos (Castro Fernandez et al., 2016). Este enfoque funcional también puede tratar las funciones de usuario como "cajas negras", lo que impide que el sistema aplique optimizaciones automáticas y delega esta responsabilidad en el programador (Alexandrov et al., 2019).


=== 3.6.2. ¿Qué principios y abstracciones fundamentales proveen los paradigmas de paralelismo declarativo para evitar la fricción semántico-sintáctica en lenguajes de propósito general modernos?

El principio más fundamental y recurrente es la separación de incumbencias (separation of concerns). Este enfoque permite que el desarrollador se concentre en la lógica secuencial del algoritmo (el "qué") sin entrelazarla con los detalles de la ejecución paralela (el "cómo"). El paralelismo se especifica como una capa ortogonal, permitiendo que el programa principal mantenga su estructura original (Ramirez & Santosa, 2003; Utting et al., 2014). El programador declara la intención de paralelizar, mientras que el sistema gestiona de forma transparente la creación de hilos, la sincronización y la planificación, separando así la especificación lógica de la ejecución física (Fan et al., 2021; Yoshida et al., 2017).

Para evitar la fricción sintáctica, estos paradigmas se integran en el lenguaje de forma natural y poco intrusiva. Las estrategias clave incluyen:

Uso de directivas en comentarios o pragmas: Un enfoque clásico, popularizado por OpenMP, es implementar las directivas como comentarios (ej. //omp) o pragmas que son ignorados por los compiladores estándar (Vikas et al., 2014; Kambites et al., 2001). Esto asegura que el código siga siendo una extensión no fundamental del lenguaje y mantenga su compilación y ejecución secuencial, facilitando la portabilidad y el desarrollo incremental (Brunschen & Brorsson, 2000).

Aprovechamiento de la sintaxis nativa: Los enfoques más modernos utilizan características idiomáticas del propio lenguaje para una integración más fluida. Esto incluye el uso de decoradores y gestores de contexto en Python (Piñeiro & Pichel, 2026), anotaciones en Java (Castro Fernandez et al., 2016), o la simple sustitución de iteradores estándar por sus equivalentes paralelos en Rust (Perugini & Kosmidis, 2025).

Para evitar la fricción semántica, el paradigma reemplaza la gestión manual y de bajo nivel de hilos por abstracciones de alto nivel. En lugar de manejar hilos directamente, el programador trabaja con conceptos como:

Iteradores paralelos que se encargan de la partición de datos (Perugini & Kosmidis, 2025).

Cláusulas declarativas que definen el alcance de los datos (private, reduction) y las estrategias de planificación (schedule) (Xiaowen, 2014).

Unidades de trabajo atómicas o dependencias de datos entre tareas (Powers & Alaghband, 2007; Yoshida et al., 2017).

Mediante estas abstracciones, la responsabilidad de la implementación compleja, como la división de la carga de trabajo, la planificación de hilos y la sincronización, se delega completamente a la biblioteca o al runtime, simplificando drásticamente el modelo mental del programador (Perugini & Kosmidis, 2025).


=== 3.6.3. ¿Qué propuestas existen en la literatura para implementar paralelismo declarativo en lenguajes de propósito general modernos?

La literatura documenta un espectro diverso de arquitecturas para implementar paralelismo declarativo en lenguajes modernos, aunque la estrategia predominante ha sido la de los traductores de código fuente a fuente (source-to-source). Este enfoque consiste en un preprocesador o un compilador que transforma el código anotado con directivas en código concurrente nativo del lenguaje anfitrión, el cual a su vez invoca una librería de runtime para la gestión de hilos y la sincronización. Propuestas fundacionales como OdinMP/CCp para C (Brunschen & Brorsson, 2000) y JOMP para Java (Kambites et al., 2001) establecieron este patrón canónico. OdinMP/CCp traduce las directivas OpenMP a código C que utiliza pthreads, extrayendo las regiones paralelas a una función centralizada. De manera análoga, JOMP convierte las directivas en código Java que instancia clases internas para encapsular las regiones paralelas, delegando la ejecución a su librería de runtime. Este modelo ha demostrado ser adaptable y persistente, evolucionando para aplicarse a lenguajes más recientes.

Un ejemplo notable de esta adaptación es la tesis de Madridejos Zamorano (2015), que presenta uno de los primeros prototipos para Go. Su sistema utiliza un preprocesador externo llamado "goprep" para analizar el código fuente y transformar las directivas, especificadas como comentarios (//pragma gomp), en código concurrente que emplea las primitivas idiomáticas de Go: goroutines para la ejecución y channels para la sincronización y comunicación (Madridejos Zamorano, 2015). De forma similar, la implementación para Zig de Kacs et al. (2024) se basa en un preprocesador integrado en el proceso de compilación. Este reescribe el código mediante la técnica de "function outlining" para extraer las regiones paralelas y las convierte en llamadas directas al runtime de OpenMP de LLVM, demostrando una integración más profunda con las cadenas de herramientas existentes. En el ecosistema de Python, OMP4Py (Piñeiro & Pichel, 2026) moderniza este enfoque utilizando características idiomáticas del lenguaje, como decoradores y la manipulación directa del Árbol de Sintaxis Abstracta (AST), para lograr la transformación de manera más elegante y menos intrusiva.

Más allá de los preprocesadores, la literatura explora arquitecturas alternativas que buscan una integración más profunda o dinámica. Hydra (Powers & Alaghband, 2007) representa una desviación significativa de este modelo al proponer un sistema para Java basado en anotaciones y una recompilación en tiempo de ejecución (runtime). En lugar de una traducción estática, Hydra analiza el bytecode anotado en el momento de la carga y lo recompila para generar código paralelo adaptado específicamente a la arquitectura de hardware subyacente. Un enfoque aún más abstracto es el propuesto por Ramirez & Santosa (2003), quienes describen una metodología donde la coordinación no se basa en directivas, sino en un "almacén de restricciones" externo y un lenguaje lógico que define las relaciones temporales entre eventos del programa, logrando una separación casi total entre la lógica funcional y la de concurrencia.


=== 3.6.4. ¿Qué retos (sintácticos, semánticos, de compilación o de rendimiento) identifica la literatura al intentar implementar paradigmas de paralelismo de alto nivel en lenguajes de propósito general modernos?

A nivel semántico y de lenguaje, uno de los desafíos más fundamentales es la adaptación de un estándar como OpenMP, con raíces en C y Fortran, a lenguajes de más alto nivel que no exponen primitivas de bajo nivel. Por ejemplo, lenguajes como Java carecen de acceso directo a instrucciones de hardware, lo que impide una implementación eficiente de directivas como atomic o flush con la semántica de rendimiento que se espera de ellas (Kambites et al., 2001). Este problema persiste en lenguajes modernos; en Zig, las operaciones atómicas nativas no soportan la multiplicación, lo que obliga a los desarrolladores a implementar manualmente la cláusula reduction para este operador (Kacs et al., 2024). Quizás el reto semántico más complejo es el manejo de excepciones, cuyo modelo de flujo de control único es inherentemente incompatible con la ejecución paralela de múltiples hilos (Kambites et al., 2001). La necesidad de proponer nuevos conceptos para gestionar la propagación y sincronización de excepciones en un entorno paralelo es tan significativa que ha sido objeto de investigaciones dedicadas (Fan et al., 2017). Además, algunos lenguajes presentan barreras intrínsecas, siendo el caso más notable el Global Interpreter Lock (GIL) en Python, que impide la ejecución simultánea de hilos en código intensivo en CPU y neutraliza en gran medida el beneficio del paralelismo multihilo (Piñeiro & Pichel, 2026).

Desde la perspectiva de la compilación y las herramientas, el principal obstáculo técnico es a menudo la imposibilidad de modificar directamente el Árbol de Sintaxis Abstracta (AST) de un compilador, como se documentó en el caso de Zig, lo que obliga a recurrir a un enfoque de preprocesador (Kacs et al., 2024). Este enfoque alternativo, sin embargo, introduce su propio reto: la falta de contexto semántico. Al operar antes que el compilador principal, el preprocesador no tiene acceso a información de tipos completa, lo que dificulta la reescritura de código y la gestión de variables, como se evidenció en la propuesta para Go, donde el sistema tenía problemas para reconocer variables declaradas con inferencia de tipo (Madridejos Zamorano, 2015; Kacs et al., 2024).

En el ámbito del rendimiento, los desafíos van más allá de la simple paralelización del trabajo. La literatura reporta que la eficiencia es muy sensible al algoritmo de reparto de trabajo (scheduling), demostrando que no existe una única estrategia que sea óptima para todos los tipos de problemas (Madridejos Zamorano, 2015). Otros retos de rendimiento incluyen el alto costo de arranque (startup overhead) que pueden introducir los sistemas con análisis o compilación en tiempo de ejecución, como se observó en el sistema Hydra (Powers & Alaghband, 2007). Asimismo, el rendimiento puede ser afectado de manera impredecible por las políticas de planificación de hilos del sistema operativo subyacente (Brunschen & Brorsson, 2000). Finalmente, surgen retos de integración al intentar que el modelo de paralelismo coexista con otros sistemas de hilos, como el modelo de despacho de eventos de las interfaces gráficas (GUI), que requiere un manejo especial para evitar interbloqueos y garantizar la capacidad de respuesta de la aplicación (Vikas et al., 2014).


=== 3.6.5. ¿Qué conjunto de pruebas y métricas tanto de rendimiento como de expresividad se usan en la literatura para evaluar abstracciones de paralelismo de alto nivel en lenguajes de propósito general modernos?

La selección de benchmarks se divide principalmente en tres categorías:

Algoritmos Numéricos y de Computación Científica: Esta es la categoría más común. Se emplean suites de benchmarks estandarizadas como NAS Parallel Benchmarks (NPB) (Kacs et al., 2024) y Java Grande Forum Benchmark Suite (Yoshida et al., 2017; Vikas et al., 2014). Además, es frecuente el uso de algoritmos individuales emblemáticos como la Transformada Rápida de Fourier (FFT) (Piñeiro & Pichel, 2026; Perugini & Kosmidis, 2025; Xiaowen, 2014), la multiplicación de matrices (Perugini & Kosmidis, 2025; Utting et al., 2014; Madridejos Zamorano, 2015), la descomposición LU (Piñeiro & Pichel, 2026) y otros cálculos como Montecarlo, Raytracer y la estimación de Pi (Yoshida et al., 2017; Piñeiro & Pichel, 2026).

Aplicaciones de Procesamiento de Datos y del Mundo Real: Para evaluar la aplicabilidad en dominios no puramente científicos, se utilizan benchmarks industriales como LDBC y aplicaciones de producción reales para tareas como monitoreo de ciberseguridad y detección de fraude (Fan et al., 2021). También se incluyen algoritmos de análisis de datos como k-means clustering (Alexandrov et al., 2019), Wordcount (Piñeiro & Pichel, 2026), regresión logística y filtrado colaborativo (Castro Fernandez et al., 2016), así como aplicaciones a medida con interfaces gráficas (GUI) como generadores de fractales y web crawlers para casos de uso especializados (Vikas et al., 2014).

Micro-benchmarks: Un tercer enfoque utiliza pruebas pequeñas y específicas para aislar y medir aspectos concretos del sistema, como la sobrecarga (overhead) pura de las directivas, el costo de la sincronización o el impacto de las herramientas de análisis (Fan et al., 2016; Guitart et al., 2001; Kambites et al., 2001).

Las métricas se pueden clasificar en cuantitativas para el rendimiento y cualitativas para las características del lenguaje.

Métricas de Rendimiento Cuantitativas: Las métricas más fundamentales y utilizadas de forma casi universal son el tiempo de ejecución y el speedup (la aceleración obtenida respecto a una versión secuencial). Prácticamente todos los estudios mencionados los utilizan como su principal indicador de rendimiento (p. ej., Piñeiro & Pichel, 2026; Perugini & Kosmidis, 2025; Kacs et al., 2024; Yoshida et al., 2017). A partir de estas, se derivan otras métricas como la eficiencia paralela (Piñeiro & Pichel, 2026; Perugini & Kosmidis, 2025) y la escalabilidad (Utting et al., 2014). En trabajos que se centran en el costo de la propia abstracción, la métrica clave es la sobrecarga (overhead), medida en tiempo absoluto o como un porcentaje del total (Fan et al., 2016; Guitart et al., 2001).

Métricas Cualitativas de Expresividad y Productividad: La expresividad del lenguaje o la abstracción, la facilidad para expresar ideas complejas de forma clara y concisa se evalúa de forma cualitativa en varios estudios (Perugini & Kosmidis, 2025; Utting et al., 2014; Madridejos Zamorano, 2015). De manera similar, la productividad del programador y la programabilidad se discuten cualitativamente (Perugini & Kosmidis, 2025; Fan et al., 2016). En algunos casos, esta evaluación se apoya en datos cuantitativos como la comparación de líneas de código (LoC) necesarias para implementar una solución (Vikas et al., 2014).


=== 3.6.6. ¿Qué beneficios y ventajas potenciales se identifican en la literatura al implementar estas abstracciones de paralelismo de alto nivel en lenguajes de propósito general modernos?

El beneficio más destacado es un aumento drástico en la productividad del desarrollador. Al ofrecer interfaces declarativas de alto nivel (como directivas tipo pragma), estas abstracciones liberan al programador de la gestión manual de hilos, la sincronización y el balanceo de carga (Xiaowen, 2014; Senghor & Konate, 2013). Esto permite que el código paralelo se mantenga muy cercano a su versión secuencial, lo que mejora la legibilidad y la mantenibilidad (Vikas et al., 2014). El resultado es un ciclo de desarrollo más rápido, con una reducción significativa en las líneas de código necesarias para lograr la concurrencia (Vikas et al., 2014) y la automatización de la generación de código paralelo complejo (Yoshida et al., 2017; Yoshida et al., 2014).

A pesar de su alto nivel de abstracción, estas implementaciones demuestran consistentemente un rendimiento altamente competitivo. En múltiples estudios, el rendimiento es comparable e incluso superior al de implementaciones de referencia en lenguajes como C y Fortran, o a la paralelización manual con hilos nativos (Perugini & Kosmidis, 2025; Kacs et al., 2024; Madridejos Zamorano, 2015). Estas ganancias se atribuyen a optimizaciones automáticas, un mejor balanceo de carga dinámica y la capacidad de los compiladores modernos para realizar una vectorización automática más eficaz (Yoshida et al., 2017; Perugini & Kosmidis, 2025). Los sistemas unificados también logran un rendimiento superior al eliminar la sobrecarga de integrar múltiples herramientas especializadas (Fan et al., 2021).

Estas abstracciones mejoran significativamente la fiabilidad y la correctitud del software paralelo. Al gestionar la concurrencia de forma automática, previenen por diseño clases enteras de errores comunes, como los interbloqueos (deadlocks) y las condiciones de carrera (Powers & Alaghband, 2007; Fan et al., 2016). Lenguajes como Rust llevan esto un paso más allá al ofrecer seguridad de memoria garantizada en tiempo de compilación, lo cual es crucial para sistemas críticos (Perugini & Kosmidis, 2025). Además, se introducen mecanismos mejorados de manejo de excepciones que aseguran una terminación controlada de las regiones paralelas, aumentando la robustez general (Fan et al., 2016).

Un beneficio clave es la portabilidad del código y la independencia de la arquitectura. El código escrito con estas abstracciones puede a menudo ejecutarse sin modificaciones en diversas plataformas, desde dispositivos móviles Android hasta servidores multinúcleo (Yoshida et al., 2017; Xiaowen, 2014). También se destaca la compatibilidad con el ecosistema de un lenguaje, permitiendo, por ejemplo, paralelizar código Python que utiliza bibliotecas de terceros (Piñeiro & Pichel, 2026) o interoperar de forma fluida con otros marcos de ciencia de datos como TensorFlow y Spark (Fan et al., 2021; Alexandrov et al., 2019).


== 3.7. Conclusiones

La revisión de la literatura demuestra de manera concluyente que los lenguajes de propósito general modernos, a pesar de su potencia y flexibilidad, presentan limitaciones fundamentales para abordar eficazmente los problemas de paralelismo de cómputo intensivo. La gestión manual de la concurrencia a través de hilos nativos impone una carga cognitiva significativa y es propensa a errores, mientras que barreras a nivel de lenguaje, como el Global Interpreter Lock (GIL) en Python, y la incompatibilidad de paradigmas, como el choque entre el manejo de excepciones y la ejecución paralela, crean una fricción semántico-sintáctica que obstaculiza el desarrollo de software de alto rendimiento.

Para superar estas barreras, la investigación converge en que los paradigmas de paralelismo declarativo ofrecen una solución robusta y eficaz. El principio clave que sustenta estos enfoques es una estricta separación de incumbencias, que permite a los desarrolladores enfocarse en la lógica del algoritmo (el "qué") mientras delegan la complejidad de la ejecución paralela (el "cómo") al sistema. Mediante el uso de abstracciones de alto nivel, como iteradores paralelos y directivas tipo pragma, se logra una integración idiomática y poco intrusiva en el lenguaje anfitrión, preservando la legibilidad y la mantenibilidad del código secuencial original.

La implementación de estas abstracciones se ha explorado principalmente a través de preprocesadores de código fuente (source-to-source), que transforman el código anotado antes de la compilación. Si bien este enfoque ha demostrado ser funcional, la literatura también identifica retos importantes, como la falta de contexto semántico durante la reescritura del código y la necesidad de adaptar las directivas a las limitaciones primitivas de cada lenguaje. Asimismo, se ha observado que el rendimiento puede ser muy sensible a los algoritmos de reparto de trabajo (scheduling), lo que subraya la importancia de la optimización a nivel del sistema de ejecución (runtime).

En conclusión, la adopción de abstracciones de paralelismo de alto nivel en lenguajes de propósito general modernos representa un avance significativo hacia la democratización del cómputo paralelo. Los beneficios son claros y consistentes: un drástico aumento en la productividad del desarrollador, la obtención de un rendimiento competitivo que a menudo iguala o supera a las implementaciones manuales, y una mejora sustancial en la fiabilidad y portabilidad del software. No obstante, los desafíos técnicos en la compilación y la optimización del rendimiento destacan la necesidad de seguir investigando. Los trabajos futuros deberían orientarse hacia una integración más profunda con las cadenas de herramientas de los compiladores y el desarrollo de runtimes más inteligentes y adaptativos para materializar plenamente el potencial de estos paradigmas.

