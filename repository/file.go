package repository

import (
	"context"
	"dishes-service-backend/entity"
	"fmt"
	"net/http"

	"github.com/Falokut/go-kit/http/client"
	"github.com/pkg/errors"
)

type File struct {
	cli     *client.Client
	baseUrl string
}

func NewFile(cli *client.Client, baseUrl string) File {
	return File{
		cli:     cli,
		baseUrl: baseUrl,
	}
}

func (r File) GetFileUrl(category string, filename string) string {
	return fmt.Sprintf("%s/%s/%s", r.baseUrl, category, filename)
}

func (r File) UploadFile(ctx context.Context, req entity.UploadFileRequest) error {
	endpoint := fmt.Sprintf("/%s/%s", req.Category, req.Filename)
	_, err := r.cli.Post(endpoint).
		RequestBody(req.Content).
		StatusCodeToError().
		Do(ctx)
	if err != nil {
		return errors.WithMessagef(err, "call storage service %s", endpoint)
	}
	return nil
}

func (r File) DeleteFile(ctx context.Context, category string, fileName string) error {
	endpoint := fmt.Sprintf("/%s/%s", category, fileName)
	resp, err := r.cli.Delete(endpoint).Do(ctx)
	if err != nil {
		return errors.WithMessagef(err, "call storage service %s", endpoint)
	}

	switch {
	case resp.StatusCode() == http.StatusNotFound:
		return nil
	case !resp.IsSuccess():
		return errors.Errorf("call storage service %s, unexpected status: %d", endpoint, resp.StatusCode())
	}
	return nil
}
