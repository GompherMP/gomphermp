
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