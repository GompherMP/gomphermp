
= Construcciones de Paralelismo Estructurado

== Directiva parallel
Define una región paralela, instanciando un equipo de goroutines.

*Sintaxis Formal:*

#figure(
  ```go
  //gompher parallel [private(list) | firstprivat(list) | shared(list)]
  bloque
  ```,
  caption: [Gramática formal de la directiva parallel]
)

=== Caso 1: Región Básica

#figure(
  ```go
  func main() {
      //gompher parallel
      {
          fmt.Println("Hola desde el equipo paralelo")
      }
  }
  ```,
  caption: [Creación de región paralela]
)

*Explicación:* Se crea un equipo de N goroutines. Cada una ejecuta el bloque de impresión de manera concurrente. Al finalizar el bloque, existe una barrera implícita donde la goroutine maestra espera a las demás.


=== Caso 2: Gestión de Datos (Private vs Shared)

#figure(
  ```go
  var global int = 10
  var local int = 5

  //gompher parallel shared(global) private(local)
  {
      // 'local' es una nueva variable (valor 0 o basura)
      local = 1
      // 'global' es la misma dirección de memoria para todos
      global = global + local
  }
  // Al salir, 'local' original sigue siendo 5. 'global' ha cambiado.
  ```,
  caption: [Alcance de variables]
)

*Explicación:* Este ejemplo ilustra la diferencia de memoria. `shared` mantiene la referencia original, mientras que `private` crea una instancia aislada en el stack de cada goroutine.

== Directiva for
Distribuye las iteraciones de un bucle entre las goroutines del equipo actual.

*Sintaxis Formal:*

#figure(
  ```go
  //gompher for [private(list) | firstprivate(list) | lastprivate(list) | reduction(op:list) | schedule(kind[, chunk])]
  bucle_for_canonico
  ```,
  caption: [Sintaxis de la construcción de bucle paralelo (for)]
)

=== Forma canónica del bucle (`bucle_for_canonico`)

Las directivas de bucle (`for`, `parallel for`, `taskloop`) requieren que el `for` anotado tenga *forma canónica*, de modo que el número de iteraciones sea calculable antes de ejecutar y el espacio se pueda repartir entre las goroutines del equipo:

#figure(
  ```go
  for v := lb; v relop b; paso {
      // cuerpo; v no se reasigna aquí
  }
  ```,
  caption: [Forma canónica del bucle]
)

Reglas:
- *Inicialización:* `v := lb`, con una única variable de inducción. `lb` es invariante.
- *Condición:* `v relop b`, con la inducción a la izquierda y `relop` ∈ `{<, <=, >, >=}`. `b` es invariante.
- *Paso:* `v++`, `v--`, `v += c` o `v -= c` (con `c` invariante). La dirección del paso debe concordar con la condición (ascendente con `<`/`<=`, descendente con `>`/`>=`).
- *Inducción inmutable:* el cuerpo no reasigna `v`.

GompherMP *normaliza* el bucle al espacio `[0, M)` (donde `M` es el número de iteraciones, calculado a partir de `lb`, `b` y el paso) y recupera la variable de inducción dentro del cuerpo con `v := lb +/- k*paso`. La forma estrecha `for v := 0; v < N; v++` se emite directamente sobre `[0, N)` sin remapeo (sin sobrecosto). GompherMP rechaza en compilación los bucles no canónicos.

#figure(
  ```go
  //gompher parallel for reduction(+:sum)
  for i := 1; i <= N; i++ {   // inicio en 1, cota inclusiva
      sum += i                 //   -> normalizado a [0, N), i recuperado
  }
  ```,
  caption: [Forma canónica no trivial: inicio no nulo y cota inclusiva]
)

=== Ejemplo de Reparto Estático

#figure(
  ```go
  var datos [100]int
  //gompher parallel
  {
      //gompher for
      for i := 0; i < 100; i++ {
          datos[i] = i * i
      }
  }
  ```,
  caption: [Reparto de trabajo estático]
)

*Explicación:* El runtime divide el espacio de iteración [0, 100) en bloques (chunks) y asigna cada bloque a una goroutine del equipo existente.

== Directiva parallel for

Aunque en la teoría de lenguajes se define como una construcción combinada (un atajo sintáctico para una región `parallel` que contiene un único bloque `for` en su interior), se documenta en esta sección como una directiva principal debido a su alta frecuencia de uso práctico y legibilidad.

