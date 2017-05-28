var $ = document.querySelectorAll.bind(document);
var $id = document.getElementById.bind(document);

var App = {
	debug: true,
	user: null,

	init: function () {
		this.userPanel = $id("User");
		this.loginPanel = $id("Login");
		this.messagePanel = $id("StatusMessage");
	},

	userRefresh: function () {
		if (this.user == null) {
			this.userPanel.style.display = "none";
			this.loginPanel.style.display = "block";
		} else {
			$id("WhoAmI").innerText = this.user.username;
			this.loginPanel.style.display = "none";
			this.userPanel.style.display = "block";
		}
	},

	log: function (data) {
		if (this.debug) {
			console.log(data);
		}
	},

	message: function (message, isHtml = false) {
		if (isHtml) this.messagePanel.innerHTML = message;
		else this.messagePanel.innerText = message;
		this.messagePanel.cssClass = "";
		this.messagePanel.style.display = "block";
		this.fadeMessage();
	},

	error: function (message, isHtml = false) {
		if (isHtml) this.messagePanel.innerHTML = message;
		else this.messagePanel.innerText = message;
		this.messagePanel.cssClass = "error";
		this.messagePanel.style.display = "block";
		this.fadeMessage();
	},

	fadeMessage: function() {
		var s = this.messagePanel.style;
		var fade = function() {(s.opacity -= .05) <= 0 ? s.display = "none" : setTimeout(fade, 40)};
		s.opacity = 1;
		setTimeout(fade, 3000);
	},

	userPanel: null,
	loginPanel: null,
	messagePanel: null
};

window.addEventListener("load", function () {
	App.init();
	App.message("Initialized");
});