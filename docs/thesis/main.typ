#import "template.typ": thesis-template
#show: thesis-template

#include "frontmatter/title_page.typ"

#outline(title: "TABLA DE CONTENIDO", depth: 3)
#pagebreak()
#outline(title: "ÍNDICE DE TABLAS", target: figure.where(kind: table))
#pagebreak()
#outline(title: "ÍNDICE DE FIGURAS", target: figure.where(kind: image))
#pagebreak()

#include "chapters/01_general.typ"
#include "chapters/02_background.typ"
#include "chapters/03_state_of_art.typ"
#include "chapters/04_results.typ"
#include "backmatter/references.typ"
#include "backmatter/appendices.typ"
