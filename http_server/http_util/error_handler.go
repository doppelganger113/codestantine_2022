package http_util

import (
	"api/core/exception"
	"errors"
	"github.com/rs/zerolog"
	"net/http"
)

type FailureResponse struct {
	Err string `json:"err"`
}

func NewFailureResponse(msg string) *FailureResponse {
	return &FailureResponse{Err: msg}
}

var serverErrorFailure = NewFailureResponse("Server error")
var forbiddenFailure = NewFailureResponse("Forbidden")
var notFoundFailure = NewFailureResponse("Not found")

func HandleError(logger *zerolog.Logger, w http.ResponseWriter, err error) {

	var forbiddenFail exception.Forbidden
	if errors.As(err, &forbiddenFail) {
		WriteJson(w, http.StatusForbidden, forbiddenFailure)
		return
	}

	var notFoundFail exception.NotFound
	if errors.As(err, &notFoundFail) {
		WriteJson(w, http.StatusNotFound, notFoundFailure)
		return
	}

	var invalidArgumentFail exception.InvalidArgument
	if errors.As(err, &invalidArgumentFail) {
		WriteJson(w, http.StatusBadRequest, &FailureResponse{
			Err: invalidArgumentFail.Error(),
		})
		return
	}

	logger.Err(err).Msg("Unhandled error")

	WriteJson(w, http.StatusInternalServerError, serverErrorFailure)
}
