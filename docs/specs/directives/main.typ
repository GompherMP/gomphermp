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

#set par(justify: true)

#v(1em)

#include "sections/00_summary.typ"

#v(1em)
#outline(title: "Índice")

#pagebreak()

#include "sections/01_introduction.typ"
#include "sections/02_structured_parallelism.typ"
#include "sections/03_task_parallelism.typ"
#include "sections/04_data_clauses.typ"