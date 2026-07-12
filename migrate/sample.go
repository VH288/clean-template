package migrate

import "clean-template/internal/entity"

func getSampleModel() (res []any) {
	res = append(res, &entity.Sample{})
	return res
}
