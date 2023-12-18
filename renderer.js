const go = new Go();
WebAssembly.instantiateStreaming(fetch("linebreak.wasm"), go.importObject).then((result) => {
    go.run(result.instance);
    drawtext();
});

const canvas = document.querySelector("canvas");
const ctx = canvas.getContext("2d");
ctx.font = "20px Garamond";

function drawtext() {
    var width = document.getElementById("inputhsize").value
    canvas.width = width;
    ctx.font = "20px Garamond";
    ctx.clearRect(0, 0, width, canvas.height);
    var obj = {
        text: document.getElementById("rendertext").value,
        width: width,
    }
    getpositions(obj).forEach(element => {
        ctx.fillText(element.char,element.xpos,element.ypos);
    });
}
