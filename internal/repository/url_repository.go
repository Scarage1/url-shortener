package repository

import (
	"github.com/Scarage1/url-shortener/internal/model"

	"gorm.io/gorm"
)


type URLRepository struct {
	DB *gorm.DB
}


func NewURLRepository(db *gorm.DB) *URLRepository {

	return &URLRepository{
		DB: db,
	}
}


func (r *URLRepository) Create(url *model.URL) error {

	return r.DB.Create(url).Error
}

func (r *URLRepository) FindByOriginalURL(
	originalURL string,
) (*model.URL, error) {


	var url model.URL


	err := r.DB.Where(
		"original_url = ?",
		originalURL,
	).First(&url).Error


	if err != nil {
		return nil, err
	}


	return &url, nil
}


func (r *URLRepository) FindByShortCode(code string) (*model.URL, error) {

	var url model.URL


	err := r.DB.Where(
		"short_code = ?",
		code,
	).First(&url).Error


	if err != nil {
		return nil, err
	}


	return &url, nil
}

func (r *URLRepository) Update(
	url *model.URL,
) error {

	return r.DB.Save(url).Error
}