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

#show figure.caption: set align(center)

#show figure.where(kind: table): set align(center)
#show figure.where(kind: table): set figure.caption(position: top)
#show figure.where(kind: table): set figure(supplement: [Tabla])

#show figure.where(kind: raw): set align(left)
#show figure.where(kind: raw): set figure(supplement: [Bloque de código])

#show figure.where(kind: image): set align(center)
#show figure.where(kind: image): set figure(supplement: [Figura])

#align(center)[
  #text(17pt, weight: "bold")[Informe de Cobertura de Pruebas — Módulo de Mecanismos de Sincronización de GompherMP]

  #v(1em)
  Jorge David Alejandro Contreras \
  Patricia Natividad Cántaro Márquez \
  17 de mayo de 2026
]

#set par(justify: true)

#v(1em)

#include "sections/00_summary.typ"

#v(1em)
#outline(title: "Índice")

#pagebreak()

#include "sections/01_module.typ"
#include "sections/02_functional_coverage.typ"
#include "sections/03_test_results.typ"
