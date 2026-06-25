#let thesis-template(body) = {
  set document(
    title: "Implementación de paralelismo basado en directivas adaptando el estándar OpenMP al lenguaje Go",
  )
  set page(
    paper: "a4",
    margin: (top: 2.5cm, bottom: 2.5cm, left: 2.75cm, right: 2.75cm),
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
    // Level 1 has "Capítulo N." hardcoded in the body; render body only.
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

  // Table figures must break across pages; otherwise a tall table overflows.
  // table.header(...) repeats the header on every page.
  show figure.where(kind: table): set block(breakable: true)
  show figure.where(kind: table): it => {
    set figure.caption(position: top)
    // Sticky caption: stays with the table so it isn't orphaned on a break.
    show figure.caption: set block(sticky: true)
    it
  }

  // Outline entries: fixed 1.5em indent per level, uniform weight. The default
  // indented() aligns bodies by sibling prefix width, which misaligned "Anexos"
  // (numbered children) vs "Referencias" (none).
  //
  // Show the prefix only if it has real content: chapters use a numbering
  // function that returns nothing at level 1, giving an EMPTY prefix ([], not
  // none); without this check it got an extra h(0.5em) and shifted right.
  show outline.entry: it => {
    set text(weight: "regular")
    let prefix = it.prefix()
    let has-num = prefix != none and prefix != []
    let tail = {
      it.body()
      box(width: 1fr, it.fill)
      h(0.4em)
      it.page()
    }
    // With a visible number (Tabla N / 1.1.): two-column grid so wrapped lines
    // hang-indent under the title. Without one (chapters/Referencias/Anexos):
    // body starts at the level margin.
    let row = if has-num {
      grid(columns: (auto, 1fr), column-gutter: 0.5em, prefix, tail)
    } else {
      tail
    }
    link(it.element.location(), pad(left: 1.5em * (it.level - 1), row))
  }

  body
}
