package repository

import (
	"github.com/jinzhu/gorm"
)

type Repository interface {
	GetAll(uow *UOW, out interface{}, queryProcessors ...QueryProcessor) error
	GetByID(uow *UOW, out interface{}, id interface{}) error
	Add(uow *UOW, out interface{}) error
	Limit(limit interface{}) QueryProcessor
	Offset(limit interface{}) QueryProcessor
	Filter(condition string, args ...interface{}) QueryProcessor
	Count(limit, offset int, totalCount *int) QueryProcessor
	Update(uow *UOW, updated_value interface{}, id interface{}) error
	DeleteById(uow *UOW, id interface{}) error
}

type GormRepositoryMySQL struct {
}

func NewGormRepositoryMySQL() Repository {
	return &GormRepositoryMySQL{}
}
func executeQueryProcessors(db *gorm.DB, out interface{}, queryProcessors ...QueryProcessor) (*gorm.DB, error) {
	var err error
	for i := 0; i < len(queryProcessors); i++ {
		db, err = queryProcessors[i](db, out)
		if err != nil {
			return nil, err
		}
	}
	return db, nil

}

func (g *GormRepositoryMySQL) Count(limit, offset int, totalCount *int) QueryProcessor {

	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		// db.Limit(-1)
		// db.Offset(-1)
		if totalCount != nil {
			err := db.Model(out).Count(totalCount).Error
			if err != nil {
				return db, err
			}

		}
		return db, nil

	}

}
func (g *GormRepositoryMySQL) GetAll(uow *UOW, out interface{}, queryProcessors ...QueryProcessor) error {

	db, err := executeQueryProcessors(uow.DB, out, queryProcessors...)
	if err != nil {
		return err
	}
	return db.Find(out).Error

}
func (g *GormRepositoryMySQL) GetByID(uow *UOW, out interface{}, id interface{}) error {
	return uow.DB.First(out, id).Error

}
func (g *GormRepositoryMySQL) Add(uow *UOW, out interface{}) error {
	return uow.DB.Create(out).Error
}

func (g *GormRepositoryMySQL) Limit(limit interface{}) QueryProcessor {

	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		db = db.Limit(limit)
		return db, nil
	}
}
func (g *GormRepositoryMySQL) Offset(Offset interface{}) QueryProcessor {

	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		db = db.Offset(Offset)
		return db, nil
	}
}
func (g *GormRepositoryMySQL) Filter(condition string, args ...interface{}) QueryProcessor {

	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		db = db.Where(condition, args...)
		return db, nil
	}
}
func (g *GormRepositoryMySQL) Update(uow *UOW, updated_value interface{}, id interface{}) error {
	return uow.DB.Save(updated_value).Error
}

func (g *GormRepositoryMySQL) DeleteById(uow *UOW, out interface{}, id interface{}) error {
	// var tempInt interface{}
	// output:=g.GetByID(uow)
	return uow.DB.Delete(out, id).Error
}

type UOW struct {
	DB       *gorm.DB
	Commited bool
}

func NewUnitOfWork(DB *gorm.DB) *UOW {
	return &UOW{DB: DB.Begin(), Commited: false}
}

func (u *UOW) RollBack() {
	if u.Commited {
		return
	}
	u.DB.Rollback()
}
func (u *UOW) Commit() {
	if u.Commited {
		return
	}
	u.DB.Commit()
	u.Commited = true
}
