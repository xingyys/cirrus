package store

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/lack-io/cirrus/config"
)

var (
	ErrInvalidConfig = errors.New("invalid config")
	ErrConnectDB     = errors.New("connect database")
	ErrDBWrite       = errors.New("write to database")
	ErrDBRead        = errors.New("read from database")
)

type Pagination struct {
	Page int `json:"page"`

	Size int `json:"size"`

	Total int64 `json:"total"`
}

// 商品信息
type Good struct {
	ID uint64 `gorm:"column:id;primaryKey"`

	// UID 唯一ID
	UID string `json:"uid" gorm:"column:uid"`

	// URL 所在网址
	URL string `json:"url" gorm:"column:url"`

	ScaleOut bool `json:"scaleOut" gorm:"column:scaleout"`

	Brandless bool `json:"brandless" gorm:"column:"brandless"`

	// Comments 评论数
	Comments int `json:"comments" gorm:"column:comments"`

	// Express 快递信息
	Express string `json:"express" gorm:"column:express"`

	// 入库时间
	Timestamp int64 `json:"timestamp" gorm:"column:timestamp"`
}

type Store struct {
	cfg *config.Store

	db *gorm.DB
}

func NewStore(cfg *config.Store) (*Store, error) {
	s := &Store{cfg: cfg}

	switch cfg.DB {
	case config.Sqlite:
		if cfg.Sqlite == nil || cfg.Sqlite.Name == "" {
			return nil, fmt.Errorf("%w: missing cfg sqlite", ErrInvalidConfig)
		}
		var err error
		s.db, err = gorm.Open(sqlite.Open(cfg.Sqlite.Name), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrConnectDB, err)
		}
	}

	DB, err := s.db.DB()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectDB, err)
	}

	DB.SetConnMaxLifetime(time.Second * 5)
	DB.SetMaxIdleConns(50)
	DB.SetMaxOpenConns(10)

	err = s.db.Table("goods").AutoMigrate(&Good{})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDBWrite, err)
	}

	return s, nil
}

func (s *Store) GetGoods(pg *Pagination) ([]*Good, error) {
	goods := make([]*Good, 0)

	db := s.db.Table("goods").Order("timestamp desc")
	if pg != nil {
		_ = db.Count(&pg.Total)
		db = db.Limit(pg.Size).Offset((pg.Page - 1) * pg.Size)
	}

	err := db.Find(&goods).Error
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDBRead, err)
	}

	return goods, nil
}

func (s *Store) GetGoodsByTimeout(start, end int64, pg *Pagination) ([]*Good, error) {
	goods := make([]*Good, 0)

	db := s.db.Table("goods").Order("timestamp desc")
	if pg != nil {
		db = db.Limit(pg.Size).Offset((pg.Page - 1) * pg.Size)
	}

	db = db.Where("timestamp > ? AND timestamp < ?", start, end)
	if pg != nil {
		_ = db.Count(&pg.Total)
		db = db.Limit(pg.Size).Offset((pg.Page - 1) * pg.Size)
	}

	err := db.Find(&goods).Error
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDBRead, err)
	}

	return goods, nil
}

func (s *Store) AddGood(good *Good) error {
	err := s.db.Table("goods").Create(good).Error
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDBWrite, err)
	}

	return nil
}

func (s *Store) DelGroup(id int64) (*Good, error) {
	good := &Good{}
	err := s.db.Table("goods").Delete(good, "id = ?", id).Error
	return good, err
}

func (s *Store) Reset() {

}

