#let thesis-template(body) = {
  set document(
    title: "Implementación de paralelismo basado en directivas adaptando el estándar OpenMP al lenguaje Go",
  )
  set page(
    paper: "a4",
    margin: (top: 2.5cm, bottom: 2.5cm, left: 3cm, right: 2.5cm),
    numbering: "1",
    number-align: center,
  )
  set text(font: "Times New Roman", size: 12pt, lang: "es")
  set par(justify: true, leading: 1.5em, spacing: 1.5em)
  set heading(numbering: none)

  show heading.where(level: 1): it => {
    pagebreak(weak: true)
    v(1em)
    text(size: 14pt, weight: "bold")[#it.body]
    v(1em)
  }
  show heading.where(level: 2): it => {
    v(0.8em)
    text(size: 13pt, weight: "bold")[#it.body]
    v(0.6em)
  }
  show heading.where(level: 3): it => {
    v(0.6em)
    text(size: 12pt, weight: "bold")[#it.body]
    v(0.4em)
  }
  show heading.where(level: 4): it => {
    v(0.4em)
    text(size: 12pt, weight: "bold", style: "italic")[#it.body]
    v(0.3em)
  }

  show figure.where(kind: table): it => {
    set figure.caption(position: top)
    it
  }

  body
}
