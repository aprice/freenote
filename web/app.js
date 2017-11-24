'use strict';

var App = {
	debug: false,

	messagePanel: null,
	confirmationPanel: null,

	init: function () {
		this.messagePanel = $id("StatusMessage");
		this.confirmationPanel = $id("Confirmation");
		
		if (typeof(Version) != "undefined") {
			$id("AppVersion").innerText = Version;
		}

		User.init();
		NoteList.init();
		NoteEditor.init();
	},

	log: function (data) {
		if (this.debug) {
			console.log(data);
		}
	},

	message: function (message, isHtml = false) {
		var mp = this.messagePanel;
		if (isHtml) mp.innerHTML = message;
		else mp.innerText = message;
		mp.cssClass = "";
		mp.style.display = "block";
		setTimeout(function () { fade(mp, 800) }, 3000);
	},

	error: function (message, isHtml = false) {
		var mp = this.messagePanel;
		if (isHtml) mp.innerHTML = message;
		else mp.innerText = message;
		mp.cssClass = "error";
		mp.style.display = "block";
		setTimeout(function () { fade(mp, 800) }, 3000 );
	},

	confirm: function(message, isHtml = false, confirmCB = null, cancelCB = null) {
		var cp = this.confirmationPanel;
		if (isHtml) $1(".message", cp).innerHTML = message;
		else $1(".message", cp).innerText = message;
		cp.style.display = "block";
		$1(".confirmButton", cp).onclick = function() {
			fade(cp, 600);
			if (confirmCB) confirmCB();
		};
		$1(".cancelButton", cp).onclick = function () {
			fade(cp, 600);
			if (cancelCB) cancelCB();
		};
	},
};

window.addEventListener("load", function () {
	App.init();
});
