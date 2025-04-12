package paltergeist

// Cloud Run Service.
type CloudRunService struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

func (c CloudRunService) TrapId() string {
	return c.Name
}

func (c CloudRunService) Kind() Kind {
	return KindCompute
}

func (c CloudRunService) Type() Type {
	return CloudRunServiceResource
}
