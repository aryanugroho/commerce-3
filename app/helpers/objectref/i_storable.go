package objectref

import "github.com/ottemo/foundation/env"

func (it *DBObjectRef) Save() error {
	return env.ErrorNew("not implemented")
}

func (it *DBObjectRef) Load(id string) error {
	return env.ErrorNew("not implemented")
}

func (it *DBObjectRef) Delete() error {
	return env.ErrorNew("not implemented")
}