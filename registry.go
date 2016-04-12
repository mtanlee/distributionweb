package distributionweb

import (
	"fmt"
	registry "github.com/mtanlee/distributionweb/registry/v2"
)

type Registry struct {
	ID             string                   `json:"id,omitempty" gorethink:"id,omitempty"`
	Name           string                   `json:"name,omitempty" gorethink:"name,omitempty"`
	Addr           string                   `json:"addr,omitempty", gorethink:"addr,omitempty"`
	User           string                   `json:"user,omitempty", gorethink:"user,omitempty"`
	Password       string                   `json:"password,omitempty", gorethink:"passwd,omitempty`
	Email          string                   `json:"email,omitempty", gorethink:"email,omitempty"`
	registryClient *registry.RegistryClient `json:"-" gorethink:"-"`
}

func NewRegistry(id, name, addr, user, password, email string) (*Registry, error) {
	rClient, err := registry.NewRegistryClient(fmt.Sprintf(addr), nil)
	if err != nil {
		return nil, err
	}

	return &Registry{
		ID:             id,
		Name:           name,
		Addr:           addr,
		User:           user,
		Password:       password,
		Email:          email,
		registryClient: rClient,
	}, nil
}

func (r *Registry) Repositories() ([]*registry.Repository, error) {
	user := r.User
	passwd := r.Password
	res, err := r.registryClient.Search("", user, passwd)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *Registry) Repository(name, tag string) (*registry.Repository, error) {
	user := r.User
	passwd := r.Password
	return r.registryClient.Repository(name, tag, user, passwd)
}

func (r *Registry) DeleteRepository(name string) error {
	//	user := r.User
	//	passwd := r.Password
	return r.registryClient.DeleteRepository(name)
}
