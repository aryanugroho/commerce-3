package page

import (
	"time"

	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
)

// GetID returns id for cms block
func (it *DefaultCMSPage) GetID() string {
	return it.id
}

// SetID sets id for cms block
func (it *DefaultCMSPage) SetID(newID string) error {
	it.id = newID
	return nil
}

// Load loads cms page information from DB
func (it *DefaultCMSPage) Load(id string) error {
	collection, err := db.GetCollection(ConstCmsPageCollectionName)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	dbValues, err := collection.LoadByID(id)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	it.SetID(utils.InterfaceToString(dbValues["_id"]))

	it.Identifier = utils.InterfaceToString(dbValues["identifier"])
	it.Enabled = utils.InterfaceToBool(dbValues["enabled"])

	it.Title = utils.InterfaceToString(dbValues["title"])
	it.Content = utils.InterfaceToString(dbValues["content"])

	it.CreatedAt = utils.InterfaceToTime(dbValues["created_at"])
	it.UpdatedAt = utils.InterfaceToTime(dbValues["updated_at"])

	return nil
}

// Delete removes current cms block from DB
func (it *DefaultCMSPage) Delete() error {
	collection, err := db.GetCollection(ConstCmsPageCollectionName)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = collection.DeleteByID(it.GetID())
	if err != nil {
		return env.ErrorDispatch(err)
	}

	return env.ErrorDispatch(err)
}

// Save stores current cms block to DB
func (it *DefaultCMSPage) Save() error {
	collection, err := db.GetCollection(ConstCmsPageCollectionName)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	// packing data before save
	currentTime := time.Now()

	storingValues := make(map[string]interface{})

	storingValues["_id"] = it.GetID()

	storingValues["enabled"] = it.GetEnabled()

	storingValues["identifier"] = it.GetIdentifier()

	storingValues["title"] = it.GetTitle()
	storingValues["content"] = it.GetContent()

	if it.CreatedAt.IsZero() {
		it.CreatedAt = currentTime
	}

	storingValues["created_at"] = it.CreatedAt
	storingValues["updated_at"] = currentTime

	newID, err := collection.Save(storingValues)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	it.SetID(newID)

	return nil
}
