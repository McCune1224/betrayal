package data

import "gorm.io/gorm"

type Insult struct {
	gorm.Model
	Insult string `gorm:"not null"`
}

type InsultModel struct {
	DB *gorm.DB
}

func (im *InsultModel) Insert(insult Insult) error {
	return im.DB.Create(&insult).Error
}

func (im *InsultModel) GetRandomInsult() (Insult, error) {
	var i Insult
	return i, im.DB.Order("RANDOM()").First(&i).Error
}

func (im *InsultModel) GetTotalInsults() (int64, error) {
	var count int64
	return count, im.DB.Model(&Insult{}).Count(&count).Error
}

func (im *InsultModel) DeleteInsult(id uint) error {
	return im.DB.Delete(&Insult{}, id).Error
}
