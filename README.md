revel-generator
===============

**revel-generator** is a simple commandline tool to help you generating code for
building custom, standalone [Revel](http://github.com/revel/revel) applications.

It consists mainly of methods extraced from Revel's CLI and brings no new features
to the table. If you're happy with the way revel work, you probably don't need
**revel-generator**.

## Usage ##

Add a call to **revel-generator** into you build toolchain.
It should look somewhat like this:

`revel-generator [OPTIONS] APPLICATION_IMPORT_PATH`

With possible options being:

  * `-m RUN_MODE` for setting the default RunMode of the application. Also, this will
    generate the code in the context of the configured RunMode. By default this will
    be an empty string.

  * `-r ROUTES_FILE` for specifying the routes file to generate. The default is
    `app/routes/routes.go`.

  * `-t TARGET_FILE` for specifying the target file to generate. The default is
    `app/setup.go`. This file will contain the `Initialize()`, `InitializeWithRunMode()`,
    and `SetupRevel()` functions.

To your server application add code like this:

```
package main

import (
	"APPLICATION_IMPORT_PATH/app"
	"github.com/revel/revel"
)

func main() {
	app.Initialize()
	revel.Run(revel.HttpPort)
}
```

This will initialize Revel, register all controllers, etc., and serve the application for you.