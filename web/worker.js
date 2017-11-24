'use strict';

var CACHE_NAME = 'freenote.1';
var urlsToCache = [
	'/',
	'index.html',
	'theme.css',
	'small-screen.css',
	'font-awesome.min.css',
	'fonts/fontawesome-webfont.woff2',
	'app.js',
	'auth.js',
	'db.js',
	'notes.js',
	'utils.js',
	'version.js'
];

self.addEventListener('install', function (event) {
	// Perform install steps
	event.waitUntil(
		caches.open(CACHE_NAME)
			.then(function (cache) {
				return cache.addAll(urlsToCache);
			})
	);
});

self.addEventListener('activate', function (event) {
});

self.addEventListener('fetch', function (event) {
	event.respondWith(
		caches.match(event.request).then(function (response) {
			return response || fetch(event.request);
		})
	);
});
/*
self.addEventListener('sync', function (event) {
	if (event.id == 'update-notes') {
		event.waitUntil(
			caches.open(CACHE_NAME).then(function (cache) {
				return cache.add('/users/'+getUserId+'/notes');
			})
		);
	}
});
*/