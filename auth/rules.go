package auth

type (
	ACL struct {
		RoleName    string        `json:"role_name,omitempty"`
		Description string        `json:"description,omitempty"`
		Rules       []*AccessRule `json:"rules,omitempty"`
	}

	AccessRule struct {
		Path    string   `json:"path,omitempty"`
		Methods []string `json:"methods,omitempty"`
	}
)

func DefaultACLs() []*ACL {
	acls := []*ACL{}
	adminACL := &ACL{
		RoleName:    "admin",
		Description: "Administrator",
		Rules: []*AccessRule{
			{
				Path:    "*",
				Methods: []string{"*"},
			},
		},
	}
	acls = append(acls, adminACL)

	eventsACLRO := &ACL{
		RoleName:    "events:ro",
		Description: "Events Read Only",
		Rules: []*AccessRule{
			{
				Path:    "/api/events",
				Methods: []string{"GET"},
			},
		},
	}
	acls = append(acls, eventsACLRO)

	eventsACLRW := &ACL{
		RoleName:    "events:rw",
		Description: "Events",
		Rules: []*AccessRule{
			{
				Path:    "/api/events",
				Methods: []string{"GET", "POST", "DELETE"},
			},
		},
	}
	acls = append(acls, eventsACLRW)
	registriesACLRO := &ACL{
		RoleName:    "registries:ro",
		Description: "Registries Read Only",
		Rules: []*AccessRule{
			{
				Path:    "/api/registry",
				Methods: []string{"GET"},
			},
		},
	}
	acls = append(acls, registriesACLRO)

	registriesACLRW := &ACL{
		RoleName:    "registries:rw",
		Description: "Registries",
		Rules: []*AccessRule{
			{
				Path:    "/api/registry",
				Methods: []string{"GET", "POST", "DELETE"},
			},
		},
	}
	acls = append(acls, registriesACLRW)

	return acls
}
