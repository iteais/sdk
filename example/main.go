package main

import (
	"github.com/xid/sdk/example/controllers"
	_ "github.com/xid/sdk/example/migrations"
	"github.com/xid/sdk/pkg"
)

func main() {

	app := pkg.NewApplication()
	app.AppendGetEndpoint("/user/:id", controllers.GetById())
	app.Run()

}
