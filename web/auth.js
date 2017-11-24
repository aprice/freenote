'use strict';

var User = {
	user: null,
	userPanel: null,
	loginPanel: null,
	editModal: null,

	init: function () {
		if (this.user == null && window.localStorage.getItem("user")) {
			this.user = JSON.parse(window.localStorage.getItem("user"));
		}

		this.userPanel = $id("User");
		this.loginPanel = $id("Login");
		this.editModal = $id("EditUserModal");

		var self = this;
		$1(".confirmButton", this.loginPanel).addEventListener("click", function(evt) {
			self.login();
		});
		$id("Logout").addEventListener("click", function(evt) {
			self.logout();
		});
		$id("EditUser").addEventListener("click", function(evt) {
			self.editUser();
		});
		$1(".confirmButton", this.editModal).addEventListener("click", function(evt) {
			self.editUser();
		});
		$1(".cancelButton", this.editModal).addEventListener("click", function(evt) {
			self.dismissUserModal();
		});

		restRequest({
			url: "/session",
		}).then(function (payload) {
			self.user = payload;
			self.refresh();
			NoteList.loadNotes();
		}).catch(function (error) {
			App.log(error);
			if (error.status == 401) {
				App.log("Not logged in");
				self.user = null;
				self.refresh();
			} else {
				var msg = error.responseText || error.statusText;
				App.error("Authentication failed: " + msg);
			}
		});
	},

	loggedIn: function() {
		return this.user != null;
	},

	refresh: function () {
		if (this.user == null) {
			window.localStorage.removeItem("user");
			hide(this.userPanel);
			show(this.loginPanel);
		} else {
			window.localStorage.setItem("user", JSON.stringify(this.user));
			$id("WhoAmI").innerText = this.user.username;
			hide(this.loginPanel);
			show(this.userPanel);
		}
	},

	login: function() {
		var user = $id("UsernameField").value;
		var pass = $id("PasswordField").value;
		$id("UsernameField").value = "";
		$id("PasswordField").value = "";
		var self = this;
		restRequest({
			method: "POST",
			url: "/session",
			body: "username=" + encodeURIComponent(user) + "&password=" + encodeURIComponent(pass),
			ctype: "application/x-www-form-urlencoded",
		}).then(function(payload) {
			self.user = payload;
			self.refresh();
			NoteList.loadNotes();
		});
	},

	logout: function() {
		var self = this;
		restRequest({method: "DELETE", url: "/session"}).then(function(payload) {
			self.user = null;
			DataStore.delete();
			NoteList.notes = null;
			NoteEditor.currentNote = null;
			NoteEditor.mode = "md";
			NoteEditor.refresh();
			NoteList.refresh();
			self.refresh();
		});
	},

	editUser: function() {
		$id("ConfirmPasswordField").setCustomValidity("");
		show(this.editModal);
	},

	dismissUserModal: function() {
		hide(this.editModal);
	},

	saveUser: function() {
		var pw = $id("NewPasswordField").value;
		var conf = $id("ConfirmPasswordField").value;
		if (pw != conf) {
			$id("ConfirmPasswordField").setCustomValidity("Confirm password must match password.");
			return;
		}
		restRequest({
			method: this.user._links.password.method,
			url: this.user._links.password.href,
			ctype: "text/plain",
			body: pw,
		}).then(function(payload) {
			App.message("Password updated.");
		});
		this.dismissUserModal();
	},
};