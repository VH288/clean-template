package migrate

import "clean-template/internal/models"

func MigrateList() []any {
	return []any{&models.Sample{}}
}
