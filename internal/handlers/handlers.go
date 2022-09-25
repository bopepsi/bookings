package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/bopepsi/bookings/internal/config"
	"github.com/bopepsi/bookings/internal/driver"
	"github.com/bopepsi/bookings/internal/forms"
	"github.com/bopepsi/bookings/internal/helpers"
	"github.com/bopepsi/bookings/internal/models"
	"github.com/bopepsi/bookings/internal/render"
	"github.com/bopepsi/bookings/internal/repository"
	"github.com/bopepsi/bookings/internal/repository/dbrepo"
	"github.com/go-chi/chi"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home is the handler for the home page
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	// remoteIP := r.RemoteAddr
	// m.App.Session.Put(r.Context(), "remote_ip", remoteIP)

	render.Template(w, r, "home.page.html", &models.TemplateData{})
}

// About is the handler for the about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	// perform some logic
	// stringMap := make(map[string]string)
	// stringMap["test"] = "Hello, again"

	// remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")
	// stringMap["remote_ip"] = remoteIP

	// send data to the template
	render.Template(w, r, "about.page.html", &models.TemplateData{
		// StringMap: stringMap,
	})
}

// Reservation renders the make a reservation page and displays form
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {

	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, errors.New("Cannot get reservation from session"))
		return
	}

	room, err := m.DB.GetRoomByID(res.RoomID)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	res.Room.RoomName = room.RoomName

	m.App.Session.Put(r.Context(), "reservation", res)

	data := make(map[string]interface{})
	data["reservation"] = res

	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	render.Template(w, r, "make-reservation.page.html", &models.TemplateData{
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
}

// PostReservation handles the posting of a reservation form
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	// err = errors.New("this is an error msg")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// sd := r.Form.Get("start_date")
	// ed := r.Form.Get("end_date")

	// in format
	// layout := "2006-01-02"
	// startDate, err := time.Parse(layout, sd)
	// if err != nil {
	// 	helpers.ServerError(w, err)
	// 	return
	// }
	// endDate, err := time.Parse(layout, ed)
	// if err != nil {
	// 	helpers.ServerError(w, err)
	// 	return
	// }

	// fmt.Println(r.Form.Get("room_id"))

	// roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	// if err != nil {
	// 	helpers.ServerError(w, err)
	// 	return
	// }

	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)

	if !ok {
		helpers.ServerError(w, errors.New("cannot get from session"))
		return
	}

	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")

	m.App.Session.Put(r.Context(), "reservation", res)

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3, r)
	// form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = res
		render.Template(w, r, "make-reservation.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	var newReservationID int
	newReservationID, err = m.DB.InsertReservation(res)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	restriction := models.RoomRestriction{
		StartDate:     res.StartDate,
		EndDate:       res.EndDate,
		RoomID:        res.RoomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	htmlMsgCustomer := fmt.Sprintf(`
	<strong>Confirmation</strong>
	Dear %s, <br/>
	This is confirm your reservation from %s to %s`, res.FirstName,
		res.StartDate.Format("2006-01-02"), res.EndDate.Format("2006-01-02"))

	// send email notifications
	msgToCustomer := models.MailData{
		To:       res.Email,
		From:     "bo@gmail.com",
		Subject:  "Reservation Confirmation",
		Content:  htmlMsgCustomer,
		Template: "basic.html",
	}

	htmlMsgOwner := fmt.Sprintf(`
	<strong>Confirmation</strong>
	Hi there! <br/>
	A reservation just made from %s to %s by %s on room %s.`,
		res.StartDate.Format("2006-01-02"), res.EndDate.Format("2006-01-02"), res.FirstName, res.Room.RoomName)

	msgToOwner := models.MailData{
		To:       "owner@gmail.com",
		From:     "bo@gmail.com",
		Subject:  "Reservation Notice",
		Content:  htmlMsgOwner,
		Template: "basic.html",
	}

	m.App.MailChan <- msgToCustomer

	m.App.MailChan <- msgToOwner

	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// Generals renders the room page
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "generals.page.html", &models.TemplateData{})
}

// Majors renders the room page
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "majors.page.html", &models.TemplateData{})
}

// Availability renders the search availability page
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "search-availability.page.html", &models.TemplateData{})
}

// PostAvailability handles post
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, start)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	endData, err := time.Parse(layout, end)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endData)
	fmt.Println(rooms)
	if len(rooms) == 0 {
		fmt.Println("No rooms available")
		m.App.Session.Put(r.Context(), "error", "No availability")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	reservation := models.Reservation{
		StartDate: startDate,
		EndDate:   endData,
	}

	m.App.Session.Put(r.Context(), "reservation", reservation)

	render.Template(w, r, "choose-room.page.html", &models.TemplateData{
		Data: data,
	})
	// w.Write([]byte(fmt.Sprintf("start date is %s and end is %s", start, end)))
}

type jsonResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomId    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// AvailabilityJSON handles request for availability and sends JSON response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {

	sd := r.Form.Get("start")
	ed := r.Form.Get("end")

	roomId, _ := strconv.Atoi(r.Form.Get("room_id"))

	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	available, _ := m.DB.SearchAvailabilityByDatesByRoomID(startDate, endDate, roomId)

	resp := jsonResponse{
		OK:        available,
		StartDate: sd,
		EndDate:   ed,
		RoomId:    strconv.Itoa(roomId),
	}

	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// Contact renders the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.html", &models.TemplateData{})
}

// ReservationSummary displays the res summary page
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.ErrorLog.Println("can't get item from session")
		// log.Println("can't get item from session")
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation

	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.EndDate.Format("2006-01-02")

	stringMap := map[string]string{"start_date": sd, "end_date": ed}

	render.Template(w, r, "reservation-summary.page.html", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
}

func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, err)
		return
	}

	res.RoomID = roomID
	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	//id,s,e
	queryID := r.URL.Query().Get("id")
	ID, _ := strconv.Atoi(queryID)

	startDate := r.URL.Query().Get("s")
	endDate := r.URL.Query().Get("e")

	layout := "2006-01-02"
	sd, _ := time.Parse(layout, startDate)
	ed, _ := time.Parse(layout, endDate)

	res := models.Reservation{
		RoomID:    ID,
		StartDate: sd,
		EndDate:   ed,
	}

	m.App.Session.Put(r.Context(), "reservation", res)
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)

}

func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {

	render.Template(w, r, "login.page.html", &models.TemplateData{
		Form: forms.New(nil),
	})
}

func (m *Repository) PostShowLogin(w http.ResponseWriter, r *http.Request) {

	// renew token when login or logout
	_ = m.App.Session.RenewToken(r.Context())
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	form := forms.New(r.PostForm)
	form.Required("email", "password")

	if !form.Valid() {
		//* take user back
		render.Template(w, r, "login.page.html", &models.TemplateData{
			Form: form,
		})
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "Invalid login credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
	}

	// store id in session
	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "logged in")
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-dashboard.page.html", &models.TemplateData{})
}
