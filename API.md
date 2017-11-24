# Freenote API

Routes:

```
/ - HTML GUI
	session/ - session API (GET: current session info, POST: log in, DELETE: log out current  session)
	users/ - users API (GET: list, POST: register)
		{username} - (GET: view)
		{id}/ - (GET: view, PUT: replace, PATCH: modify, DELETE: delete)
			password - (PUT: update)
			notes/ - (GET: list, POST: create)
				{path} (GET: view)
				{id} (HEAD: metadata, GET: view, PUT: replace, PATCH: modify, DELETE: delete)
	debug/ - only available with dev tag
		pprof/
		expvar/
```

All routes support OPTIONS, which will return, at a minimum, appropriate Accept headers.

IDs are all UUIDs and can be provided in standard UUID format or as Base64 bytes.

Most routes return results including a collection of hypermedia links. In JSON, this is
under the `_links` key in the root of the result. In XML, it is under the `Links` element
under the root. Individual items return links for `edit`, `save`, and `delete`; collections
return links for `next`, `previous`, and `create`. Some items or collections may return
additional links by type.

Collections (`/users/` and `/users/{id}/notes/`) support sorting and pagination using query string
parameters as follows:

- `start=n` where `n` is an integer start index in the record set
- `length=n` where `n` is the number of records to return per page
- `sort=s` where `s` is the name of a sortable field
- `order=s` where `s` is either `asc` or `desc`

Notes collection (`/users/{id}/notes`) takes additional filter parameters:

- `folder=path` where `path` is a note folder path; only notes in this folder
will be returned
- `modifiedSince=date` where `date` is an RFC3339 date; only notes modified more
recently than this date (exclusive) will be returned

Authentication:

- HTTP Basic
- Cookie

Supported request content types:

- application/json
- application/xml
- text/plain (POST/PUT notes only - _NIY_)
- text/markdown (POST/PUT notes only - _NIY_)

Supported response content types:

- application/json
- application/javascript (JSONP; cb query string specifies function to call with result)
- application/xml
- text/html (simple plain semantic HTML document, no JS/CSS - _NIY_)
- text/markdown (GET single note only - _NIY_)
- text/plain (GET single note only - _NIY_)
