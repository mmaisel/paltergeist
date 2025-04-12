package paltergeist

// User is a human IAM account
type User struct {
	// Name of the User.
	Name string
	// Email uniquely identifies the User
	Email string
}

func (u User) TrapId() string {
	return u.Email
}

func (u User) Kind() Kind {
	return KindIdentity
}

func (u User) Type() Type {
	return UserResource
}

// ServiceAccount is a non-human IAM account.
type ServiceAccount struct {
	// Id is the name that uniquely identifies the ServiceAccount.
	Id string `json:"id" description:"The name that uniquely identifies the ServiceAccount."`
	// Name is the friendly name of the ServiceAccount.
	Name string `json:"name" description:"Unique identifier for the book"`
	// Description provides a brief explanation of the purpose or functionality of the ServiceAccount.
	Description string `json:"description"`
	// Email of the service account.
	Email string `json:"email"`
}

func (sa ServiceAccount) TrapId() string {
	return sa.Id
}

func (sa ServiceAccount) Kind() Kind {
	return KindIdentity
}

func (sa ServiceAccount) Type() Type {
	return ServiceAccountResource
}

type RoleBinding struct {
	// PrincipalId is the Id of the ServiceAccount or User.
	PrincipalId string
	// Role granted to the Member.
	Role string
}
