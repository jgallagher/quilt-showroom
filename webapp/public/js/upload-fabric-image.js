filepicker.setKey('AlTJ9WeNMRyujGjGoZ5gMz');

//Setup Aviary
var featherEditor = new Aviary.Feather({
	//Get an api key for Aviary at http://www.aviary.com/web-key
	apiKey: 'd5776eb93',
	apiVersion: 2,
	onSave: function(imageID, newURL) {
		$('#image-fabrics').append('<li class="span4"><div class="thumbnail"><img src="'+newURL+'"/><p>'+$('#fabricname').val()+'</p></div></li>');
		$('#no-image-fabrics').hide();
		$.post("upload-fabric", {
			"name": $('#fabricname').val(),
			"url": newURL
		});
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

		//Launching the Aviary Editor
		featherEditor.launch({
			image: preview,
			url: fpfile.url
		});
	});
};
