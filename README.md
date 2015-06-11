# Holst

Holst is a (minimalist) *Recommentations Engine as a Service*. Using a RESTful API, it allows users to create and manage their own sets of data (comprised of *Users*, *Items*, and *Links*), and to retrieve customized recommendations, based on that information.

In a nutshell, a user can have one or multiple relations (links) with an item, which are subsequently used to compound the recommendations for a given user.

## Getting started

* Install and start [Neo4j](http://neo4j.com/).
* Install dependencies, and build the app: `go get & go build`.
* Copy the sample config file `config.json.sample` into a new file `config.json`, under the same directory, and edit this file according to your system configuration
* Run the app: `./holst`


## Authorization & Authentication
Current implementation includes temporary authentication mechanism. All requests must include an `Authorization` header, using [Basic access authentication](http://en.wikipedia.org/wiki/Basic_access_authentication). The username and password are directly set up in the `config.json` file.

## Manage data (CRUD operations)

The data model consists of *users*, *items*, and *links*. Every model contains a unique identifier, which can be explicitly provided, or automatically generated when issuing a `POST` request.


## Users

* id [required]
* name [optional]


### Create/Update users

* POST /user
* **Payload**: JSON-encoded user. Explicitly providing an ID will update the existing entity, if found. Otherwise a new user will be created. The lack of ID will always result in a new entity creation.

* Example Request:
		
	```json	
	POST /user
	{ 
	  "name" : "Roger"
	}	
	```
* Example Response
	
	```json
	201 Created
	{
      "id": "8a7ddc079f118bb1",
      "name": "Roger"
    }
	```
	
### Read users

* GET /user/{id}

* Example Request:
		
	```
	GET /user/8a7ddc079f118bb1	
	```
* Example Response
	
	```json
	200 OK
	{
      "id": "8a7ddc079f118bb1",
      "name": "Roger"
    }
	```
	
### Delete users

* DELETE /user/{id}

* Example Request:
		
	```	
	DELETE /user/8a7ddc079f118bb1	
	```
* Example Response
	
	```
	200 OK
	```

## Items

* id [required]
* name [optional]
* categories [optional]

### Create/Update items

* POST /item
* **Payload**: JSON-encoded item. Explicitly providing an ID will update the existing entity, if found. Otherwise a new item will be created. The lack of ID will always result in a new entity creation.

* Example Request:
		
	```json
	POST /item
	{ 
	  "name" : "Pulp Fiction",
	  "categories": ["movie", "1994"]
	}	
	```
* Example Response
	
	```json
	201 Created
	{
      "id": "8a7ddc079f118bb2",
      "name": "Pulp Fiction",
      "categories": ["movie", "1994"]
    }
	```

### Read items

* GET /item/{id}

* Example Request:
		
	```	
	GET /item/8a7ddc079f118bb2
	```
* Example Response
	
	```json
	200 OK
	{
      "id": "8a7ddc079f118bb2",
      "name": "Pulp Fiction",
      "categories": ["movie", "1994"]
    }
	```
	
### Delete items

* DELETE /item/{id}

* Example Request:
		
	```	
	DELETE /item/8a7ddc079f118bb2
	```
* Example Response
	
	```
	200 OK
	```
	
## Links

* id [required]
* userId [required]
* itemId [required]
* type [optional]
* score [optional]

### Create/Update links

Although a link's ID can also be automatically generated, both the userId and itemId need to be specified upfront. Additionally, you can pass the link *type* and *score*. These can be used for filtering results. A type can be any keyword relevant to your application (for example: **buy**, **like**, etc)

* Example Request:
		
	```json
	POST /link
	{
	  "userId" : "8a7ddc079f118bb1",
	  "itemId" : "8a7ddc079f118bb2",
	  "type" : "like"
	}	
	```
* Example Response
	
	```json
	201 Created
	{
      "id": "8a7ddc079f118bb3",
      "userId" : "8a7ddc079f118bb1",
	  "itemId" : "8a7ddc079f118bb2",
	  "type" : "like"
    }
	```

	
### Read links

* GET /link/{id}

* Example Request:
		
	```	
	GET /link/8a7ddc079f118bb3
	```
* Example Response
	
	```json
	200 OK
	{
      "id": "8a7ddc079f118bb3",
      "userId" : "8a7ddc079f118bb1",
      "itemId" : "8a7ddc079f118bb2",
      "type" : "like"
    }
	```
	
### Delete links

* DELETE /link/{id}

* Example Request:
		
	```	
	DELETE /item/8a7ddc079f118bb3
	```
* Example Response
	
	```
	200 OK
	```

# Recommendations

Recommendations are performed using collaborative filtering. Providing a user id, and optionally, item categories and link type, the system will return a list of recommended items.

### Getting recommendations

* GET /recommend/{userId}?[category=cat1&category=cat2]&[type=t]

Where *category* can hold multiple values (just concatenate the parameters in the query), and type is the link type. For example, querying only for `type=like` will only return results based on that relationship type. Each recommendation has a *strength* attribute, which represents the number of occurrences of such item. Results are ordered by this attribute.

* Example Request:
		
	```	
	GET /recommend/8a7ddc079f118bb1?category=movie&type=buy
	```
* Example Response
	
	```json
	200 OK
    [
       {
         "item": {
             "id": "aaaa",
             "name": "Casablanca"
         },
         "strength": 7
       },
       {
         "item": {
             "id": "bbbb",
             "name": "Star Wars"
         },
         "strength": 4
       },
       {
         "item": {
             "id": "cccc",
             "name": "The Lord of the Rings"
         },
         "strength": 3
       }
    ]
	```


## License

This software is distributed under the BSD-style license found in the LICENSE file.

