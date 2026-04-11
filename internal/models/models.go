package models

import (
	"time"
)

type SkillLevel string

const (
	Beginner     SkillLevel = "BEGINNER"
	Intermediate SkillLevel = "INTERMEDIATE"
	Advanced     SkillLevel = "ADVANCED"
)

// 1. Admin
type Admin struct {
	ID       string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Username string `gorm:"unique;not null" json:"username"`
	Password string `gorm:"not null" json:"-"`
}

// 2. CV
type CV struct {
	ID            string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FileURL       string    `gorm:"not null" json:"file_url"`
	DownloadCount int       `gorm:"default:0" json:"download_count"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// 3. Skill
type SkillCategory struct {
	ID     string  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name   string  `gorm:"not null" json:"name"`
	Skills []Skill `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE;" json:"skills"`
}

type Skill struct {
	ID         string        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name       string        `gorm:"not null" json:"name"`
	IconURL    *string       `json:"icon_url"`
	Level      SkillLevel    `gorm:"type:text;default:BEGINNER" json:"level"`
	CategoryID string        `gorm:"type:uuid" json:"category_id"`
	Category   SkillCategory `gorm:"foreignKey:CategoryID" json:"-"`
}

// 4. ProjectTechStack — master data pool
type ProjectTechStack struct {
	ID      string  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name    string  `gorm:"not null;uniqueIndex" json:"name"`
	IconURL *string `json:"icon_url"`
	SkillID *string `gorm:"type:uuid" json:"skill_id"`
	Skill   *Skill  `gorm:"foreignKey:SkillID" json:"skill,omitempty"`

	Projects []Project `gorm:"many2many:project_tech_stack_items;" json:"-"`
}

func (ProjectTechStack) TableName() string {
	return "project_tech_stack_pool"
}

// 5. Project
type Project struct {
	ID          string             `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Title       string             `gorm:"not null" json:"title"`
	ShortStory  string             `gorm:"type:text" json:"short_story"`
	Description string             `gorm:"type:text" json:"description"`
	IsFeatured  bool               `gorm:"default:false" json:"is_featured"`
	ProjectDate time.Time          `gorm:"default:now()" json:"project_date"` // editable, bulan+tahun project
	CreatedAt   time.Time          `json:"created_at"`                   // auto, kapan data diinput
	Images      []ProjectImage     `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;" json:"images"`
	Links       []ProjectLink      `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;" json:"links"`
	TechStack   []ProjectTechStack `gorm:"many2many:project_tech_stack_items;" json:"tech_stack"`
}

type ProjectImage struct {
	ID          string  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ImageURL    string  `gorm:"not null" json:"image_url"`
	StoragePath string  `gorm:"default:''" json:"storage_path"`
	SortOrder   int     `gorm:"default:0" json:"sort_order"`
	ProjectID   string  `gorm:"type:uuid;not null" json:"project_id"`
	Project     Project `json:"-"`
}

type ProjectLink struct {
	ID        string  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Label     string  `gorm:"not null" json:"label"`
	URL       string  `gorm:"not null" json:"url"`
	ProjectID string  `gorm:"type:uuid;not null" json:"project_id"`
	Project   Project `json:"-"`
}

// 6. Experience
type Experience struct {
	ID        string            `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Title     string            `gorm:"not null" json:"title"`
	Location  string            `gorm:"not null" json:"location"`
	StartDate time.Time         `json:"start_date"`
	EndDate   *time.Time        `json:"end_date"`
	Points    []ExperiencePoint `gorm:"foreignKey:ExperienceID;constraint:OnDelete:CASCADE;" json:"points"`
}

type ExperiencePoint struct {
	ID           string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Content      string     `gorm:"type:text" json:"content"`
	ExperienceID string     `gorm:"type:uuid" json:"experience_id"`
	Experience   Experience `json:"-"`
}

// 7. Certificate
type Certificate struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Title       string    `gorm:"not null" json:"title"`
	Issuer      string    `gorm:"not null" json:"issuer"`
	Date        time.Time `json:"date"`
	FileURL     string    `gorm:"default:''" json:"file_url"`
	StoragePath string    `gorm:"default:''" json:"storage_path"` // path di Supabase Storage untuk delete
	LinkURL     *string   `json:"link_url"`
}

// 8. Contact
type ContactMessage struct {
	ID      string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name    string    `gorm:"not null" json:"name"`
	Email   string    `gorm:"not null" json:"email"`
	Message string    `gorm:"type:text" json:"message"`
	IsRead  bool      `gorm:"default:false" json:"is_read"`
	SentAt  time.Time `json:"sent_at"`
}