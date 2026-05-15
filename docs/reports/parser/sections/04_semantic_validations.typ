= Validaciones Semánticas

Más allá del análisis sintáctico de cada directiva en aislamiento, el parser implementa cuatro categorías de validación semántica que detectan errores de uso antes de que el código sea transformado. Cada validación cuenta con pruebas dedicadas.

Por consistencia con la sección anterior, los nombres de las pruebas omiten el prefijo `Test`: las pruebas que comienzan con `Parse` corresponden a `TestParse_*` en `parser_test.go`, y las que comienzan con `Clauses` corresponden a `TestExtractClauses_*`.

== Validación de tipo de nodo objetivo

Cada directiva tiene un tipo de nodo Go que puede anotar. Los directivos `for`, `parallel for` y `taskloop` requieren una sentencia `*ast.ForStmt`. El directivo `atomic` requiere `*ast.ExprStmt`, `*ast.AssignStmt` o `*ast.IncDecStmt`. Los directivos restantes (`parallel`, `sections`, `single`, `master`, `critical`, `task`, `taskgroup`, `section`) requieren un `*ast.BlockStmt`. Si el usuario coloca una directiva sobre un nodo incompatible, el parser rechaza el código con un mensaje explícito antes de la fase de transformación.

#figure(
  table(
    columns: (1fr, 1fr),
    align: (left + horizon, left + horizon),
    table.header([*Prueba*], [*Comportamiento verificado*]),

    [`Parse/ForOnNonForStmt`],         [Rechazo de `for` sobre un bloque],
    [`Parse/ParallelForOnNonForStmt`], [Rechazo de `parallel for` sobre un bloque],
    [`Parse/TaskloopOnNonForStmt`],    [Rechazo de `taskloop` sobre un bloque],
    [`Parse/AtomicOnBlockStmt`],       [Rechazo de `atomic` sobre un bloque],
    [`Parse/AtomicOnAssignStmt`],      [Aceptación de `atomic` sobre una asignación],
    [`Parse/ParallelOnNonBlockStmt`],  [Rechazo de `parallel` sobre un bucle `for`],
  ),
  caption: [Pruebas de validación de tipo de nodo objetivo],
)

== Validación de adyacencia (línea en blanco)

Una directiva GompherMP debe estar ubicada exactamente en la línea inmediatamente anterior al bloque de código que anota. Cualquier separación —ya sea una línea en blanco u otro comentario interpuesto— se considera un error porque puede causar que la directiva se asocie silenciosamente con un nodo incorrecto del AST.

Esta regla se verifica mediante la prueba `Parse/BlankLineBetweenDirectiveAndBlock`, que construye un programa con una línea en blanco intencional entre la directiva y su bloque objetivo y comprueba que el parser retorna un error.

== Validación de contexto jerárquico

La directiva `//gompher section` solo tiene sentido dentro del bloque de una directiva `//gompher sections`. El parser realiza una pasada de validación posterior al análisis para verificar mediante contención de posiciones que esta relación se respeta.

Esta regla se verifica mediante la prueba `Parse/SectionOutsideSections`, que coloca una directiva `section` en el cuerpo principal de una función (fuera de cualquier bloque `sections`) y comprueba que el parser la rechaza.

== Validación de cláusulas con paréntesis vacíos

Las cláusulas que reciben listas de variables (`private`, `firstprivate`, `lastprivate`, `shared`) requieren al menos una variable. El parser detecta el caso de paréntesis vacíos y emite un mensaje claro indicando qué cláusula está incompleta, en lugar de un mensaje genérico de "cláusula desconocida".

#figure(
  table(
    columns: (1fr, 1fr),
    align: (left + horizon, left + horizon),
    table.header([*Patrón rechazado*], [*Prueba que la verifica*]),
    [`private()`],                              [`Clauses/EmptyPrivate`],
    [`shared(  )` (con espacios en blanco)],    [`Clauses/EmptyShared`],
  ),
  caption: [Validación de listas de variables vacías],
)

= Conclusión

El módulo Parser alcanza una cobertura del 100% de instrucciones ejecutables, con 90 pruebas que cubren tanto el análisis sintáctico de cada directiva y cláusula soportada como las validaciones semánticas adicionales. La suite verifica el comportamiento correcto sobre código válido y el rechazo explícito de cada combinación inválida documentada en la especificación.
