
= Construcciones de Paralelismo Estructurado

*Sintaxis Formal:*

== Directiva parallel
Define una región paralela, instanciando un equipo de goroutines.

*Sintaxis Formal:*

#figure(
```go
//gompher parallel [private(list) | firstprivate(list) | shared(list)]
bloque
```,
caption: [Ejemplo de sintaxis]
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

```go
//gompher for [private(list) | firstprivate(list)]
bucle_for_canonico
```

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

== Directiva sections
Define un conjunto de bloques de trabajo independientes distribuibles.

*Sintaxis Formal:*

```go
//gompher sections [private(list) | firstprivate(list)]
{
    //gompher section
    bloque
    [//gompher section
    bloque]...
}
```

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

```go
//gompher single [private(list) | firstprivate(list)]
bloque
```

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

```go
//gompher master
bloque
```

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

```go
//gompher critical [nombre_opcional]
bloque
```

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

```go
//gompher barrier
```

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

```go
//gompher atomic [read | write | update]
bloque
```
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

```go
//gompher for schedule(kind[, chunk_size])
bloque
```

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

=== Caso 3: Uso de schedule guided

#figure(
  ```go
//gompher parallel
{
    //gompher for schedule(guided)
    for i := 0; i < 100; i++ {
        trabajo(i)
    }
}
  ```,
  caption: [Uso de schedule guided]
)

*Explicación:* Similar a dynamic pero los chunks comienzan grandes y se van reduciendo progresivamente hasta llegar a 1. El tamaño de cada chunk se calcula como las iteraciones restantes divididas entre el número de goroutines. Ofrece un balance entre el bajo overhead de static y la flexibilidad de dynamic.