*Sintaxis Formal:*

#figure(
  ```go
  //gompher parallel for [private(list) | firstprivate(list) | shared(list) | schedule(kind[, chunk_size])]
  bucle_for_canonico
  ```,
  caption: [Sintaxis formal de la construcción combinada parallel for]
)

=== Ejemplo de Uso Directo

#figure(
  ```go
  var vectorDestino [1000]int
  
  //gompher parallel for schedule(static, 50)
  for i := 0; i < 1000; i++ {
      vectorDestino[i] = operacion(i)
  }
  ```,
  caption: [Distribución inmediata de iteraciones con parallel for]
)

*Explicación:* Se crea el equipo de goroutines y se distribuye el espacio de iteración del bucle, aplicando las políticas de `schedule` y alcance de variables especificadas.

== Directiva sections
Define un conjunto de bloques de trabajo independientes distribuibles.

*Sintaxis Formal:*

#figure(
  ```go
  //gompher sections [private(list) | firstprivate(list) | lastprivate(list) | reduction(op:list)]
  {
      //gompher section
      bloque
      [//gompher section
      bloque]...
  }
  ```,
  caption: [Gramática para la definición de secciones independientes]
)

La forma combinada `parallel sections` provisiona el equipo y distribuye las secciones en un solo paso; acepta las mismas cláusulas de `sections` más `shared(list)` (heredada del lado `parallel`):

#figure(
  ```go
  //gompher parallel sections [private(list) | firstprivate(list) | lastprivate(list) | reduction(op:list) | shared(list)]
  ```,
  caption: [Sintaxis de la construcción combinada parallel sections]
)

=== Ejemplo de Paralelismo Funcional

#figure(
  ```go
  //gompher parallel sections
  {
      //gompher section
      { decodificarVideo() }

      //gompher section
      { decodificarAudio() }
  }
  ```,
  caption: [Secciones independientes]
)

*Explicación:* Cada bloque `section` es una unidad de trabajo que se asigna dinámicamente a las goroutines disponibles del equipo.

== Directiva single
Ejecuta el bloque asociado en una única goroutine del equipo.

*Sintaxis Formal:*

#figure(
  ```go
  //gompher single [private(list) | firstprivate(list)]
  bloque
  ```,
  caption: [Sintaxis formal de la directiva de ejecución única (single)]
)

=== Ejemplo de Ejecución Única

#figure(
  ```go
  //gompher parallel
  {
      procesar() // Ejecutado por todos
      //gompher single
      {
          log.Println("Checkpoint") // Ejecutado solo por uno
      }
      // Barrera implícita aquí
  }
  ```,
  caption: [Ejecución única]
)

*Explicación:* Garantiza que el código se ejecute una sola vez, útil para E/S o inicializaciones, sin romper la región paralela.

== Directiva master
Ejecuta el bloque asociado únicamente en la goroutine maestra del equipo. A diferencia de `single`, no implica sincronización.

*Sintaxis Formal:*

#figure(
  ```go
  //gompher master
  bloque
  ```,
  caption: [Sintaxis formal de la directiva de ejecución maestra (master)]
)

=== Ejemplo de Ejecución Maestra

#figure(
  ```go
  //gompher parallel
  {
      trabajoParalelo()

      //gompher master
      {
          fmt.Println("Soy el maestro, no espero a nadie")
      }
      // A diferencia de single, NO hay barrera implícita aquí.
      // Las otras goroutines continúan inmediatamente.

      masTrabajo()
  }
  ```,
  caption: [Uso de master sin barrera]
)

*Explicación:* El bloque es ejecutado solo por la goroutine con ID 0 (maestra). Las demás goroutines saltan el bloque y continúan su ejecución sin esperar en una barrera.

== Directiva critical
Garantiza exclusión mutua para el bloque asociado.

*Sintaxis Formal:*

#figure(
  ```go
  //gompher critical [nombre_opcional]
  bloque
  ```,
  caption: [Gramática de declaración para regiones críticas]
)

=== Ejemplo de Protección de Recurso

#figure(
  ```go
  var contador int
  //gompher parallel
  {
      //gompher critical
      {
          contador++
      }
  }
  ```,
  caption: [Uso de critical]
)

*Explicación:* El runtime serializa el acceso al bloque, previniendo condiciones de carrera en variables compartidas.

