package account

import (
	"strconv"

	"github.com/kataras/iris/v12"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	qorutils "github.com/qor/qor/utils"
	"github.com/qor/render"
	"github.com/qor/validations"
	"go-tenancy/config/application"
	"go-tenancy/middleware"
	"go-tenancy/models/users"
	"go-tenancy/utils/funcmapmaker"
	"golang.org/x/crypto/bcrypt"
)

// New new account app
func New(config *Config) *App {
	return &App{Config: config}
}

// App account app
type App struct {
	Config *Config
}

// Config account config struct
type Config struct {
}

// ConfigureApplication configure application
func (app App) ConfigureApplication(application *application.Application) {
	controller := &Controller{View: render.New(&render.Config{AssetFileSystem: application.AssetFS.NameSpace("account")}, "app/account/views")}

	funcmapmaker.AddFuncMapMaker(controller.View)
	app.ConfigureAdmin(application.Admin)

	application.IrisApp.PartyFunc("/admin/account", func(account iris.Party) {
		account.Use(middleware.Authorize)
		account.Get("/", controller.Orders)
		account.Post("/add_user_credit", middleware.AuthorizeloggedInHalfHour, controller.AddCredit) // role: logged_in_half_hour

		account.Get("/profile", controller.Profile)
		account.Post("/profile", controller.Update)
	})
}

// ConfigureAdmin configure admin interface
func (App) ConfigureAdmin(Admin *admin.Admin) {
	Admin.AddMenu(&admin.Menu{Name: "用户管理", Priority: 3})
	user := Admin.AddResource(&users.User{}, &admin.Config{Name: "用户列表", Menu: []string{"用户管理"}, IconName: "pe-7s-users"})
	user.Meta(&admin.Meta{Name: "Gender", Config: &admin.SelectOneConfig{Collection: []string{"Male", "Female", "Unknown"}}})
	user.Meta(&admin.Meta{Name: "Birthday", Type: "date"})
	user.Meta(&admin.Meta{Name: "Role", Config: &admin.SelectOneConfig{Collection: []string{"Admin", "Maintainer", "Member"}}})
	user.Meta(&admin.Meta{Name: "Password",
		Type:   "password",
		Valuer: func(interface{}, *qor.Context) interface{} { return "" },
		Setter: func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			if newPassword := qorutils.ToString(metaValue.Value); newPassword != "" {
				bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
				if err != nil {
					_ = context.DB.AddError(validations.NewError(user, "Password", "Can't encrpt password"))
					return
				}
				u := resource.(*users.User)
				u.Password = string(bcryptPassword)
			}
		},
	})
	user.Meta(&admin.Meta{Name: "Confirmed", Valuer: func(user interface{}, ctx *qor.Context) interface{} {
		if user.(*users.User).ID == 0 {
			return true
		}
		return user.(*users.User).Confirmed
	}})
	user.Meta(&admin.Meta{Name: "DefaultBillingAddress", Config: &admin.SelectOneConfig{Collection: userAddressesCollection}})
	user.Meta(&admin.Meta{Name: "DefaultShippingAddress", Config: &admin.SelectOneConfig{Collection: userAddressesCollection}})

	user.Filter(&admin.Filter{
		Name: "Role",
		Config: &admin.SelectOneConfig{
			Collection: []string{"Admin", "Maintainer", "Member"},
		},
	})

	user.IndexAttrs("ID", "Email", "Name", "Gender", "Role", "Balance")
	user.ShowAttrs(
		&admin.Section{
			Title: "Basic Information",
			Rows: [][]string{
				{"Name"},
				{"Email", "Password"},
				{"Avatar"},
				{"Gender", "Role", "Birthday"},
				{"Confirmed"},
			},
		},
		&admin.Section{
			Title: "Credit Information",
			Rows: [][]string{
				{"Balance"},
			},
		},
		&admin.Section{
			Title: "Accepts",
			Rows: [][]string{
				{"AcceptPrivate", "AcceptLicense", "AcceptNews"},
			},
		},
		&admin.Section{
			Title: "Default Addresses",
			Rows: [][]string{
				{"DefaultBillingAddress"},
				{"DefaultShippingAddress"},
			},
		},
		"Addresses",
	)
	user.EditAttrs(user.ShowAttrs())
}

func userAddressesCollection(resource interface{}, context *qor.Context) (results [][]string) {
	var (
		user users.User
		DB   = context.DB
	)

	DB.Preload("Addresses").Where(context.ResourceID).First(&user)

	for _, address := range user.Addresses {
		results = append(results, []string{strconv.Itoa(int(address.ID)), address.Stringify()})
	}
	return
}
