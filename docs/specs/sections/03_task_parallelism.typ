
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


== Directiva taskloop

Distribuye las iteraciones de un bucle generando una tarea por cada chunk de iteraciones. Combina el comportamiento de task y for en una sola directiva.

*Sintaxis Formal:*

```go
//gompher taskloop [grainsize(n) | firstprivate(list) | private(list)]
bloque
```

=== Caso 1: Generación Automática de Tareas

#figure(
  ```go
//gompher parallel
{
    //gompher single
    {
        //gompher taskloop
        for i := 0; i < 10; i++ {
            procesar(i)
        }
    }
}
  ```,
  caption: [Uso básico de taskloop]
)
*Explicación:* Se genera automáticamente una tarea por cada iteración del bucle. El uso de single garantiza que solo una goroutine genere las tareas, evitando duplicados.

=== Caso 2: Control de Granularidad

#figure(
  ```go
//gompher parallel
{
    //gompher single
    {
        //gompher taskloop grainsize(5)
        for i := 0; i < 100; i++ {
            procesar(i)
        }
    }
}
  ```,
  caption: [Uso de grainsize en taskloop]
)

*Explicación:* Con grainsize(5) se generan 20 tareas de 5 iteraciones cada una, reduciendo el overhead de crear 100 tareas individuales.

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
