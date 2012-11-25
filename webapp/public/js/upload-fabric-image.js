filepicker.setKey('AlTJ9WeNMRyujGjGoZ5gMz');

var orig_fpfile = null;

//Setup Aviary
var featherEditor = new Aviary.Feather({
	//Get an api key for Aviary at http://www.aviary.com/web-key
	apiKey: 'd5776eb93',
	apiVersion: 2,
	onSave: function(imageID, newURL) {
		var toremove = orig_fpfile;
		var fabricname = $("#fabricname").val();
		var input = {
			url: newURL,
			filename: "from-aviary.png",
			mimetype: "image/png",
		};
		filepicker.remove(toremove);
		filepicker.store(input, function(fpfile) {
			console.log(fpfile);
			filepicker.stat(fpfile, {width:true, height:true},
				function(metadata) {
					console.log(metadata);
					var w = metadata.width;
					var h = metadata.height;
					var rescale = w;
					if (h > w) {
						rescale = h;
					}
					// rescale so max dimention is 128 pixels
					rescale = 128 / rescale;
					w = Math.round(w * rescale);
					h = Math.round(h * rescale);
					console.log("should rescale to "+w+"x"+h);
					filepicker.convert(fpfile, {
						width: w,
						height: h,
						fit: 'scale',
					}, function(newFpfile) {
						filepicker.remove(fpfile);
						$('#image-fabrics').append('<li class="span4"><div class="thumbnail"><img src="'+newFpfile.url+'"/><p>'+fabricname+'</p></div></li>');
						$('#no-image-fabrics').hide();
						$.post("upload-fabric", {
							"name": fabricname,
							"url": newFpfile.url,
						});
					});
				});
		});
		featherEditor.close();
	},
	appendTo: 'web_demo_pane'
});

//Giving a placeholder image while Aviary loads
var preview = document.getElementById('web_demo_preview');

//When the user clicks the button, import a file using Filepicker.io
var editPane = document.getElementById('start_web_demo');
editPane.onclick = function(){
	filepicker.pick({mimetype: 'image/*'}, function(fpfile) {
		//Showing the preview
		preview.src = fpfile.url;
		orig_fpfile = fpfile;

		//Launching the Aviary Editor
		featherEditor.launch({
			fileFormat: "png",
			image: preview,
			url: fpfile.url
		});
	});
};
