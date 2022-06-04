package http_server

import (
	"api/auth"
	"api/core/exception"
	"api/http_server/authenticator"
	"api/http_server/http_util"
	"api/http_server/middleware"
	"api/image"
	"api/storage"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"net/http"
)

const maxBodyLimitBytes = 30 * 1024 * 1024 // 20MB

func ImagesRouter(handler ImagesHandler, logger *zerolog.Logger, authenticator authenticator.Authenticator) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", FetchImages(handler, logger))
		r.Get("/{imageId}", FetchImage(handler, logger))
		r.Post("/",
			middleware.Authorize(AddImage(handler, logger), authenticator, auth.RoleAdmin),
		)
		r.Patch("/{imageId}",
			middleware.Authorize(UpdateImage(handler, logger), authenticator, auth.RoleAdmin),
		)
		r.Delete("/{imageId}",
			middleware.Authorize(DeleteOne(handler, logger), authenticator, auth.RoleAdmin),
		)
	}
}

func FetchImage(handler ImagesHandler, logger *zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		imageId := chi.URLParam(r, "imageId")
		img, err := handler.GetOne(ctx, imageId)
		if err != nil {
			http_util.HandleError(logger, w, err)
			return
		}

		http_util.WriteJson(w, http.StatusOK, img)
	}
}

func FetchImages(handler ImagesHandler, logger *zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		page := http_util.ToUint(r.URL.Query().Get("page"))
		size := http_util.ToUint(r.URL.Query().Get("size"))

		order := storage.ToOrderOr(r.URL.Query().Get("order"), storage.OrderDescending)
		limit, offset := storage.PagingToLimitOffset(page, size)

		imageList, err := handler.Get(ctx, limit, offset, order)
		if err != nil {
			http_util.HandleError(logger, w, err)
			return
		}

		http_util.WriteJson(w, http.StatusOK, imageList)
	}
}

type UploadImageDto struct {
	Name   string
	Format image.Format
}

func (dto UploadImageDto) validate() error {
	if len(dto.Name) < 5 || len(dto.Name) > 200 {
		return exception.InvalidArgument{
			Reason: "Name should be between 5 and 250 characters",
		}
	}

	if !dto.Format.IsSupported() {
		return exception.InvalidArgument{
			Reason: fmt.Sprintf("Unsupported format %s", dto.Format),
		}
	}

	return nil
}

func AddImage(handler ImagesHandler, logger *zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(maxBodyLimitBytes)
		if err != nil {
			http_util.WriteJson(
				w,
				http.StatusBadRequest,
				http_util.NewFailureResponse("failed parsing multipart form data"),
			)
			return
		}

		_, originalFileHeader, err := r.FormFile("originalFile")
		if err != nil {
			http_util.WriteJson(
				w,
				http.StatusBadRequest,
				http_util.NewFailureResponse("missing originalFile"),
			)
			return
		}
		_, croppedFileHeader, err := r.FormFile("croppedFile")
		if err != nil {
			http_util.WriteJson(
				w,
				http.StatusBadRequest,
				http_util.NewFailureResponse("missing croppedFile"),
			)
			return
		}

		data := &UploadImageDto{}
		data.Name = r.PostFormValue("name")
		data.Format = image.Format(r.PostFormValue("format"))
		if err = data.validate(); err != nil {
			http_util.WriteBadRequestJson(w, err)
			return
		}

		ctx := r.Context()
		authorization, err := auth.ExtractAuthorizationDto(ctx, string(middleware.UserAuthDtoKey))
		if err != nil {
			http_util.HandleError(logger, w, err)
			return
		}

		img, err := handler.UploadAndResize(
			ctx,
			authorization,
			data.Name,
			data.Format,
			originalFileHeader,
			croppedFileHeader,
		)
		if err != nil {
			http_util.HandleError(logger, w, err)
			return
		}

		http_util.WriteJson(w, http.StatusCreated, img)
	}
}

func UpdateImage(handler ImagesHandler, logger *zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(maxBodyLimitBytes)
		if err != nil {
			http_util.WriteJson(
				w,
				http.StatusBadRequest,
				http_util.NewFailureResponse("failed parsing multipart form data"),
			)
			return
		}

		_, originalFileHeader, err := r.FormFile("originalFile")
		if err != nil {
			http_util.WriteJson(
				w,
				http.StatusBadRequest,
				http_util.NewFailureResponse("missing originalFile"),
			)
			return
		}
		_, croppedFileHeader, err := r.FormFile("croppedFile")
		if err != nil {
			http_util.WriteJson(
				w,
				http.StatusBadRequest,
				http_util.NewFailureResponse("missing croppedFile"),
			)
			return
		}

		data := &UploadImageDto{}
		data.Name = r.PostFormValue("name")
		data.Format = image.Format(r.PostFormValue("format"))
		if err = data.validate(); err != nil {
			http_util.WriteBadRequestJson(w, err)
			return
		}

		ctx := r.Context()
		authorization, err := auth.ExtractAuthorizationDto(ctx, string(middleware.UserAuthDtoKey))
		if err != nil {
			http_util.HandleError(logger, w, err)
			return
		}

		img, err := handler.UploadAndResize(
			ctx,
			authorization,
			data.Name,
			data.Format,
			originalFileHeader,
			croppedFileHeader,
		)
		if err != nil {
			http_util.HandleError(logger, w, err)
			return
		}

		http_util.WriteJson(w, http.StatusCreated, img)
	}
}

func DeleteOne(handler ImagesHandler, logger *zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		imageId := chi.URLParam(r, "imageId")
		authDto, err := auth.ExtractAuthorizationDto(ctx, string(middleware.UserAuthDtoKey))
		if err != nil {
			http_util.HandleError(logger, w, err)
			return
		}
		err = handler.DeleteOne(ctx, authDto, imageId)
		if err != nil {
			http_util.HandleError(logger, w, err)
			return
		}

		http_util.WriteJson(w, http.StatusNoContent, nil)
	}
}
