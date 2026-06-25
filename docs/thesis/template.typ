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

  set heading(numbering: (..nums) => {
    let n = nums.pos()
    if n.len() != 1 {
    //   "Capítulo " + str(n.at(0)) + "."
    // } else {
      numbering("1.1.1.1.", ..n)
    }
  })


  set figure.caption(separator: [. ])
  set figure(numbering: "1")

  show heading.where(level: 1): it => {
    pagebreak(weak: true)
    v(1em)
    set text(size: 20pt, weight: "bold")
    // Level 1 carries "Capítulo N." written by hand in the
    // title's body. That's why we only show it.body
    it.body
    v(1em)
  }
  show heading.where(level: 2): it => {
    v(0.6em)
    set text(size: 16pt, weight: "bold")
    it
    v(0.6em)
  }
  show heading.where(level: 3): it => {
    v(0.6em)
    set text(size: 14pt, weight: "bold")
    it
    v(0.6em)
  }
  show heading.where(level: 4): it => {
    v(0.6em)
    set text(size: 12pt, weight: "bold", style: "italic")
    it
    v(0.6em)
  }

  // Las tablas dentro de un figure deben poder partirse entre páginas (si no,
  // una tabla más alta que la página se desborda). Con table.header(...) la
  // cabecera se repite automáticamente en cada página.
  show figure.where(kind: table): set block(breakable: true)
  show figure.where(kind: table): it => {
    set figure.caption(position: top)
    it
  }

  // Entradas del índice con sangría explícita y fija por nivel (1.5em). Se evita
  // el método indented() por defecto, que alinea los cuerpos según el ancho de
  // los prefijos de los hermanos: como "Anexos" tiene hijos numerados y
  // "Referencias" no, ese cálculo los desalineaba. Aquí todas las entradas de un
  // mismo nivel arrancan en la misma posición y con el mismo peso.
  //
  // El prefijo (número) solo se muestra si tiene contenido real: los capítulos,
  // al usar la función de numeración que en nivel 1 no devuelve nada, dan un
  // prefijo VACÍO ([]), que no es none; sin este filtro recibían un h(0.5em) de
  // más y quedaban 0.5em a la derecha de "Anexos" (numbering: none -> prefijo none).
  show outline.entry: it => {
    set text(weight: "regular")
    let prefix = it.prefix()
    let has-num = prefix != none and prefix != []
    link(
      it.element.location(),
      pad(left: 1.5em * (it.level - 1), {
        if has-num {
          prefix
          h(0.5em)
        }
        it.body()
        box(width: 1fr, it.fill)
        h(0.4em)
        it.page()
      }),
    )
  }

  body
}
