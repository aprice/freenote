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

function fade(el, ms, step) {
	if (!step) el.style.opacity = 1;
	if ((el.style.opacity -= .05) <= 0) {
		el.style.display = "none";
		el.style.opacity = 1;
	} else {
		setTimeout(function () { fade(el, ms, true) }, ms * .05);
	}
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
