package dbmigrate

import (
	"log"
	"time"

	"gorm.io/gorm"
)

type Migration struct {
	ID         uint   `gorm:"primarykey"`
	Name       string `gorm:"index:,unique"`
	ExecutedBy string
	RanAt      time.Time
}

type Client struct {
	DB *gorm.DB
}

func NewClient(db *gorm.DB) *Client {
	return &Client{DB: db}
}

func (c *Client) InitializeMigrationTracking() error {
	return c.DB.AutoMigrate(&Migration{})
}

func (c *Client) PerformMigration(migrationName, executedBy string, migrationFunc func(*gorm.DB) error) error {
	var migration Migration
	result := c.DB.Where("name = ?", migrationName).First(&migration)
	if result.RowsAffected == 0 {
		err := migrationFunc(c.DB)
		if err != nil {
			log.Printf("Migration failed (%s): %v\n", migrationName, err)
			return err
		}
		c.DB.Create(&Migration{Name: migrationName, RanAt: time.Now(), ExecutedBy: executedBy})
		log.Printf("Migration executed and recorded: %s\n", migrationName)
	}
	log.Printf("Migration already executed: %s\n", migrationName)
	return nil
}
