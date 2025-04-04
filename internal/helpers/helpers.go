package helpers

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/malalwan/slaash/internal/config"
	"github.com/malalwan/slaash/internal/models"
)

var app *config.AppConfig

// NewHelpers sets up app config for helpers
func NewHelpers(a *config.AppConfig) {
	app = a
}

func ClientError(w http.ResponseWriter, status int) {
	app.InfoLog.Println("Client error with status of", status)
	http.Error(w, http.StatusText(status), status)
}

func ServerError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.ErrorLog.Println(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func IsAuthenticated(r *http.Request) bool {
	exists := app.Session.Exists(r.Context(), "user")
	// added via : session.Put(r.Context(), "user", user) //user of type User struct that we registered
	return exists
}

func GetOtf(models.VisitTable) (bool, error) {
	return true, nil
}
