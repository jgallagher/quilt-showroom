function Quilt(id, canvas, container) {
    var quilt = {};
    var margin = 32;
    var ctx = canvas[0].getContext('2d');

    quilt.spacing = 16;
    quilt.shapes = {};

    quilt.init = function(data) {
        var i;

        console.log(data);
        quilt.id = data.Id;
        quilt.width = data.Width;
        quilt.height = data.Height;
        canvas[0].width = quilt.width + 2*margin;
        canvas[0].height = quilt.height + 2*margin;
        quilt.grid = quilt.buildGrid();

        for (i = 0; i < data.ColorPolys.length; i++) {
            quilt.addPoly(data.ColorPolys[i]);
        }
        for (i = 0; i < data.ImagePolys.length; i++) {
            var p = data.ImagePolys[i];
            p.Color = '999';
            quilt.addPoly(p);
        }
        canvas.setLayerGroup("polys", { click: function(layer) { }, });

        // add the grid layer with index 0 (i.e., underneath the polys)
        canvas.addLayer({
            method: 'draw',
            index: 0,
            fn: function(ctx) {
                ctx.drawImage(quilt.grid, margin, margin);
            },
        });
        canvas.drawLayers();
    };

    quilt.deleteOnClick = function() {
        quilt.disableOverlay();
        canvas.setLayerGroup("polys", {
            click: function(layer) {
                console.log("remove layer " + layer.name + "," + layer.polyid);
                canvas.removeLayer(layer.name);
                canvas.drawLayers();
                $.post("/quilts/"+quilt.id+"/poly-delete", {
                    polyid: layer.polyid,
                });
            },
        });
    };

    quilt.buildOverlay = function(width, height, polys) {
        var buf = document.createElement('canvas');
        var ctx = buf.getContext('2d');
        var i, j;

        // disable other event handlers on polygon clicks
        canvas.setLayerGroup("polys", { click: function(layer) {} });

        buf.width = width;
        buf.height = height;

        for (i = 0; i < polys.length; i++) {
            var p = polys[i];
            ctx.fillStyle = '#999';
            ctx.strokeStyle = '#000';
            ctx.strokeWidth = 1;

            ctx.beginPath();
            ctx.moveTo(p.Coords[0][0], p.Coords[0][1]);
            for (j = 1; j < p.Coords.length; j++) {
                ctx.lineTo(p.Coords[j][0], p.Coords[j][1]);
            }
            ctx.closePath();
            ctx.fill();
            ctx.stroke();
        }

        canvas.removeLayer("overlay");
        canvas.drawImage({
            source: buf,
            layer: true,
            name: "overlay",
            x: margin, y: margin,
            fromCenter: false,
            opacity: 0.7,
            visible: false,
        });
        canvas.drawLayers();
        var mousemove = function(e) {
            var args = {};
            var x = e.pageX - this.offsetLeft + container.scrollLeft() - margin;
            var y = e.pageY - this.offsetTop + container.scrollTop() - margin;

            // hide if we're off the main quilt surface
            if (x < 0 || x > quilt.width || y < 0 || y > quilt.height) {
                args.visible = false;
            } else {
                args.visible = true;
            }

            // snap to grid
            var spacing = quilt.spacing;
            x = spacing * Math.floor((x - width/2) / spacing);
            y = spacing * Math.floor((y - height/2) / spacing);
            if (x + width > quilt.width) {
                x = quilt.width - width;
            }
            if (y + height > quilt.height) {
                y = quilt.height - height;
            }

            args.x = Math.max(0, x) + margin;
            args.y = Math.max(0, y) + margin;
            canvas.setLayer("overlay", args);
            canvas.drawLayers();
        };
        var click = function(e) {
            console.log("got click on canvas");
        };
        canvas.on({mousemove: mousemove, click: click});
        quilt.disableOverlay = function() {
            canvas.removeLayer("overlay");
            canvas.off({mousemove: mousemove, click: click});
            canvas.drawLayers();
        };
    };

    quilt.addPoly = function(poly) {
        var name = "poly-" + poly.Id;
        var args = {
            method: 'drawLine',
            fillStyle: '#' + poly.Color,
            strokeStyle: '#000',
            strokeWidth: 1,
            closed: true,
            layer: true,
            name: name,
            group: 'polys',
            click: function() {},
            polyid: poly.Id,
            index: 0,
        }
        for (i = 0; i < poly.Coords.length; i++) {
            args['x'+(i+1)] = margin + poly.Coords[i][0];
            args['y'+(i+1)] = margin + poly.Coords[i][1];
        }
        if (poly.Url !== undefined) {
            var img = new Image();
            img.src = poly.Url;
            img.onload = function() {
                var pattern = ctx.createPattern(img, 'repeat');
                canvas.setLayer(name, { fillStyle: pattern });
                canvas.drawLayers();
            }
        }
        canvas.addLayer(args);
    };

    quilt.buildGrid = function() {
        var buf = document.createElement('canvas');
        var c = $(buf);
        var x, y;
        var width = quilt.width;
        var height = quilt.height;
        var spacing = quilt.spacing;

        buf.width = width + 1;
        buf.height = height + 1;

        $.jCanvas({ strokeStyle: "#ccc", strokeWidth: 1 });
        for (x = 0; x <= width; x = Math.min(x + spacing, width)) {
            c.drawLine({
                x1: x+0.5, y1: 0.5,
                x2: x+0.5, y2: 0.5 + height });
            if (x === width) {
                break;
            }
        }
        for (y = 0; y <= height; y = Math.min(y + spacing, height)) {
            c.drawLine({
                x1: 0.5,         y1: y+0.5,
                x2: 0.5 + width, y2: y+0.5});
            if (y === height) {
                break;
            }
        }
        $.jCanvas();
        return buf;
    };

    $.get("/quilts/"+id+"/json", quilt.init);
    return quilt;
};

function FormWatcher(quilt) {
    var fw = {};

    fw.watch = function(type, inputs) {
        switch (type) {
            case "rectangle":
                fw.handler = function(e) {
                    var w = inputs.width.val() * quilt.spacing;
                    var h = inputs.height.val() * quilt.spacing;
                    quilt.buildOverlay(w, h,
                            [{Coords: [[0,0],[w,0],[w,h],[0,h],[0,0]]}]);
                };
                break;

            case "triangle":
                fw.handler = function(e) {
                    var w = inputs.width.val() * quilt.spacing;
                    var h = inputs.height.val() * quilt.spacing;
                    var c;
                    switch (inputs.orient.val()) {
                        case "nw": c = [[0,0],[w,0],[0,h],[0,0]];   break;
                        case "ne": c = [[0,0],[w,0],[w,h],[0,0]];   break;
                        case "sw": c = [[0,0],[w,h],[0,h],[0,0]];   break;
                        case "se": c = [[w,0],[w,h],[0,h],[w,0]];   break;
                        case "n":  c = [[0,h],[w/2,0],[w,h],[0,h]]; break;
                        case "e":  c = [[0,0],[w,h/2],[0,h],[0,0]]; break;
                        case "w":  c = [[w,0],[0,h/2],[w,h],[w,0]]; break;
                        case "s":  c = [[0,0],[w,0],[w/2,h],[0,0]]; break;
                    }
                    quilt.buildOverlay(w, h, [{Coords: c}]);
                };
                break;
        }
        fw.handler(null);
        $.each(inputs, function(key, val) {
            val.on('change', fw.handler);
        });
    };

    fw.reattach = function() {
        fw.handler(null);
    };

    return fw;
}
