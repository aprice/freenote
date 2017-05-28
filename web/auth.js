function Login() {
	var user = $id("UsernameField").value;
	var pass = $id("PasswordField").value;
	var r = new XMLHttpRequest();
	r.open("POST", "/session", true);
	r.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
	r.onreadystatechange = function () {
		if (r.readyState != 4) return;
		if (r.status < 400) {
			App.log("Success: " + r.responseText);
			App.user = JSON.parse(r.responseText);
			App.userRefresh();
		} else {
			App.error("Authentication failed: " + r.responseText);
		}
	};
	r.send("username=" + encodeURIComponent(user) + "&password=" + encodeURIComponent(pass));
}

function Logout() {
	var r = new XMLHttpRequest();
	r.open("DELETE", "/session", true);
	r.onreadystatechange = function () {
		if (r.readyState != 4) return;
		if (r.status < 400) {
			App.log("Success: " + r.responseText);
			App.user = null;
			App.userRefresh();
		} else {
			App.error("Logout failed: " + r.responseText);
		}
	};
	r.send();
}

window.addEventListener("load", function() {
	var r = new XMLHttpRequest();
	r.open("GET", "/session", true);
	r.onreadystatechange = function () {
		if (r.readyState != 4) return;
		if (r.status < 400) {
			App.log("Success: " + r.responseText);
			App.user = JSON.parse(r.responseText);
			App.userRefresh();
		} else if (r.status == 401) {
			App.log("Not logged in");
			App.user = null;
			App.userRefresh();
		} else {
			App.error("Authentication failed: " + r.responseText);
		}
	};
	r.send();
});