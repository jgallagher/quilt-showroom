function Quilt(id, canvas) {
    var quilt = {};
    var margin = 32;
    var ctx = canvas[0].getContext('2d');

    quilt.spacing = 16;
    quilt.shapes = {};

    quilt.init = function(data) {
        var i;

        console.log(data);
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
        canvas.setLayerGroup("polys", {
            click: function(layer) {
                       console.log("click on " + layer.name);
                   },
        });

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

    quilt.addPoly = function(poly) {
        var args = {
            method: 'drawLine',
            fillStyle: '#' + poly.Color,
            strokeStyle: '#000',
            strokeWidth: 1,
            closed: true,
            layer: true,
            name: poly.Id,
            group: 'polys',
            click: function() {},
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
                canvas.setLayer(poly.Id, { fillStyle: pattern });
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
};
