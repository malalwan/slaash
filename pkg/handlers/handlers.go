package handlers

import (
	"net/http"

	"github.com/malalwan/slaash/pkg/config"
	"github.com/malalwan/slaash/pkg/models"
	"github.com/malalwan/slaash/pkg/render"
)

// repo used by handlers
var Repo *Repository

// repository type
type Repository struct {
	App *config.AppConfig
}

// Creates a new repo
func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}

// Sets the repo for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home Page Calls
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	remoteIP := r.RemoteAddr
	m.App.Session.Put(r.Context(), "remote_ip", remoteIP)

	render.RenderTemplate(w, "home.page.tmpl", &models.TemplateData{}) //to pass empty template data
}

// About Page Calls
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["test"] = "Hello, again!"

	remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")

	stringMap["remote_ip"] = remoteIP

	render.RenderTemplate(w, "about.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}
