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
                case "squeezeoverfullboxes":
                    document.getElementById(idname).checked = (value == "true");
                    break;
                case "hangingpunctuationend":
                    document.getElementById(idname).checked = (value == "true");
                    break;
                default:
                    break;
            }
        }
    })
    drawtext();
}

function drawtext() {
    var zoom = document.getElementById("zoom").value;
    var hsize = document.getElementById("hsize").value
    canvas.width = hsize * zoom * 1.3;
    ctx.clearRect(0, 0, hsize, canvas.height);
    var obj = {
        demeritsfitness: document.getElementById("demeritsfitness").value,
        fontsize: document.getElementById("fontsize").value,
        hsize: hsize,
        hyphenate: document.getElementById("hyphenate").checked,
        squeezeoverfullboxes: document.getElementById("squeezeoverfullboxes").checked,
        hangingpunctuationend: document.getElementById("hangingpunctuationend").checked,
        hyphenpenalty: document.getElementById("hyphenpenalty").value,
        leading: document.getElementById("leading").value,
        text: document.getElementById("rendertext").value,
        tolerance: document.getElementById("tolerance").value,
        zoom: document.getElementById("zoom").value,
    }
    const items = [
        "demeritsfitness",
        "fontsize",
        "hsize",
        "hyphenate",
        "hyphenpenalty",
        "leading",
        "squeezeoverfullboxes",
        "hangingpunctuationend",
        "tolerance",
        "zoom",
    ]
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

    ctx.lineWidth = 0.2;
    ctx.beginPath();
    ctx.moveTo(hsize, 0);
    ctx.lineTo(hsize, canvas.height);
    ctx.stroke();

    var tbl = document.getElementById("rtable");
    if (tbl != null) {
        tbl.remove();
    }
    var table = document.createElement('table');
    table.setAttribute("class", "table table-sm")
    table.setAttribute("id", "rtable")
    var tablediv = document.getElementById('tablediv');
    tablediv.append(table)
    var thead = document.createElement("thead");
    var headrow = thead.insertRow();
    var th;
    ["Line", "Adj. ratio", "Total demerits", "Fitness", "Badness"].forEach(function (elt) {
        th = document.createElement("th");
        th.innerText = elt;
        headrow.appendChild(th);
    })
    table.appendChild(thead);
    posinfo.lines.forEach(function (row) {

        var tr = table.insertRow(); //Create a new row

        var tdlinenumber = tr.insertCell();
        tdlinenumber.innerText = row.line;
        var td
        td = tr.insertCell();
        td.innerText = row.r;
        td = tr.insertCell();
        td.innerText = row.demerits;
        td = tr.insertCell();
        td.innerText = row.fitness;

        td = tr.insertCell();
        td.innerText = row.badness;
    });
}


function clearCookies() {
    document.cookie.split(";").forEach(function (c) { document.cookie = c.replace(/^ +/, "").replace(/=.*/, "=;expires=" + new Date().toUTCString() + ";path=/"); });
}