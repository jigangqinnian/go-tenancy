package products

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/fatih/color"
	"github.com/kataras/iris/v12"
	"github.com/qor/render"
	"go-tenancy/models/products"
	"go-tenancy/utils"
)

// Controller products controller
type Controller struct {
	View *render.Render
}

// Index products index page
func (ctrl Controller) Index(ctx iris.Context) {
	var (
		Products []products.Product
		tx       = utils.GetDB(ctx.Request())
	)

	tx.Preload("Category").Find(&Products)

	if err := ctrl.View.Execute("index", map[string]interface{}{"Products": Products}, ctx.Request(), ctx.ResponseWriter()); err != nil {
		color.Red(fmt.Sprintf(" ctrl.View.Execute %v", err))
	}
}

// Gender products gender page
func (ctrl Controller) Gender(ctx iris.Context) {
	var (
		Products []products.Product
		tx       = utils.GetDB(ctx.Request())
	)

	tx.Where(&products.Product{Gender: ctx.Params().Get("gender")}).Preload("Category").Find(&Products)

	if err := ctrl.View.Execute("gender", map[string]interface{}{"Products": Products}, ctx.Request(), ctx.ResponseWriter()); err != nil {
		color.Red(fmt.Sprintf(" ctrl.View.Execute %v", err))
	}
}

// Show product show page
func (ctrl Controller) Show(ctx iris.Context) {
	var (
		product        products.Product
		colorVariation products.ColorVariation
		codes          = strings.Split(ctx.Params().Get("code"), "_")
		productCode    = codes[0]
		colorCode      string
		tx             = utils.GetDB(ctx.Request())
	)

	if len(codes) > 1 {
		colorCode = codes[1]
	}

	if tx.Preload("Category").Where(&products.Product{Code: productCode}).First(&product).RecordNotFound() {
		http.Redirect(ctx.ResponseWriter(), ctx.Request(), "/", http.StatusFound)
	}

	tx.Preload("Product").Preload("Color").Preload("SizeVariations.Size").Where(&products.ColorVariation{ProductID: product.ID, ColorCode: colorCode}).First(&colorVariation)

	ctrl.View.Execute("show", map[string]interface{}{"CurrentColorVariation": colorVariation}, ctx.Request(), ctx.ResponseWriter())
}

// Category category show page
func (ctrl Controller) Category(ctx iris.Context) {
	var (
		category products.Category
		Products []products.Product
		tx       = utils.GetDB(ctx.Request())
	)

	if tx.Where("code = ?", ctx.Params().Get("code")).First(&category).RecordNotFound() {
		http.Redirect(ctx.ResponseWriter(), ctx.Request(), "/", http.StatusFound)
	}

	tx.Where(&products.Product{CategoryID: category.ID}).Preload("ColorVariations").Find(&Products)

	ctrl.View.Execute("category", map[string]interface{}{"CategoryName": category.Name, "Products": Products}, ctx.Request(), ctx.ResponseWriter())
}
