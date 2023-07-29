package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"golang.binggl.net/monorepo/internal/bookmarks/app"
	pkgerr "golang.binggl.net/monorepo/pkg/errors"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"
)

func encodeError(err error, logger logging.Logger, w http.ResponseWriter, r *http.Request) {
	if err == nil {
		panic("encodeError with nil error")
	}

	logger.ErrorRequest(fmt.Sprintf("got error during request: %v", err), r)

	t := "about:blank"
	var (
		pd            *pkgerr.ProblemDetail
		errNotFound   *app.NotFoundError
		errValidation *app.ValidationError
	)
	if errors.As(err, &errNotFound) {
		pd = &pkgerr.ProblemDetail{
			Type:   t,
			Title:  "object cannot be found",
			Status: http.StatusNotFound,
			Detail: errNotFound.Error(),
		}
	} else if errors.As(err, &errValidation) {
		pd = &pkgerr.ProblemDetail{
			Type:   t,
			Title:  "error in parameter-validaton",
			Status: http.StatusBadRequest,
			Detail: errValidation.Error(),
		}
	} else {
		pd = &pkgerr.ProblemDetail{
			Type:   t,
			Title:  "cannot service the request",
			Status: http.StatusInternalServerError,
			Detail: err.Error(),
		}
	}
	writeProblemJSON(logger, w, pd)
}

func respondJSON(w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func writeProblemJSON(logger logging.Logger, w http.ResponseWriter, pd *pkgerr.ProblemDetail) {
	w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	w.WriteHeader(pd.Status)
	b, err := json.Marshal(pd)
	if err != nil {
		logger.Error(fmt.Sprintf("could not marshal json %v", err))
	}
	_, err = w.Write(b)
	if err != nil {
		logger.Error(fmt.Sprintf("could not write bytes using http.ResponseWriter: %v", err))
	}
}

func queryParam(r *http.Request, name string) string {
	keys, ok := r.URL.Query()[name]
	if !ok || len(keys[0]) < 1 {
		return ""
	}
	return keys[0]
}

func ensureUser(r *http.Request) *security.User {
	user, ok := security.UserFromContext(r.Context())
	if !ok || user == nil {
		panic("the sucurity context user is not available!")
	}
	return user
}
