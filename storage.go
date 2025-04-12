package paltergeist

// Bucket storage.
type Bucket struct {
	Name         string `json:"name"`
	Location     string `json:"location"`
	StorageClass string `json:"storageClass"`
}

func (b Bucket) TrapId() string {
	return b.Name
}

func (b Bucket) Kind() Kind {
	return KindStorage
}

func (b Bucket) Type() Type {
	return BucketResource
}
