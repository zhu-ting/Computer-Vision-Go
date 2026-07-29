package repository

import "gorm.io/gorm"

// gormDB is a type alias so repository functions can accept *gorm.DB
// without importing gorm in every file's import block — the actual
// import lives here.
type gormDB = gorm.DB
