= Cláusulas de Datos
== Cláusula private
Especifica que cada goroutine tendrá su propia copia local de la variable, independiente de las demás. La copia no se inicializa con el valor original y al finalizar la construcción el valor original no se modifica.

*Sintaxis Formal:*

#figure(
  ```go
  private(list)
  ```,
  caption: [Sintaxis formal de la cláusula de privatización de memoria (private)]
)

=== Caso 1: En Región Paralela

#figure(
  ```go
  x := 10
  //gompher parallel private(x)
  {
      x = obtenerID()
  }
  ```,
  caption: [Uso de private en región paralela]
)

*Explicación:* Cada goroutine recibe su propia copia de x sin inicializar, por lo que no hereda el valor 10. Las modificaciones no afectan al valor original ni a las copias de otras goroutines. Al finalizar, x sigue siendo 10.

=== Caso 2: En Generación de Tareas

#figure(
  ```go
  resultado := 0

  //gompher parallel
  {
      //gompher single
      {
          for i := 0; i < 5; i++ {
              //gompher task private(resultado)
              {
                  resultado = calcular(i)
              }
          }
      }
  }
  ```,
  caption: [Uso de private en generación de tareas]
)

*Explicación:* Cada tarea recibe su propia copia de `resultado` sin inicializar. Sin `private`, todas las tareas compartirían la misma variable causando una condición de carrera. Al finalizar, `resultado` sigue siendo 0.

== Cláusula firstprivate
Superconjunto de `private`. Cada goroutine recibe su propia copia local de la variable, inicializada con el valor original antes de entrar a la construcción.

*Sintaxis Formal:*

#figure(
  ```go
  firstprivate(list)
  ```,
  caption: [Gramática de la cláusula de privatización con inicialización (firstprivate)]
)

=== Caso 1: En Región Paralela

#figure(
  ```go
  x := 10
  //gompher parallel firstprivate(x)
  {
      x = x + obtenerID()
  }
  ```,
  caption: [Uso de firstprivate en región paralela]
)

*Explicación:* Cada goroutine recibe su propia copia de x inicializada con el valor 10. A diferencia de private, la copia sí hereda el valor original. Al finalizar, x sigue siendo 10.

=== Caso 2: En Generación de Tareas

#figure(
  ```go
  base := 100

  //gompher parallel
  {
      //gompher single
      {
          for i := 0; i < 5; i++ {
              //gompher task firstprivate(base)
              {
                  resultado := base + calcular(i)
              }
          }
      }
  }
  ```,
  caption: [Uso de firstprivate en generación de tareas]
)

*Explicación:* Cada tarea captura el valor de `base` (100) en el momento de su creación. Esto es especialmente importante en bucles, donde sin `firstprivate` todas las tareas podrían capturar el valor final de `base` en lugar del valor al momento de creación. Al finalizar, `base` sigue siendo 100.

== Cláusula lastprivate
Superconjunto de `private`. Cada goroutine recibe su propia copia local de la variable, y al finalizar la construcción el valor de la última iteración secuencial se copia al original.

*Sintaxis Formal:*

#figure(
  ```go
  lastprivate(list)
  ```,
  caption: [Sintaxis de la cláusula de privatización con retención de último valor (lastprivate)]
)

=== Caso 1: En Bucle Paralelo

#figure(
  ```go
  x := 0
  //gompher parallel for lastprivate(x)
  for i := 0; i < 10; i++ {
      x = i * 2
  }
  ```,
  caption: [Uso de lastprivate en bucle paralelo]
)
*Explicación:* Cada goroutine trabaja con su propia copia de x. Al finalizar, la copia de la goroutine que ejecutó la última iteración secuencial (i=9) se copia al x original, quedando con valor 18.

=== Caso 2: En Secciones Paralelas

#figure(
  ```go
  x := 0

  //gompher parallel sections lastprivate(x)
  {
      //gompher section
      { x = 1 }

      //gompher section
      { x = 2 }
  }
  ```,
  caption: [Uso de lastprivate en secciones paralelas]
)

*Explicación:* En el contexto de sections, la sección que aparece lexicamente última en el código es la que copia su valor al original. Al finalizar, x vale 2.

== Cláusula shared
Declara que una o más variables son compartidas por todas las goroutines. Todas las referencias a la variable apuntan a la misma dirección de memoria. Es el comportamiento por defecto para variables declaradas fuera de la construcción.

*Sintaxis Formal:*

#figure(
  ```go
  shared(list)
  ```,
  caption: [Gramática de la cláusula de compartición de memoria (shared)]
)

=== Caso 1: En Región Paralela

#figure(
  ```go
  x := 0

  //gompher parallel shared(x)
  {
      //gompher critical
      { x++ }
  }
  ```,
  caption: [Uso de shared en región paralela]
)

*Explicación:* Todas las goroutines acceden a la misma variable x. El programador es responsable de la sincronización — en este caso con critical para evitar condiciones de carrera.

=== Caso 2: En Generación de Tareas

#figure(
  ```go
  resultado := 0

  //gompher parallel shared(resultado)
  {
      //gompher single
      {
          //gompher task shared(resultado)
          { resultado = calcular() }
      }
  }
  ```,
  caption: [Uso de shared en generación de tareas]
)

*Explicación:* La tarea accede a la misma variable resultado del contexto que la creó. El programador debe garantizar que resultado no sea accedida por otras tareas simultáneamente sin sincronización.

== Cláusula reduction

Realiza una reducción sobre una variable compartida usando un operador. Cada goroutine trabaja con su propia copia local inicializada con el valor neutro del operador, y al finalizar todas las copias se combinan con el operador para producir el resultado final.

*Sintaxis Formal:*

#figure(
  ```go
  reduction(operador:list)
  ```,
  caption: [Sintaxis formal de la cláusula de reducción de datos (reduction)]
)

=== Operadores Soportados

#figure(
  table(
    columns: (auto, auto, auto),
    inset: 10pt,
    align: horizon,
    [*Operador*], [*Valor inicial*], [*Descripción*],
    [`+`], [`0`], [Suma],
    [`*`], [`1`], [Producto],
    [`-`], [`0`], [Resta],
    [`&&`], [`1`], [AND lógico],
    [`||`], [`0`], [OR lógico],
    [`max`], [`mínimo del tipo`], [Máximo],
    [`min`], [`mínimo del tipo`], [Mínimo]
  ),
  caption: [Operadores de reducción soportados y sus valores neutros iniciales]
)

=== Caso 1: En Región Paralela

#figure(
  ```go
  suma := 0

  //gompher parallel for reduction(+:suma)
  for i := 0; i < 10; i++ {
      suma += i
  }
  ```,
  caption: [Uso de reduction en región paralela]
)

*Explicación:* Cada goroutine tiene su propia copia de suma inicializada en 0. Cada una acumula su parte del resultado independientemente. Al finalizar el bucle, todas las copias se suman para producir el resultado final correcto en suma, evitando condiciones de carrera.

=== Caso 2: En Generación de Tareas

#figure(
  ```go
  suma := 0

  //gompher parallel
  {
      //gompher single
      {
          for i := 0; i < 10; i++ {
              //gompher task reduction(+:suma)
              { suma += calcular(i) }
          }
      }
  }
  ```,
  caption: [Uso de reduction en generación de tareas]
)

*Explicación:* Cada tarea tiene su propia copia de suma. Al finalizar todas las tareas, las copias se combinan con el operador + para producir el resultado final. A diferencia de usar critical, reduction es más eficiente porque evita serializar el acceso a la variable.



