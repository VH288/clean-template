package interfaces

import (
	"context"

	"clean-template/internal/models"

	"github.com/gin-gonic/gin"
)

type ISampleRepo interface {
	ListSample(ctx context.Context, param models.SampleParam) ([]models.Sample, error)
	GetSampleByID(ctx context.Context, id int) (models.Sample, error)
	CreateSample(ctx context.Context, sample models.Sample) error
	UpdateSample(ctx context.Context, sample models.Sample, id int) error
	DeleteSample(ctx context.Context, id int) error
}

type ISampleService interface {
	ListSample(ctx context.Context, param models.SampleParam) ([]models.Sample, error)
	GetSample(ctx context.Context, id int) (models.Sample, error)
	CreateSample(ctx context.Context, sample models.Sample) error
	UpdateSample(ctx context.Context, sample models.Sample, id int) error
	DeleteSample(ctx context.Context, id int) error
}

type ISampleHandler interface {
	ListSample(c *gin.Context)
	GetSample(c *gin.Context)
	CreateSample(c *gin.Context)
	UpdateSample(c *gin.Context)
	DeleteSample(c *gin.Context)
}
