= Cláusulas de gestión de datos

Esta sección detalla la transformación de las cinco cláusulas de gestión de
datos, la pieza del módulo que va más allá de una traducción directa: cada
cláusula inyecta varias sentencias en el código generado.

El transformador resuelve el tipo de cada variable de cláusula con `go/types`,
de modo que funciona incluso cuando el tipo se infiere de una asignación corta
(`:=`) o del resultado de una llamada. La siguiente tabla resume la técnica de
cada cláusula:

#figure(
  table(
    columns: (auto, 1fr),
    align: (left, left),
    [*Cláusula*], [*Técnica de transformación*],
    [`private(x)`],      [Declara una copia fresca `var x T` al inicio de la closure. Su valor cero tapa a la variable externa del mismo nombre durante la región.],
    [`firstprivate(x)`], [Captura el valor externo una vez antes de la región (`_fp_x := x`) y, dentro de la closure, lo usa para inicializar la copia (`x := _fp_x`).],
    [`shared(x)`],       [No requiere código: las closures de Go capturan por referencia, que es exactamente la semántica compartida. Se acepta y se trata como una operación nula.],
    [`lastprivate(x)`],  [Captura la dirección de la variable antes de la región (`_lp_x := &x`) y le da a la copia privada `var x T`. La iteración secuencialmente última, o la sección léxicamente última, escribe su copia de vuelta a través del puntero (`*_lp_x = x`).],
    [`reduction(op:x)`], [Captura la dirección (`_red_x := &x`), inicializa un acumulador privado con la identidad del operador (`var x T = identidad`), deja que el cuerpo acumule sobre la copia y, al final, combina cada parcial de vuelta a través del puntero bajo una sección crítica (`*_red_x op= x`).],
  ),
  caption: [Técnica de transformación de cada cláusula de gestión de datos],
)

La cláusula `reduction` admite los operadores `+`, `-`, `*`, `&&`, `||`, `max` y
`min`. Para `+`, `-`, `*`, `&&` y `||` el acumulador se inicializa con la
identidad del operador (`0`, `1`, `true` o `false`). Para `max` y `min` el
acumulador parte del valor inicial de la variable, capturado antes de la región,
lo que produce el mismo resultado que la identidad de tipo extremo sin requerir
constantes específicas del tipo.

Cuando una directiva no implementa una cláusula que recibe, esta se rechaza con
un diagnóstico en lugar de descartarse en silencio, de modo que ninguna cláusula
válida pasa inadvertida.

= Conclusión

El módulo transformador alcanza una cobertura del 98.5% de instrucciones
ejecutables, con 190 pruebas que cubren las catorce directivas del lenguaje, las
cinco cláusulas de gestión de datos, la directiva `atomic`, las tres políticas de
planificación de bucles y la normalización de la forma canónica. La suite
completa pasa adicionalmente bajo el detector de carreras de Go. La cobertura no
cubierta corresponde a guardas defensivas para condiciones que el flujo de
ejecución no alcanza en operación normal, y que se conservan por robustez.
