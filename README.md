# linebreak-web

A playground for the line breaking algorithm of [boxes and glue](https://github.com/boxesandglue/boxesandglue).

## About

linebreak-web typesets a paragraph with boxes and glue and draws the result on an HTML canvas. Sliders and fields change the line width, font size, leading, tolerance, hyphen penalty, fitness demerits and emergency stretch, and switches turn hyphenation, squeezing of overfull lines and hanging punctuation on and off. Every change typesets the paragraph again, so you see at once how a parameter moves the line breaks.

It is a test bed, not a finished tool: the controls follow what is being worked on in boxes and glue.

## Running

You need a Go compiler.

    # build
    GOARCH=wasm GOOS=js go build -o linebreak.wasm
    # serve
    python3 -m http.server
    # open http://localhost:8000/ in a browser

## Technique

A small Go layer calls the boxes and glue library and is compiled to [WebAssembly](https://en.wikipedia.org/wiki/WebAssembly). A short JavaScript interface passes the settings to it and draws each glyph at the position it returns.

## Thanks

Many thanks to Didier Verna for [etap](https://github.com/didierverna/etap), which inspired this playground. etap is far more advanced.
