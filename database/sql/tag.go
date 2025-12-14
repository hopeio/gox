package sql

import (
	"reflect"

	"github.com/hopeio/gox/reflect/structtag"
)

const (
	CondiTagName = "sqlcond" // e.g: `sqlcond:"column:id;operate:="`
	// e.g: `sqlcond:"operate:id = ?"`
	// e.g: `sqlcond:"-"`
	// e.g: `sqlcond:"embedded"`
)

type ConditionTag struct {
	Column     string `meta:"column"`
	Operate    string `meta:"operate"`
	EmptyValid bool   `meta:"emptyvalid"`
}

func GetConditionTagTag(tag reflect.StructTag) (*ConditionTag, error) {
	return structtag.ParseSettingTagToStruct[ConditionTag](tag.Get(CondiTagName), ';')
}
