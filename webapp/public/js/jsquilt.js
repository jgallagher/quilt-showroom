/*globals $ console document */ 

function Quilt(width, height, spacing, density) {
    var quilt = {};
    var orig = $('#quilt-canvas');

    quilt.__defineGetter__('density', function() {
        return density;
    });
    quilt.__defineSetter__('density', function(val) {
        density = val;
        quilt.grid = quilt.rebuild_grid();
    });
    quilt.__defineGetter__('spacing', function() {
        return spacing;
    });
    quilt.__defineSetter__('spacing', function(val) {
        spacing = val;
        quilt.grid = quilt.rebuild_grid();
    });

    quilt.shapes = [];
    quilt.add_shape = function(shape) {
        quilt.shapes.push(shape);
    };

    quilt.rebuild_shapes = function() {
        var buf = document.createElement('canvas');
        var ctx = buf.getContext('2d');
        var s;
        var i, j;
        var func;
        buf.width = orig.width();
        buf.height = orig.height();

        for (i = 0; i < quilt.shapes.length; i++) {
            s = quilt.shapes[i];
            func = function() {
                ctx.save();
                if (s.color !== undefined) {
                    ctx.fillStyle = s.color;
                } else {
                    var pattern = ctx.createPattern(s.img, 'repeat');
                    ctx.fillStyle = pattern;
                }
                ctx.beginPath();
                ctx.moveTo(s.poly[0][0], s.poly[0][1]);
                for (j = 1; j < s.poly.length; j++) {
                    ctx.lineTo(s.poly[j][0], s.poly[j][1]);
                }
                ctx.closePath();
                ctx.clip();
                ctx.fill();
                ctx.stroke();
                if (s.img !== undefined) {
                    // we're being called in onload - need to force a redraw
                    // of our parent quilt object
                    quilt.shapes = buf;
                    orig.drawLayers();
                }
                ctx.restore();
            };
            if (s.img !== undefined) {
                s.img.onload = func;
            } else {
                func();
            }
        }

        quilt.shapes = buf;
    };

    quilt.build_overlay = function(width, height, polys) {
        var buf = document.createElement('canvas');
        var ctx = buf.getContext('2d');
        var i;
        var j;
        var vert;
        var p;

        buf.width = width;
        buf.height = height;

        for (i = 0; i < polys.length; i++) {
            p = polys[i];
            if (p.fillStyle !== undefined) {
                ctx.fillStyle = p.fillStyle;
            }
            if (p.strokeStyle !== undefined) {
                ctx.strokeStyle = p.strokeStyle;
            }
            vert = p.vertices;

            ctx.beginPath();
            ctx.moveTo(vert[0], vert[1]);
            for (j = 2; j < vert.length; j += 2) {
                ctx.lineTo(vert[j], vert[j+1]);
            }
            ctx.closePath();
            ctx.fill();
            ctx.stroke();
        }

        quilt.overlay = buf;
    };

    quilt.rebuild_grid = function() {
        var buf = document.createElement('canvas');
        var c = $(buf);
        var x = 0;
        var y;
        var px;
        var py;
        var incr = spacing;
        var pwidth = width * density;
        var pheight = height * density;

        buf.width = orig.width();
        buf.height = orig.height();

        $.jCanvas({ strokeStyle: "#ccc", strokeWidth: 1 });
        for (x = 0; x <= width; x = Math.min(x + incr, width)) {
            px = 0.5 + density*x;
            c.drawLine({
                x1: px, y1: 0.5,
                x2: px, y2: 0.5 + pheight
            });
            if (x === width) {
                break;
            }
        }
        for (y = 0; y <= height; y = Math.min(y + incr, height)) {
            py = 0.5 + density*y;
            c.drawLine({
                x1: 0.5,           y1: py,
                x2: 0.5 + pwidth, y2: py
            });
            if (y === height) {
                break;
            }
        }
        $.jCanvas();
        return buf;
    };

    quilt.draw_overlay = function(ctx, offx, offy) {
        var overlay = quilt.overlay;
        var pos = quilt.overlay_pos;

        if (overlay === undefined || pos === undefined) {
            return;
        }

        ctx.drawImage(overlay, pos.x + offx, pos.y + offy);
    };

    quilt.mousemove = function(x, y) {
        var overlay;
        var owidth;
        var oheight;
        var per_grid;
        var pwidth = width * density;
        var pheight = height * density;

        if (x < 0 || x > pwidth || y < 0 || y > pheight) {
            delete quilt.overlay_pos;
            return;
        }

        overlay = quilt.overlay;
        if (overlay === undefined) {
            return;
        }

        owidth = overlay.width;
        oheight = overlay.height;
        per_grid = spacing * density;

        x = per_grid*Math.floor((x - owidth/2) / per_grid);
        y = per_grid*Math.floor((y - oheight/2) / per_grid);

        if (x + overlay.width > pwidth) {
            x = pwidth - overlay.width;
        }
        if (y + overlay.height > pheight) {
            y = pheight - overlay.height;
        }
        quilt.overlay_pos = { x: Math.max(0, x), y: Math.max(0, y) };
    };

    quilt.grid = quilt.rebuild_grid();

    return quilt;
}
