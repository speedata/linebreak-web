const go = new Go();
WebAssembly.instantiateStreaming(fetch("linebreak.wasm"), go.importObject).then((result) => {
    go.run(result.instance);
    allCookies = document.cookie;
    init();
});

const canvas = document.querySelector("canvas");
const ctx = canvas.getContext("2d");

function init() {
    document.cookie.split("; ").forEach(function (elt) {
        const split = elt.split("=");
        const idname = split[0]
        const field = document.getElementById(idname)
        if (field != null) {
            const value = split[1];
            field.value = value;
            switch (idname) {
                case "hsize":
                        rangeValue.innerText = value;
                    break;
                case "hyphenate":
                    document.getElementById(idname).checked = (value == "true");
                    break;
                default:
                    break;
            }
            if (idname == "hsize") {
            }
        }
    })
    drawtext();
}

function drawtext() {
    var zoom = document.getElementById("zoom").value;
    var hsize = document.getElementById("hsize").value
    canvas.width = hsize * zoom;
    ctx.clearRect(0, 0, hsize, canvas.height);
    var obj = {
        text: document.getElementById("rendertext").value,
        hsize: hsize,
        fontsize: document.getElementById("fontsize").value,
        leading: document.getElementById("leading").value,
        tolerance: document.getElementById("tolerance").value,
        hyphenpenalty: document.getElementById("hyphenpenalty").value,
        hyphenate: document.getElementById("hyphenate").checked,
    }
    const items = ["fontsize", "hyphenpenalty", "leading", "hsize", "tolerance","hyphenate"]
    items.forEach(function (item, index) {
        document.cookie = item + "=" + obj[item];
    });
    var posinfo = getpositions(obj);
    canvas.height = posinfo.height * zoom;
    var fnt = obj.fontsize + "px Garamond";
    ctx.font = fnt
    ctx.scale(zoom, zoom);
    posinfo.positions.forEach(element => {
        ctx.fillText(element.char, element.xpos, element.ypos);
    });
    var tbl = document.getElementById("rtable");
    if (tbl != null) {
        tbl.remove();
    }
    var table = document.createElement('table');
    table.setAttribute("class","table")
    table.setAttribute("id","rtable")
    var tablediv = document.getElementById('tablediv');
    tablediv.append(table)

    posinfo.lines.forEach(function(row) {
      var tr = table.insertRow(); //Create a new row
      var tdlinenumber = tr.insertCell();
      var tdr = tr.insertCell();
      tdlinenumber.innerText = row.line;
      tdr.innerText = row.r;

    });
}


function clearCookies() {
    document.cookie.split(";").forEach(function (c) { document.cookie = c.replace(/^ +/, "").replace(/=.*/, "=;expires=" + new Date().toUTCString() + ";path=/"); });
}