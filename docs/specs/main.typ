// Configuración global del documento
#set page(numbering: "1")
#set heading(numbering: "1.1.")
#set text(lang: "es")
#show raw.where(block: true): it => {
  block(fill: luma(240), inset: 12pt, radius: 4pt, width: 100%)[
    #let lines = it.text.split("\n")
    #grid(
      columns: (auto, 1fr),
      column-gutter: 1em,
      row-gutter: 0.5em,
      ..lines.enumerate().map(((i, line)) => (
        text(fill: gray, size: 0.8em)[#(i + 1)],
        raw(line, lang: it.lang)
      )).flatten()
    )
  ]
}

#show figure: set align(left)
#show figure.caption: set align(center)
#set figure(
  supplement: [Bloque de código]
)

#align(center)[
  #text(17pt, weight: "bold")[Especificación Técnica de Directivas y Cláusulas para GompherMP]
  
  #v(1em)
  Jorge David Alejandro Contreras \
  Patricia Natividad Cántaro Márquez \
  19 de marzo de 2026
]

#v(1em)

#align(center)[
  #text(12pt, weight: "bold")[Resumen]
]

#pad(x: 2em)[
  #set par(justify: true)
  Este documento define la especificación técnica de GompherMP. Se detalla la sintaxis formal, se presentan ejemplos de uso para cada construcción y se explican los comportamientos esperados en el runtime.
]

#v(1em)
#outline(title: "Índice")

#pagebreak()

= Introducción y Alcance
Este documento cubre el subconjunto de directivas OpenMP adaptadas a Go. La sintaxis formal utiliza la siguiente notación: 

- `[]`: Elemento opcional.
- `|`: Alternativa exclusiva.
- `list`: Lista de variables separadas por comas.
- `bloque`: Un bloque de código Go delimitado por llaves `{ }`.

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

= Construcciones de Paralelismo de Tareas

== Directiva task
Define una unidad de trabajo explícita y asíncrona.

*Sintaxis Formal:*

```go
//gompher task [depend(tipo:list) | private(list) | firstprivate(list)]
bloque
```

=== Caso 1: Tarea Simple

#figure(
  ```go
//gompher parallel
{
    //gompher single
    {
        //gompher task
        { calculoPesado() }
    }
}
  ```,
  caption: [Generación de tarea]
)

*Explicación:* La tarea se envía a un pool. El uso de `single` es crucial para evitar crear la misma tarea múltiples veces.

=== Caso 2: Captura en Bucles

#figure(
  ```go
//gompher parallel
{
    //gompher single
    {
        for i := 0; i < 10; i++ {
            //gompher task firstprivate(i)
            {
                process(i) // 'i' capturado por valor
            }
        }
    }
}
  ```,
  caption: [Uso de firstprivate]
)

*Explicación:* `firstprivate` fuerza la captura del valor de `i` en el momento de creación, evitando problemas de clausura en bucles.

== Directiva taskwait
Sincroniza la tarea actual esperando a sus tareas hijas directas.

*Sintaxis Formal:*

```go
//gompher taskwait
```

=== Ejemplo de Sincronización Local

#figure(
  ```go
//gompher task
{
    //gompher task
    hijo1()
    //gompher task
    hijo2()

    //gompher taskwait
    fmt.Println("Hijos terminados")
}
  ```,
  caption: [Sincronización de hermanos]
)

*Explicación:* La ejecución se suspende hasta que los hijos directos finalicen.

== Directiva taskgroup
Sincroniza todas las tareas descendientes en su ámbito.

*Sintaxis Formal:*

```go
//gompher taskgroup
bloque
```

=== Ejemplo de Sincronización Profunda

#figure(
  ```go
//gompher taskgroup
{
    //gompher task
    crearArbolRecursivo()
}
  ```,
  caption: [Grupo de tareas]
)

*Explicación:* Garantiza la finalización de todo el subárbol de tareas generado.

== Cláusula de Dependencia (depend)
Define restricciones de orden de ejecución.

*Sintaxis Formal:*

```go
depend(in:list) | depend(out:list) | depend(inout:list)
```

=== Caso 1: Productor-Consumidor

#figure(
  ```go
var x int
//gompher task depend(out:x)
{ x = 1 } // Tarea A

//gompher task depend(in:x)
{ fmt.Println(x) } // Tarea B
  ```,
  caption: [Dependencia Flow (RAW)]
)

*Explicación:* La Tarea B espera a que la Tarea A finalice para asegurar la consistencia de `x`.

=== Caso 2: Cadena de Dependencias

#figure(
  ```go
var buff []byte

//gompher task depend(out:buff)
{ buff = leer() } // Paso 1

//gompher task depend(inout:buff)
{ buff = comprimir(buff) } // Paso 2

//gompher task depend(in:buff)
{ enviar(buff) } // Paso 3
  ```,
  caption: [Encadenamiento con inout]
)

*Explicación:* `inout` serializa el acceso, creando una secuencia de ejecución estricta basada en el flujo de datos del buffer.
