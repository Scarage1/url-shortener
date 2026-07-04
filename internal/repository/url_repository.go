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
	userID uint,
) (*model.URL, error) {

	var url model.URL

	err :=
		r.DB.Where(
			"original_url = ? AND user_id = ?",
			originalURL,
			userID,
		).First(
			&url,
		).Error

	return &url, err
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


func (r *URLRepository) FindByUser(
	userID uint,
) ([]model.URL, error) {

	var urls []model.URL

	err :=
		r.DB.Where(
			"user_id = ?",
			userID,
		).
			Order(
				"created_at DESC",
			).
			Find(
				&urls,
			).Error

	return urls, err
}

func (r *URLRepository) Update(
	url *model.URL,
) error {

	return r.DB.Save(url).Error
}
