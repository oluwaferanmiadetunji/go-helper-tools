# go-helper-tools
A collection of helper functions I've had to use in different projects

[![Version](https://img.shields.io/badge/goversion-1.19.x-blue.svg)](https://golang.org)
<a href="https://golang.org"><img src="https://img.shields.io/badge/powered_by-Go-3362c2.svg?style=flat-square" alt="Built with GoLang"></a>



# Helper Tools


The included tools are:

- Read JSON
- Write JSON
- Get a random string of length n
- Create a URL safe slug from a string

## Installation

`go get -u github.com/oluwaferanmiadetunji/go-helper-tools`


## Usage

```go
package main

import (
	"fmt"
	"github.com/oluwaferanmiadetunji/go-helper-tools"
)

func main() {
	// create a variable of type gohelpertools.Tools, so we can use this variable
	// to call the methods on that type
	var tools gohelpertools.Tools

	// get a random string
	randomString := tools.RandomString(10)
	fmt.Println(randomString)
}
```

### Working with JSON

In a handler, for example:

```go
// JSONPayload is the type for JSON data that we receive
type JSONPayload struct {
    Name string `json:"name"`
    Data string `json:"data"`
}

// SomeHandler is the handler to accept a post request consisting of json payload
func (app *Config) SomeHandler(w http.ResponseWriter, r *http.Request) {
    var tools gohelpertools.Tools
    
    // read json into var
    var requestPayload JSONPayload
    _ = tools.ReadJSON(w, r, &requestPayload)
	
    // do something with the data here...
    
    // create the response we'll send back as JSON
    resp := gohelpertools.JSONResponse{
        Error:   false,
        Message: "logged",
    }
    
    // write the response back as JSON
    _ = tools.WriteJSON(w, http.StatusAccepted, resp)
}
```

### Create a slug from a string

To slugify a string, we simply remove all non URL safe characters and return the
original string with a hyphen where spaces would be. Example:

```go
package main

import (
	"fmt"
	"github.com/oluwaferanmiadetunji/go-helper-tools"
)

func main() {
	toSlugify := "hello, world! These are unsafe chars: こんにちは世界*!&^%"
	fmt.Println("To slugify:", toSlugify)
	var tools gohelpertools.Tools

	slug, err := tools.Slugify(toSlugify)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Slugified:", slug)
}
```

Output from this is:

```
To slugify: hello, world! These are unsafe chars: こんにちは世界*!&^%
Slugified: hello-world-these-are-unsafe-chars
```