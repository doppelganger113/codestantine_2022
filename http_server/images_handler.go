package http_server

import (
	"api/auth"
	"api/image"
	"api/storage"
	"context"
	"mime/multipart"
)

type ImagesHandler interface {
	Get(ctx context.Context, limit, offset int, order storage.Order) (storage.ImageList, error)
	GetOne(ctx context.Context, imageId string) (storage.Image, error)
	UploadAndResize(
		ctx context.Context,
		authorization auth.AuthorizationDto,
		imageName string,
		format image.Format,
		originalFile *multipart.FileHeader,
		croppedFile *multipart.FileHeader,
	) (storage.Image, error)
	DeleteOne(
		ctx context.Context,
		auth auth.AuthorizationDto,
		imageId string,
	) error
}
