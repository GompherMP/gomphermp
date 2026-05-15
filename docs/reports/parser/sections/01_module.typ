= Descripción del Módulo

El Parser es el componente de entrada del compilador GompherMP. Recibe como entrada el código fuente Go anotado con comentarios de la forma `//gompher <directiva>` y produce un `ParseResult` que contiene el árbol sintáctico abstracto (AST) nativo de Go junto con una lista ordenada de directivas tipadas, cada una vinculada al nodo del AST que anota.

== Ubicación

El módulo reside en el directorio `internal/parser/` del repositorio y consta de los siguientes archivos:

#figure(
  table(
    columns: (auto, 1fr),
    align: (left, left),
    [*Archivo*],          [*Responsabilidad*],
    [`parser.go`],        [Pipeline principal: orquestación, recorrido del AST, validación de tipo de nodo y contexto.],
    [`directive.go`],     [Definición de la interfaz `Directive` y 14 tipos concretos.],
    [`clause.go`],        [Definición de la interfaz `Clause` y 8 tipos concretos.],
    [`clause_parser.go`], [Análisis léxico de cláusulas mediante expresiones regulares.],
    [`parser_test.go`],   [Suite completa de pruebas unitarias y de integración.],
  ),
  caption: [Archivos que componen el módulo Parser],
)

== Pipeline de procesamiento

El flujo interno del parser consta de seis etapas:

+ Análisis del código fuente Go mediante `go/parser.ParseFile`, obteniendo el AST nativo y la lista completa de comentarios.
+ Construcción del `ast.CommentMap` para asociar cada comentario con el nodo del AST que le sigue inmediatamente.
+ Recorrido del AST con `ast.Inspect` para identificar los comentarios `//gompher` y parsearlos como directivas tipadas.
+ Validación semántica de cada directiva: tipo del nodo objetivo, adyacencia con el comentario y validez de las cláusulas declaradas.
+ Validación de contexto jerárquico (por ejemplo, `section` solo puede aparecer dentro de `sections`).
+ Ordenamiento de las directivas por línea de aparición en el código fuente, preservando el orden de ejecución original.

== Metodología de pruebas

La suite de pruebas sigue una organización por granularidad, desde la unidad mínima (parseo de una cláusula individual) hasta la integración completa (parseo de un archivo Go con múltiples directivas). El archivo `parser_test.go` está estructurado en seis secciones que reflejan esta organización:

- *Parseo de cláusulas:* pruebas de la función `extractClauses` con cadenas de texto que representan cláusulas individuales o combinadas.
- *Parseo de directivas:* pruebas de la función `parseDirectiveText` con el texto completo de una directiva (kind + cláusulas), incluyendo casos válidos, rechazo de cláusulas no permitidas y manejo de errores.
- *Integración completa:* pruebas de la función pública `Parse` sobre código fuente Go real con directivas anotadas.
- *Validaciones semánticas:* pruebas de las verificaciones de tipo de nodo, adyacencia de comentarios, contexto jerárquico de `section` y rechazo de cláusulas con listas vacías.
- *Contratos de interfaz:* pruebas directas de los métodos no exportados de las interfaces `Directive` y `Clause` para cada tipo concreto.
- *Pruebas internas auxiliares:* invocaciones directas a funciones no exportadas (`validateClauses`, `makeVarListClause`, `buildDirective`) que cubren rutas de error inalcanzables a través de la API pública.

