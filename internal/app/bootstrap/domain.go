package bootstrap

type Domain struct {
	Sample      *Sample
	Healthcheck *Healthcheck
}

func getDomain(infra *Infra) Domain {
	return Domain{
		Sample:      wireSample(infra),
		Healthcheck: wireHealthcheck(infra),
	}
}
