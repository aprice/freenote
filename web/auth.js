function Login() {
	var user = $id("UsernameField").value;
	var pass = $id("PasswordField").value;
	App.rest({
		method: "POST",
		url: "/session",
		body: "username=" + encodeURIComponent(user) + "&password=" + encodeURIComponent(pass),
		ctype: "application/x-www-form-urlencoded",
		success: function (payload) {
			App.user = payload;
			App.userRefresh();
			LoadNotes();
		}
	});
}

function Logout() {
	App.rest({
		method: "DELETE",
		url: "/session",
		success: function (payload) {
			App.user = null;
			App.noteList = null;
			App.currentNote = null;
			App.userRefresh();
			App.noteListRefresh();
			App.noteRefresh();
		}
	});
}

function EditUser() {
	$id("ConfirmPasswordField").setCustomValidity("");
	$id("EditUserModal").style.display = "block";
}

function DismissUserModal() {
	$id("EditUserModal").style.display = "none";
}

function SaveUser() {
	var pw = $id("NewPasswordField").value;
	var conf = $id("ConfirmPasswordField").value;
	if (pw != conf) {
		$id("ConfirmPasswordField").setCustomValidity("Confirm password must match password.");
		return
	}
	App.rest({
		method: "PUT",
		url: "/users/" + App.user.id + "/password",
		ctype: "text/plain",
		body: pw,
		success: function (payload) {
			App.message("Password updated.");
		}
	});
	DismissUserModal();
}

window.addEventListener("load", function() {
	App.rest({
		url: "/session",
		success: function (payload) {
			App.user = payload;
			App.userRefresh();
			LoadNotes();
		},
		failed: function (status) {
			if (status == 401) {
				App.log("Not logged in");
				App.user = null;
				App.userRefresh();
			} else {
				App.error("Authentication failed: " + r.responseText);
			}		}
	});
});