== Directiva barrier
Especifica un punto de sincronización explícito.

*Sintaxis Formal:*

#figure(
  ```go
  //gompher barrier
  ```,
  caption: [Sintaxis de la directiva de barrera explícita]
)

=== Ejemplo de Sincronización Global

#figure(
  ```go
  //gompher parallel
  {
      inicializarDatosLocales()

      //gompher barrier

      // Todos esperan a que la inicialización termine antes de seguir
      procesarDatos()
  }
  ```,
  caption: [Uso de barrier explícito]
)

*Explicación:* Todas las goroutines del equipo deben alcanzar la directiva `barrier` antes de que cualquiera de ellas pueda continuar la ejecución más allá de ese punto.

== Directiva atomic
Garantiza que una expresión simple sobre una variable compartida se ejecute de forma atómica, sin interrupciones de otras goroutines.

*Sintaxis Formal:*

#figure(
  ```go
  //gompher atomic [read | write | update]
  bloque
  ```,
  caption: [Sintaxis general para operaciones de memoria atómicas]
)
=== Caso 1: Ejemplo de Update

#figure(
  ```go
var contador int
//gompher parallel
{
    //gompher atomic update
    contador++
}
  ```,
  caption: [Uso de atomic update]
)

*Explicación:* Protege la operación de modificación sobre contador. A diferencia de critical, permite que distintas goroutines operen sobre distintas variables en paralelo, siendo más eficiente.

=== Caso 2: Ejemplo de Read

#figure(
  ```go
  var x int64
  var v int64
  //gompher parallel
  {
      //gompher atomic read
      v = x
  }
  ```,
  caption: [Uso de atomic read]
)

*Explicación*: Garantiza que la lectura de x sea atómica, evitando que una goroutine lea un valor parcialmente escrito por otra.

=== Caso 3: Ejemplo de Write

#figure(
  ```go
  var x int64
  //gompher parallel
  {
      //gompher atomic write
      x = 42
  }
  ```,
  caption: [Uso de atomic write]
)

*Explicación*: Garantiza que la escritura sobre x sea atómica, evitando que otra goroutine lea un valor a medio escribir.

== Directiva schedule
Controla cómo se distribuyen las iteraciones de un for paralelo entre las goroutines del equipo, agrupándolas en chunks.

*Sintaxis Formal:*

#figure(
  ```go
  //gompher for schedule(kind[, chunk_size])
  bloque
  ```,
  caption: [Gramática de la cláusula de planificación de iteraciones (schedule)]
)

=== Caso 1: Uso de schedule static

#figure(
  ```go
  //gompher parallel
  {
      //gompher for schedule(static, 10)
      for i := 0; i < 100; i++ {
          trabajo(i)
      }
  }
  ```,
  caption: [Uso de schedule static]
)

*Explicación:* Las iteraciones se dividen en chunks de tamaño 10 y se asignan a las goroutines en round-robin antes de ejecutar. Si no se especifica chunk_size, las iteraciones se dividen en bloques aproximadamente iguales. Ideal cuando todas las iteraciones tienen un costo computacional similar.

=== Caso 2: Uso de schedule dynamic

#figure(
  ```go
  //gompher parallel
  {
      //gompher for schedule(dynamic, 5)
      for i := 0; i < 100; i++ {
          trabajoPesado(i)
      }
  }
  ```,
  caption: [Uso de schedule dynamic]
)

*Explicación:* Cada goroutine toma un chunk de 5 iteraciones y cuando termina solicita otro, hasta que no queden iteraciones disponibles. Ideal cuando las iteraciones tienen costos variables, evitando que goroutines queden ociosas esperando a las más lentas. Si no se especifica chunk_size, el valor por defecto es 1.

// === Caso 3: Uso de schedule guided

// #figure(
//   ```go
// //gompher parallel
// {
//     //gompher for schedule(guided)
//     for i := 0; i < 100; i++ {
//         trabajo(i)
//     }
// }
//   ```,
//   caption: [Uso de schedule guided]
// )

// *Explicación:* Similar a dynamic pero los chunks comienzan grandes y se van reduciendo progresivamente hasta llegar a 1. El tamaño de cada chunk se calcula como las iteraciones restantes divididas entre el número de goroutines. Ofrece un balance entre el bajo overhead de static y la flexibilidad de dynamic.
