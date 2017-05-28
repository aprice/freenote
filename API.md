# Freenote API

Routes:
```
/ - homepage
	ui/ - HTML5 UI
	static/
		css/
		js/
		img/
	session/ - session API (GET: current session info, POST: log in, DELETE: log out current  session)
		{userid}/ - (GET: list sessions, POST: log in, DELETE: log out all sessions)
			{sessionid} - (GET: session info, DELETE: log out)
	users/ - users API (GET: list, POST: register)
		{username} - (GET: view)
		{id}/ - (GET: view, PUT: replace, PATCH: modify, DELETE: delete)
			password - (PUT: update)
			notes/ - (GET: list, POST: create)
				{path} (GET: view)
				{id} (HEAD: metadata, GET: view, PUT: replace, PATCH: modify, DELETE: delete)
```

All routes support OPTIONS, which will return, at a minimum, appropriate Accept headers.

Collections (/users/ and /users/{id}/notes/) support sorting and pagination using query string
parameters as follows:
 - start=n where n is an integer start index in the record set
 - length=n where n is the number of records to return per page
 - sort=s where s is the name of a sortable field
 - order=s where s is either asc or desc

Authentication:
 - HTTP Basic
 - HTTP Bearer (OAuth2)
 - Cookie (OAuth2)

API Request content types:
 - application/json
 - application/xml
 - text/plain (POST/PUT notes only)
 - text/markdown (POST/PUT notes only)

API Response content types:
 - application/json
 - application/javascript (JSONP; cb query string specifies function to call with result)
 - application/xml
 - text/html (simple plain semantic HTML document, no JS/CSS)
 - text/markdown (GET single note only)
 - text/plain (GET single note only)
