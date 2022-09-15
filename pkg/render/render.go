package render

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/bopepsi/bookings/pkg/config"
	"github.com/bopepsi/bookings/pkg/models"
)

// serve html files
func RenderTemplateBasic(w http.ResponseWriter, tmpl string) {
	parsed, _ := template.ParseFiles("templates/"+tmpl, "templates/base.layout.html")
	err := parsed.Execute(w, nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
}

// Use a cache to store templates used on server
var cache map[string]*template.Template = map[string]*template.Template{}

func RenderTemplateBasicTwo(w http.ResponseWriter, t string) {

	var tmpl *template.Template
	var err error

	_, ok := cache[t]

	if !ok {
		fmt.Println("Adding to cache")
		err = createAndAnddToCache(t)
	} else {
		fmt.Println("Reading from cache")
	}

	tmpl = cache[t]

	err = tmpl.Execute(w, nil)

	if err != nil {
		fmt.Println(err)
	}

}

func createAndAnddToCache(t string) error {

	templateFiles := []string{fmt.Sprintf("templates/%v", t), "templates/base.layout.html"}

	//ParseFiles takes arr of paths
	tmpl, err := template.ParseFiles(templateFiles...)

	if err != nil {
		return err
	}

	cache[t] = tmpl
	return nil
}

// Use cache generated in main
var appConfig config.AppConfig

func SetupTmplCacheMap(a *config.AppConfig) {
	appConfig = *a
}

// Best way to implement render
func RenderTemplate(w http.ResponseWriter, page string, td *models.TemplateData) {
	// create a tmplate cache or read from app wide config

	var cache map[string]*template.Template

	if appConfig.UseCache {
		cache = appConfig.TemplateCache
	}

	if !appConfig.UseCache {
		cache, _ = CreateTemplateCache()
	}

	// get requested template from cache
	tmpl, ok := cache[page]
	if !ok {
		okErr := errors.New("no views found")
		log.Fatal(okErr)
	}

	// render template
	tmpl.Execute(w, *td)
}

func CreateTemplateCache() (map[string]*template.Template, error) {
	var templatesCache map[string]*template.Template = map[string]*template.Template{}

	pages, err := filepath.Glob("templates/*.page.html")
	if err != nil {
		return nil, err
	}

	// range through all files ending with *.page.html
	for _, page := range pages {
		name := filepath.Base(page)
		parsedTmplate, err := template.ParseFiles(page)

		if err != nil {
			return nil, err
		}

		layoutPages, err := filepath.Glob("templates/*.layout.html")
		if err != nil {
			return nil, err
		}
		if len(layoutPages) > 0 {
			_, err := parsedTmplate.ParseGlob("templates/*.layout.html")
			if err != nil {
				return nil, err
			}
		}
		fmt.Println(parsedTmplate)
		templatesCache[name] = parsedTmplate

	}

	return templatesCache, nil

}
