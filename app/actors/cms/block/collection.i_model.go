package block

import (
	"github.com/ottemo/commerce/app/models"
	"github.com/ottemo/commerce/app/models/cms"
	"github.com/ottemo/commerce/db"
	"github.com/ottemo/commerce/env"
)

// GetModelName returns model name
func (it *DefaultCMSBlockCollection) GetModelName() string {
	return cms.ConstModelNameCMSBlockCollection
}

// GetImplementationName returns model implementation name
func (it *DefaultCMSBlockCollection) GetImplementationName() string {
	return "Default" + cms.ConstModelNameCMSBlockCollection
}

// New returns new instance of model implementation object
func (it *DefaultCMSBlockCollection) New() (models.InterfaceModel, error) {
	dbCollection, err := db.GetCollection(ConstCmsBlockCollectionName)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return &DefaultCMSBlockCollection{listCollection: dbCollection, listExtraAtributes: make([]string, 0)}, nil
}
