package main

import (
	"github.com/iteais/sdk/example/controllers"
	_ "github.com/iteais/sdk/example/docs"
	"github.com/iteais/sdk/example/migrations"
	_ "github.com/iteais/sdk/example/migrations"
	"github.com/iteais/sdk/pkg"
)

// @title           API микросервиса
// @version         0.0.1
// @description     API микросервиса.
// @termsOfService  https://some-host.id

// @contact.name   API Support
// @contact.email  fedor@support-pc.org

// @license.name  Apache 2.0
// @license.url   https://www.apache.org/licenses/LICENSE-2.0.html

// @host      api.some-host.id
// @BasePath  /
func main() {

	app := pkg.NewApplication("api", migrations.Migrations)
	app.AppendGetEndpoint("/user/:id", controllers.GetById()).
		AppendGetEndpoint("/user/proxy", controllers.Proxy()).
		AppendSwagger("").
		Run()
}
