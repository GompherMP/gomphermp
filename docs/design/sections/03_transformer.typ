= Módulo transformador

El módulo transformador sigue internamente el siguiente flujo expresado como pseudocódigo para mayor claridad:

#figure(
  ```
Por cada NODO en REPRESENTACIÓN_INTERMEDIA:
  Si NODO corresponde a un nodo de sintaxis enriquecida GompherMP:
    TRANSFORMACION = reconocer la transformación correspondiente al NODO

    Si el contexto en REPRESENTACIÓN_INTERMEDIA determina que TRANSFORMACION es válido:
      NODO_TRANSFORMADO = Aplicar algoritmo de reemplazo de código para TRANSFORMACION

      Reemplazar NODO por NODO_TRANSFORMADO en REPRESENTACIÓN_INTERMEDIA

    Si no:
      Terminar ejecución con error    

  Si no:
    ignorar
  ```,
  caption: [Flujo del módulo transformador]
)

== Algoritmos de reemplazo de código

El algoritmo de reemplazo de código al que se hace mención en el pseudocódigo depende de cada TRANSFORMACION soportado por GompherMP. 

A continuación se muestra un ejemplo de cómo se aplica un algoritmo de reemplazo de código simplificado para el constructo `parallel for`.

#figure(
  ```go
//gompher parallel for
for i:=0; i<100; i++ {
  some_task(i)
  another_task(i)
}
  ```,
  caption: [Código Go enriquecido con el constructo GompherMP `parallel for`]
)

Este bloque de código será transformado por el algoritmo en lo siguiente:

#figure(
  ```go 
func dab070c2_ddc5_4834_94e3_9b7a440a2c69(int i){
  some_task(i)
  another_task(i)
}

gomphermp_runtime.parallel_for(dab070c2_ddc5_4834_94e3_9b7a440a2c69, 100)
  ```,
  caption: [Código Go transformado por el constructo GompherMP `parallel for`]
)

Mediante este ejemplo es posible ver que, por norma general, el proceso de transformación de código consiste en:
- Abstraer el cuerpo de un constructo dentro de una función con nombre no repetible.
- Insertar una función de runtime que recibe como argumento la función previa.

== Funciones de runtime

Para explicar las funciones de runtime, se continuará con el ejemplo previo con una implementación inocente (naive) de `parallel_for`:

#figure(
  ```go
func parallel_for(body func(int) int, iterations int) {
  chunk_size = iterations / GOMPHERMP_THREADS
  
  for thread := range GOMPHERMP_THREADS {
    go func(){
      offset = thread * chunk_size
      for i := range chunk_size {
        body(offset + i)
      }
    }
  }
}
  ```,
  caption: [Implementación de la función de runtime `parallel_for`]
)

Como se puede observar, las funciones de runtime tienden a ser las que realizan el trabajo pesado de GompherMP. En este ejemplo, parallel_for es una función de orden mayor que reparte un trabajo (`body`) de varias iteraciones (`iterations`) entre varios threads para que sea realizado de forma paralela en lugar de secuencial.