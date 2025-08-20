//go:build !exclude_from_test

package main

import (
	"github.com/iteais/sdk/example/controllers"
	_ "github.com/iteais/sdk/example/docs"
	"github.com/iteais/sdk/example/models"
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

	config := pkg.ApplicationConfig{
		MigrationPath: "example/scripts/migrations",
		DbSchemaName:  "users",
		WhiteList:     []string{"/swagger/index.html"},
	}

	app := pkg.NewApplication(config)

	app.Router.Use(pkg.HmacMiddleware("http://localhost:8801", config.WhiteList...))

	app.AppendGetEndpoint("/user/:id", controllers.GetById()).
		AppendGetEndpoint("/user/proxy", controllers.Proxy()).
		AppendGetEndpoint("/user/list", pkg.ListAction[models.User]()).
		AppendGetEndpoint("/user/update", pkg.UpdateAction[models.User]("u.id")).
		AppendSwagger("").
		Run()
}
