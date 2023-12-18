# linebreak-web

**This is purely experimental!**

## About

This is a playground for boxes and glue's line breaking algorithm.

I'd like to have a visual test for the line breaking algorithm of boxes and glue, so I render the result to an HTML canvas.

There will be more bells and whistles following, but I need a first commit, so here it is.

## Running / compiling

You need a Go compiler

    # build
    GOARCH=wasm GOOS=js go build -o linebreak.wasm
    # serve:
    python3 -m http.server
    # open localhost:8000 and then open index.html

## Technique

I use a small Go layer to the boxes and glue library (also written in Go) and compile it to [Web Assembly (wasm)](https://en.wikipedia.org/wiki/WebAssembly). There is a small JavaScript interface to the Web Assembly program.


## Thanks

Many thanks to Didier Verna for [etap](https://github.com/didierverna/etap) that inspired me to create this. etap is far (yes far!) more advanced.


