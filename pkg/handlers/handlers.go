package handlers

import (
	"html/template"
	"net/http"

	"github.com/bopepsi/bookings/pkg/config"
	"github.com/bopepsi/bookings/pkg/models"
	"github.com/bopepsi/bookings/pkg/render"
)

var Repo *Repository

type Repository struct {
	App *config.AppConfig
}

func SetupRepo(a *config.AppConfig) {
	Repo = &Repository{
		App: a,
	}
}

func (this *Repository) Index(w http.ResponseWriter, r *http.Request) {
	parsed, _ := template.ParseFiles("templates/index.page.html")
	parsed.Execute(w, nil)
}

func (this *Repository) Home(w http.ResponseWriter, r *http.Request) {

	remoteIp := r.RemoteAddr
	this.App.Session.Put(r.Context(), "remote_ip", remoteIp)

	
	render.Template(w, "home.page.html", &models.TemplateData{})
}

func (this *Repository) About(w http.ResponseWriter, r *http.Request) {
	// parsed, _ := template.ParseFiles("templates/about.page.html", "templates/base.layout.html")
	// err := parsed.Execute(w, nil)
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// }

	strMap := map[string]string{
		"test": "Hello from about page",
	}

	remoteIp := this.App.Session.GetString(r.Context(), "remote_ip")

	strMap["remote_ip"] = remoteIp

	render.Template(w, "about.page.html", &models.TemplateData{
		StringMap: strMap,
		CSRFToken: "THIS IS CSRF",
	})
}

func (this *Repository) Contact(w http.ResponseWriter, r *http.Request){
	render.Template(w, "contact.page.html", &models.TemplateData{})
}

func (this *Repository) Generals(w http.ResponseWriter, r *http.Request){
	render.Template(w, "generals.page.html", &models.TemplateData{})
}

func (this *Repository) Majors(w http.ResponseWriter, r *http.Request){
	render.Template(w, "majors.page.html", &models.TemplateData{})
}

func (this *Repository) Reservation(w http.ResponseWriter, r *http.Request){
	render.Template(w, "make-reservation.page.html", &models.TemplateData{})
}

func (this *Repository) Availability(w http.ResponseWriter, r *http.Request){
	render.Template(w, "search-availability.page.html", &models.TemplateData{})
}

func (this *Repository) PostAvailability(w http.ResponseWriter, r *http.Request){
	w.Write([]byte("this is from post req"))
}
