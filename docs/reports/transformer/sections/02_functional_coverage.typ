= Cobertura Funcional

Esta sección demuestra que cada directiva y cláusula del lenguaje cuenta con un
manejador que emite la llamada de runtime correspondiente, y que cada uno está
respaldado por pruebas dedicadas.

== Directivas transformadas

El módulo transforma las catorce directivas del lenguaje hacia las llamadas de
runtime que se indican a continuación:

#figure(
  table(
    columns: (auto, 1fr),
    align: (left, left),
    [*Directiva*], [*Llamada de runtime emitida*],
    [`parallel`],            [`Parallel(func(threadID int) { ... })`],
    [`for`],                 [`For(threadID, func(i int) { ... }, n)` (o `ForDynamic` / `ForStaticChunked` según `schedule`)],
    [`parallel for`],        [`ParallelFor(func(i int) { ... }, n)` (o las variantes `ParallelForDynamic` / `ParallelForStaticChunked`)],
    [`sections`],            [`Sections([]func(){ ... })`],
    [`parallel sections`],   [`ParallelSections([]func(){ ... })`],
    [`single`],              [`Single(func() { ... })`],
    [`master`],              [`Master(threadID, func() { ... })`],
    [`critical`],            [`Critical("", func() { ... })` o `Critical("nombre", func() { ... })`],
    [`barrier`],             [`Barrier()`],
    [`atomic`],              [`AtomicAddInt` / `AtomicStoreInt` / `AtomicLoadInt` según la operación],
    [`task`],                [`Task(func() { ... })` o `TaskWithDepend(...)` con `depend`],
    [`taskwait`],            [`Taskwait()`],
    [`taskgroup`],           [`Taskgroup(func() { ... })`],
    [`taskloop`],            [`Taskloop(func(i int) { ... }, n, grainsize)`],
  ),
  caption: [Mapeo de cada directiva a la llamada de runtime que emite el transformador],
)

== Planificación de bucles

La cláusula `schedule` selecciona el punto de entrada del runtime para las
directivas de bucle. En ausencia de `schedule`, o con `schedule(static)`, se
emite el reparto estático en bloque. `schedule(static, chunk)` se traduce al
reparto estático cíclico por bloques (`ForStaticChunked`), y
`schedule(dynamic[, chunk])` al reparto dinámico bajo demanda (`ForDynamic`).

== Normalización de bucles a su forma canónica

Las directivas de bucle (`for`, `parallel for`, `taskloop`) admiten cualquier
bucle en forma canónica y no solo la forma estrecha `for i := 0; i < N; i++`. El
transformador analiza el bucle y, cuando no es la forma simple, lo normaliza
sobre el espacio `[0, N)` que distribuye el runtime: la closure recibe un
contador de base cero y la variable de inducción del usuario se reconstruye con
`i := lb ± contador * paso` al inicio del cuerpo. La forma estrecha se emite de
forma directa, sin remapeo ni sobrecosto.

#figure(
  table(
    columns: (auto, 1fr),
    align: (left, left),
    [*Forma del bucle*], [*Tratamiento*],
    [`for i := 0; i < N; i++`], [Forma simple: el contador es la propia variable de inducción y el conteo es `N`.],
    [Inicio distinto de cero, `<=`, `>`, `>=`, paso no unitario o descendente], [Forma normalizada: contador de base cero más recuperación de la inducción y conteo calculado.],
    [Paso que no avanza la inducción, condición que no la prueba, o dirección incoherente], [Se rechaza con un diagnóstico, sin generar código incorrecto.],
  ),
  caption: [Tratamiento de las formas de bucle por el transformador],
)

== Directiva atomic

A diferencia del resto de directivas, que reescriben un bloque, `atomic` reescribe
una operación simple sobre una variable. Tiene tres modalidades, cada una
traducida a la primitiva atómica del runtime correspondiente, lo que evita la
sobrecarga de una sección crítica para una sola variable:

#figure(
  table(
    columns: (auto, auto, 1fr),
    align: (left, left, left),
    [*Modalidad*], [*Código de entrada*], [*Llamada emitida*],
    [`update`], [`x += d`], [`AtomicAddInt(&x, d)`],
    [`write`],  [`x = v`],  [`AtomicStoreInt(&x, v)`],
    [`read`],   [`v = x`],  [`v = AtomicLoadInt(&x)`],
  ),
  caption: [Transformación de la directiva atomic según su modalidad],
)
