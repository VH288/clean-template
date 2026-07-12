package migrate

func MigrateList() (res []any) {
	res = append(res, getSampleModel()...)
	return res
}
