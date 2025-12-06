package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AchievementType enum
type AchievementType string

const (
	AchievementTypeAcademic       AchievementType = "academic"
	AchievementTypeCompetition     AchievementType = "competition"
	AchievementTypeOrganization   AchievementType = "organization"
	AchievementTypePublication    AchievementType = "publication"
	AchievementTypeCertification  AchievementType = "certification"
	AchievementTypeOther          AchievementType = "other"
)

// CompetitionLevel enum
type CompetitionLevel string

const (
	CompetitionLevelInternational CompetitionLevel = "international"
	CompetitionLevelNational       CompetitionLevel = "national"
	CompetitionLevelRegional       CompetitionLevel = "regional"
	CompetitionLevelLocal          CompetitionLevel = "local"
)

// PublicationType enum
type PublicationType string

const (
	PublicationTypeJournal    PublicationType = "journal"
	PublicationTypeConference PublicationType = "conference"
	PublicationTypeBook       PublicationType = "book"
)

// Attachment model sesuai spesifikasi
type Attachment struct {
	FileName    string    `bson:"fileName" json:"file_name"`
	FileURL     string    `bson:"fileUrl" json:"file_url"`
	FileType    string    `bson:"fileType" json:"file_type"`
	UploadedAt  time.Time `bson:"uploadedAt" json:"uploaded_at"`
}

// Period model untuk organization
type Period struct {
	Start time.Time `bson:"start" json:"start"`
	End   time.Time `bson:"end" json:"end"`
}

// AchievementDetails untuk field dinamis berdasarkan tipe prestasi
// Sesuai spesifikasi dengan camelCase untuk MongoDB
type AchievementDetails struct {
	// Untuk competition
	CompetitionName  *string          `bson:"competitionName,omitempty" json:"competition_name,omitempty"`
	CompetitionLevel *CompetitionLevel `bson:"competitionLevel,omitempty" json:"competition_level,omitempty"` // 'international', 'national', 'regional', 'local'
	Rank             *int             `bson:"rank,omitempty" json:"rank,omitempty"`
	MedalType        *string          `bson:"medalType,omitempty" json:"medal_type,omitempty"`

	// Untuk publication
	PublicationType  *PublicationType `bson:"publicationType,omitempty" json:"publication_type,omitempty"` // 'journal', 'conference', 'book'
	PublicationTitle *string          `bson:"publicationTitle,omitempty" json:"publication_title,omitempty"`
	Authors          []string         `bson:"authors,omitempty" json:"authors,omitempty"`
	Publisher        *string          `bson:"publisher,omitempty" json:"publisher,omitempty"`
	ISSN             *string          `bson:"issn,omitempty" json:"issn,omitempty"`

	// Untuk organization
	OrganizationName *string  `bson:"organizationName,omitempty" json:"organization_name,omitempty"`
	Position         *string  `bson:"position,omitempty" json:"position,omitempty"`
	Period           *Period  `bson:"period,omitempty" json:"period,omitempty"` // { start: Date, end: Date }

	// Untuk certification
	CertificationName   *string    `bson:"certificationName,omitempty" json:"certification_name,omitempty"`
	IssuedBy            *string    `bson:"issuedBy,omitempty" json:"issued_by,omitempty"`
	CertificationNumber *string    `bson:"certificationNumber,omitempty" json:"certification_number,omitempty"`
	ValidUntil          *time.Time `bson:"validUntil,omitempty" json:"valid_until,omitempty"`

	// Field umum yang bisa ada
	EventDate    *time.Time             `bson:"eventDate,omitempty" json:"event_date,omitempty"`
	Location     *string                `bson:"location,omitempty" json:"location,omitempty"`
	Organizer    *string                `bson:"organizer,omitempty" json:"organizer,omitempty"`
	Score        *float64               `bson:"score,omitempty" json:"score,omitempty"`
	CustomFields map[string]interface{} `bson:"customFields,omitempty" json:"custom_fields,omitempty"` // untuk field tambahan yang tidak terdefinisi
}

// Achievement model untuk MongoDB
// Sesuai spesifikasi: Collection achievements dengan field dinamis berdasarkan tipe prestasi
type Achievement struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	StudentID       string             `bson:"studentId" json:"student_id"` // UUID reference to PostgreSQL (camelCase di MongoDB)
	AchievementType AchievementType    `bson:"achievementType" json:"achievement_type"`
	Title           string             `bson:"title" json:"title"`
	Description     string             `bson:"description" json:"description"`
	Details         AchievementDetails  `bson:"details" json:"details"` // Field dinamis berdasarkan tipe prestasi
	Attachments     []Attachment        `bson:"attachments" json:"attachments"`
	Tags            []string            `bson:"tags" json:"tags"`
	Points          float64             `bson:"points" json:"points"` // poin prestasi untuk keperluan scoring
	Status          AchievementStatus   `bson:"status" json:"status"` // Untuk workflow: draft, submitted, verified, rejected
	CreatedAt       time.Time           `bson:"createdAt" json:"created_at"`
	UpdatedAt       time.Time           `bson:"updatedAt" json:"updated_at"`
	DeletedAt       *time.Time          `bson:"deletedAt,omitempty" json:"deleted_at,omitempty"` // Soft delete
}
