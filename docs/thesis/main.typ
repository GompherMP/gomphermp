#import "template.typ": thesis-template
#show: thesis-template

#include "frontmatter/title_page.typ"

#align(center, text(size: 14pt, weight: "bold")[TABLA DE CONTENIDO])
#v(1em)
#outline(title: none, depth: 3)
#pagebreak()

#align(center, text(size: 14pt, weight: "bold")[ÍNDICE DE TABLAS])
#v(1em)
#outline(title: none, target: figure.where(kind: table))
#pagebreak()

#align(center, text(size: 14pt, weight: "bold")[ÍNDICE DE FIGURAS])
#v(1em)
#outline(title: none, target: figure.where(kind: image))
#pagebreak()

#include "chapters/01_general.typ"
#include "chapters/02_background.typ"
#include "chapters/03_state_of_art.typ"
#include "chapters/04_results.typ"
#include "chapters/05_implementation.typ"
#include "chapters/06_evaluation.typ"
#include "backmatter/references.typ"
#include "backmatter/appendices.typ"
