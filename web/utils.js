'use strict';

var $id = document.getElementById.bind(document);

function $(qry, el) {
	el = el || document;
	return el.querySelectorAll(qry);
}

function $1(qry, el) {
	return $(qry, el)[0];
}

function $el(html) {
	var tpl = document.createElement('template');
	tpl.innerHTML = html;
	return tpl.content.firstChild;
}

function $each(qry, el, fn) {
	if (typeof (el) == "function") {
		fn = el;
		el = null;
	}
	$(qry, el).forEach(fn);
}

function restRequest(options) {
	return new Promise(function(resolve,reject) {
		var r = new XMLHttpRequest();
		if (options.link) {
			options.method = options.link.method;
			options.url = options.link.href;
		}
		options.method = options.method || "GET";
		r.open(options.method, options.url);
		r.setRequestHeader("Accept", "application/json");
		r.onload = function () {
			if (this.status >= 200 && this.status < 300) {
				App.log("Success: " + r.responseText);
				var payload = null;
				if (r.responseText && r.getResponseHeader("Content-Type") == "application/json") {
				payload = JSON.parse(r.responseText);
				}
				resolve(payload);
			} else {
				reject({
					status: this.status,
					statusText: r.statusText
				});
			}
		};
		r.onerror = function () {
			reject({
				status: this.status,
				statusText: r.statusText,
				responseText: r.responseText,
			});
		};
		if (options.headers) {
			Object.keys(options.headers).forEach(function (key) {
				r.setRequestHeader(key, options.headers[key]);
			});
		}
		var params = options.params;
		if (params && typeof params === 'object') {
			options.body = Object.keys(params).map(function (key) {
				return encodeURIComponent(key) + '=' + encodeURIComponent(params[key]);
			}).join('&');
			options.ctype = "application/x-www-form-urlencoded";
		} else if (options.payload) {
			if (options.payload['_links']) delete(options.payload['_links']);
			options.body = JSON.stringify(options.payload);
			options.ctype = "application/json";
		}
		if (options.body) {
			r.setRequestHeader("Content-Type", options.ctype);
		}
		App.log("Sending " + options.method + " " + options.url + ": " + options.body);
		r.send(options.body);
	});
}

function fade(el, ms) {
	if (!el.dataset.faded) {
		el.dataset.faded = true;
		el.dataset.origOpacity = el.style.opacity;
	}
	if ((el.style.opacity -= .05) <= 0) {
		hide(el);
	} else {
		setTimeout(function () { fade(el, ms) }, ms * .05);
	}
}

function hide(el) {
	if (el.style.display == "none") return;
	el.dataset.origDisplay = el.dataset.origDisplay || el.style.display;
	el.dataset.origOpacity = el.dataset.origOpacity || el.style.opacity;
	el.style.display = "none";
}

function show(el) {
	if (el.style.display != "none") return;
	el.style.display = el.dataset.origDisplay || "block";
	el.style.opacity = el.dataset.origOpacity || 1;
}

function uuidToB64(uuid) {
	var bytes = [];
	while (uuid.length >= 2) {
		if (uuid.substring(0, 1) == "-") {
			uuid = uuid.substring(1, uuid.length);
			continue;
		}
		bytes.push(parseInt(uuid.substring(0, 2), 16));
		uuid = uuid.substring(2, uuid.length);
	}
	var b64 = btoa(String.fromCharCode.apply(null, bytes));
	b64 = b64.replace(/\+/g, "-");
	b64 = b64.replace(/\//g, "_");
	b64 = b64.replace(/=/g, "");
	return b64;
}

function getTempID() {
	return "_temp" + (new Date().getTime() + Math.random());
}

function isTempID(id) {
	return id.startsWith("_temp");
}

function appendQuery(url, param) {
	url += (url.includes("?")) ? "&" : "?";
	url += param;
	return url;
}
