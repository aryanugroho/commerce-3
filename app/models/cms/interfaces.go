// Package cms represents abstraction of business layer cms page and cms block objects
package cms

import (
	"github.com/ottemo/commerce/app/models"
	"github.com/ottemo/commerce/env"
)

// Package global constants
const (
	ConstModelNameCMSPage            = "CMSPage"
	ConstModelNameCMSPageCollection  = "CMSPageCollection"
	ConstModelNameCMSBlock           = "CMSBlock"
	ConstModelNameCMSBlockCollection = "CMSBlockCollection"

	ConstErrorModule = "cms"
	ConstErrorLevel  = env.ConstErrorLevelModel
)

// InterfaceCMSPage represents interface to access business layer implementation of cms page object
type InterfaceCMSPage interface {
	GetEnabled() bool
	SetEnabled(bool) error

	LoadByIdentifier(identifier string) error

	GetIdentifier() string
	SetIdentifier(string) error

	GetTitle() string
	SetTitle(string) error

	GetContent() string
	SetContent(string) error

	EvaluateContent() string

	models.InterfaceModel
	models.InterfaceObject
	models.InterfaceStorable
	models.InterfaceListable
}

// InterfaceCMSPageCollection represents interface to access business layer implementation of cms page collection
type InterfaceCMSPageCollection interface {
	ListCMSPages() []InterfaceCMSPage

	models.InterfaceCollection
}

// InterfaceCMSBlock represents interface to access business layer implementation of cms block object
type InterfaceCMSBlock interface {
	LoadByIdentifier(identifier string) error

	GetIdentifier() string
	SetIdentifier(string) error

	GetContent() string
	SetContent(string) error

	EvaluateContent() string

	models.InterfaceModel
	models.InterfaceObject
	models.InterfaceStorable
	models.InterfaceListable
}

// InterfaceCMSBlockCollection represents interface to access business layer implementation of cms block collection
type InterfaceCMSBlockCollection interface {
	ListCMSBlocks() []InterfaceCMSBlock

	models.InterfaceCollection
}
