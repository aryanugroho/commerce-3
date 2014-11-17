package db

const (
	DB_BASETYPE_ID       = "id"
	DB_BASETYPE_BOOLEAN  = "bool"
	DB_BASETYPE_VARCHAR  = "varchar"
	DB_BASETYPE_TEXT     = "text"
	DB_BASETYPE_INTEGER  = "int"
	DB_BASETYPE_DECIMAL  = "decimal"
	DB_BASETYPE_MONEY    = "money"
	DB_BASETYPE_FLOAT    = "float"
	DB_BASETYPE_DATETIME = "datetime"
	DB_BASETYPE_JSON     = "json"
)

type I_DBEngine interface {
	GetName() string

	CreateCollection(Name string) error
	GetCollection(Name string) (I_DBCollection, error)
	HasCollection(Name string) bool

	RawQuery(query string) (map[string]interface{}, error)
}

type I_DBCollection interface {
	Load() ([]map[string]interface{}, error)
	LoadById(id string) (map[string]interface{}, error)

	Save(map[string]interface{}) (string, error)

	Delete() (int, error)
	DeleteById(id string) error

	Iterate(iteratorFunc func(record map[string]interface{}) bool) error

	Count() (int, error)
	Distinct(columnName string) ([]interface{}, error)

	SetupFilterGroup(groupName string, orSequence bool, parentGroup string) error
	RemoveFilterGroup(groupName string) error
	AddGroupFilter(groupName string, columnName string, operator string, value interface{}) error

	AddStaticFilter(columnName string, operator string, value interface{}) error
	AddFilter(columnName string, operator string, value interface{}) error

	ClearFilters() error

	AddSort(columnName string, Desc bool) error
	ClearSort() error

	SetResultColumns(columns ...string) error

	SetLimit(offset int, limit int) error

	ListColumns() map[string]string
	GetColumnType(columnName string) string
	HasColumn(columnName string) bool

	AddColumn(columnName string, columnType string, indexed bool) error
	RemoveColumn(columnName string) error
}