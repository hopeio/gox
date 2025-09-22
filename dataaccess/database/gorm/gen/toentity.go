package gen

import (
	"fmt"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func Generator(db *gorm.DB, outPath, modelPkgPath string, tables ...string) error {
	g := gen.NewGenerator(gen.Config{
		OutPath:           outPath,
		ModelPkgPath:      modelPkgPath,
		WithUnitTest:      false,
		FieldNullable:     false,
		FieldCoverable:    false,
		FieldWithIndexTag: true,
		FieldWithTypeTag:  true,
		FieldSignable:     true,
	})

	g.UseDB(db)

	if len(tables) == 0 {
		// Execute tasks for all tables in the database
		var err error
		tables, err = db.Migrator().GetTables()
		if err != nil {
			return fmt.Errorf("GORM migrator get all tables fail: %w", err)
		}
	}

	// Execute some data table tasks
	models := make([]interface{}, len(tables))
	for i, tableName := range tables {
		models[i] = g.GenerateModel(tableName)
	}
	g.ApplyBasic(models...)

	g.Execute()
	return nil
}
