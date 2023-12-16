const go = new Go();
WebAssembly.instantiateStreaming(fetch("add.wasm"), go.importObject).then((result) => {
    go.run(result.instance);
    drawtext(240);
});

const button = document.querySelector("button");

const canvas = document.querySelector("canvas");
const ctx = canvas.getContext("2d");
ctx.font = "20px Garamond";

function drawtext(width) {
    canvas.width = width;
    ctx.font = "20px Garamond";
    ctx.clearRect(0, 0, width, canvas.height);
    getpositions(width).forEach(element => {
        ctx.fillText(element.char,element.xpos,element.ypos);
    });
}
