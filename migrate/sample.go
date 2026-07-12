package migrate

import "clean-template/internal/models"

func getSampleModel() (res []any) {
	res = append(res, &models.Sample{})
	return res
}
