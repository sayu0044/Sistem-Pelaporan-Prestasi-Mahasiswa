package service

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/repository"
)

type AchievementService interface {
	CreateAchievement(ctx context.Context, userID uuid.UUID, req *CreateAchievementRequest) (*AchievementResponse, error)
	UpdateAchievement(ctx context.Context, userID uuid.UUID, achievementID string, req *UpdateAchievementRequest) (*AchievementResponse, error)
	DeleteAchievement(ctx context.Context, userID uuid.UUID, achievementID string) error
	SubmitAchievement(ctx context.Context, userID uuid.UUID, achievementID string) (*AchievementResponse, error)
	VerifyAchievement(ctx context.Context, userID uuid.UUID, achievementID string) (*AchievementResponse, error)
	RejectAchievement(ctx context.Context, userID uuid.UUID, achievementID string, rejectionNote string) (*AchievementResponse, error)
	GetAchievements(ctx context.Context, userID uuid.UUID, page, limit int, status string) (*AchievementListResponse, error)
	GetAchievementByID(ctx context.Context, userID uuid.UUID, achievementID string) (*AchievementResponse, error)
	GetAchievementHistory(ctx context.Context, userID uuid.UUID, achievementID string) ([]AchievementHistoryResponse, error)
	UploadAttachment(ctx context.Context, userID uuid.UUID, achievementID string, filePath string) (string, error)
}

type achievementService struct {
	achievementRepo     repository.AchievementRepository
	historyRepo         repository.AchievementHistoryRepository
	studentRepo         repository.StudentRepository
	lecturerRepo        repository.LecturerRepository
	userRepo            repository.UserRepository
	roleRepo            repository.RoleRepository
}

func NewAchievementService(
	achievementRepo repository.AchievementRepository,
	historyRepo repository.AchievementHistoryRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
) AchievementService {
	return &achievementService{
		achievementRepo: achievementRepo,
		historyRepo:      historyRepo,
		studentRepo:      studentRepo,
		lecturerRepo:     lecturerRepo,
		userRepo:         userRepo,
		roleRepo:         roleRepo,
	}
}

// Request/Response DTOs
type CreateAchievementRequest struct {
	AchievementType model.AchievementType  `json:"achievement_type"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Details         model.AchievementDetails `json:"details"`
	Tags            []string               `json:"tags,omitempty"`
	Points          float64                `json:"points,omitempty"`
}

type UpdateAchievementRequest struct {
	AchievementType model.AchievementType  `json:"achievement_type,omitempty"`
	Title           string                 `json:"title,omitempty"`
	Description     string                 `json:"description,omitempty"`
	Details         *model.AchievementDetails `json:"details,omitempty"`
	Tags            []string               `json:"tags,omitempty"`
	Points          *float64                `json:"points,omitempty"`
}

type AchievementResponse struct {
	ID              string                 `json:"id"`
	StudentID       string                 `json:"student_id"`
	Student         *StudentInfo           `json:"student,omitempty"`
	AchievementType model.AchievementType  `json:"achievement_type"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Details         model.AchievementDetails `json:"details"`
	Attachments     []model.Attachment     `json:"attachments"`
	Tags            []string               `json:"tags"`
	Points          float64                `json:"points"`
	Status          model.AchievementStatus `json:"status"`
	Reference       *AchievementReferenceInfo `json:"reference,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

type AchievementListResponse struct {
	Data       []AchievementResponse `json:"data"`
	Page       int                   `json:"page"`
	Limit      int                   `json:"limit"`
	Total      int64                 `json:"total"`
	TotalPages int                   `json:"total_pages"`
}

type AchievementHistoryResponse struct {
	ID                 string                  `json:"id"`
	OldStatus          *model.AchievementStatus `json:"old_status,omitempty"` // Nullable untuk status awal
	NewStatus          model.AchievementStatus  `json:"new_status"`
	ChangedBy          string                   `json:"changed_by"`
	ChangedByUser      *UserInfo                `json:"changed_by_user,omitempty"`
	Notes              string                   `json:"notes,omitempty"`
	CreatedAt          time.Time                `json:"created_at"`
}

type StudentInfo struct {
	ID          string `json:"id"`
	StudentID   string `json:"student_id"`
	FullName    string `json:"full_name"`
	ProgramStudy string `json:"program_study"`
	AcademicYear string `json:"academic_year"`
}

type AchievementReferenceInfo struct {
	ID            string     `json:"id"`
	Status        model.AchievementStatus `json:"status"`
	SubmittedAt   *time.Time `json:"submitted_at,omitempty"`
	VerifiedAt    *time.Time `json:"verified_at,omitempty"`
	VerifiedBy    *string    `json:"verified_by,omitempty"`
	RejectionNote string    `json:"rejection_note,omitempty"`
}

type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

// Helper functions
func (s *achievementService) isStudent(ctx context.Context, userID uuid.UUID) (bool, *model.Student, error) {
	student, err := s.studentRepo.FindStudentByUserID(ctx, userID)
	if err != nil {
		return false, nil, nil
	}
	return true, student, nil
}

func (s *achievementService) isLecturer(ctx context.Context, userID uuid.UUID) (bool, *model.Lecturer, error) {
	lecturer, err := s.lecturerRepo.FindLecturerByUserID(ctx, userID)
	if err != nil {
		return false, nil, nil
	}
	return true, lecturer, nil
}

func (s *achievementService) checkRole(ctx context.Context, userID uuid.UUID) (string, error) {
	user, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return "", errors.New("user tidak ditemukan")
	}

	if user.RoleID == nil {
		return "", errors.New("user tidak memiliki role")
	}

	role, err := s.roleRepo.FindRoleByID(ctx, *user.RoleID)
	if err != nil {
		return "", errors.New("role tidak ditemukan")
	}

	roleNameLower := strings.ToLower(role.Name)
	if strings.Contains(roleNameLower, "mahasiswa") || strings.Contains(roleNameLower, "student") {
		return "student", nil
	}
	if strings.Contains(roleNameLower, "dosen") || strings.Contains(roleNameLower, "lecturer") {
		return "lecturer", nil
	}
	if strings.Contains(roleNameLower, "admin") {
		return "admin", nil
	}

	return "", errors.New("role tidak dikenali")
}

// CreateAchievement (FR-003)
func (s *achievementService) CreateAchievement(ctx context.Context, userID uuid.UUID, req *CreateAchievementRequest) (*AchievementResponse, error) {
	// Validasi user adalah mahasiswa
	isStudent, student, err := s.isStudent(ctx, userID)
	if !isStudent || err != nil {
		return nil, errors.New("hanya mahasiswa yang dapat membuat prestasi")
	}

	// Validasi required fields
	if req.Title == "" || req.Description == "" {
		return nil, errors.New("title dan description harus diisi")
	}

	// Validasi achievement type
	if req.AchievementType == "" {
		return nil, errors.New("achievement_type harus diisi")
	}

	// Validasi achievement type value
	validTypes := []model.AchievementType{
		model.AchievementTypeAcademic,
		model.AchievementTypeCompetition,
		model.AchievementTypeOrganization,
		model.AchievementTypePublication,
		model.AchievementTypeCertification,
		model.AchievementTypeOther,
	}
	validType := false
	for _, t := range validTypes {
		if req.AchievementType == t {
			validType = true
			break
		}
	}
	if !validType {
		return nil, errors.New("achievement_type tidak valid. Pilih: academic, competition, organization, publication, certification, other")
	}

	// Set default points jika tidak diisi
	points := req.Points
	if points == 0 {
		points = 0 // Default 0, bisa dihitung berdasarkan type nanti
	}

	// Create achievement di MongoDB
	achievement := &model.Achievement{
		StudentID:       student.ID.String(),
		AchievementType: req.AchievementType,
		Title:           req.Title,
		Description:     req.Description,
		Details:         req.Details,
		Attachments:     []model.Attachment{},
		Tags:            req.Tags,
		Points:          points,
		Status:          model.StatusDraft,
	}

	createdAchievement, err := s.achievementRepo.CreateAchievement(ctx, achievement)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan achievement: %v", err)
	}

	// Create reference di PostgreSQL
	reference := &model.AchievementReference{
		StudentID:          student.ID,
		MongoAchievementID: createdAchievement.ID.Hex(),
		Status:             model.StatusDraft,
	}

	if err := s.achievementRepo.CreateReference(ctx, reference); err != nil {
		// Rollback: delete dari MongoDB jika reference gagal
		s.achievementRepo.SoftDeleteAchievement(ctx, createdAchievement.ID.Hex())
		return nil, fmt.Errorf("gagal menyimpan reference: %v", err)
	}

	// Create initial history
	history := &model.AchievementHistory{
		AchievementRefID:   reference.ID,
		MongoAchievementID: createdAchievement.ID.Hex(),
		OldStatus:          nil, // Null untuk status awal (belum ada status sebelumnya)
		NewStatus:          model.StatusDraft,
		ChangedBy:          userID,
		Notes:              "Achievement dibuat",
	}

	if err := s.historyRepo.CreateHistory(ctx, history); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: Gagal membuat history: %v\n", err)
	}

	result := s.mapToAchievementResponse(ctx, createdAchievement, reference, student)
	return result, nil
}

// UpdateAchievement
func (s *achievementService) UpdateAchievement(ctx context.Context, userID uuid.UUID, achievementID string, req *UpdateAchievementRequest) (*AchievementResponse, error) {
	// Validasi user adalah mahasiswa
	isStudent, student, err := s.isStudent(ctx, userID)
	if !isStudent || err != nil {
		return nil, errors.New("hanya mahasiswa yang dapat mengupdate prestasi")
	}

	// Get achievement
	achievement, err := s.achievementRepo.FindAchievementByID(ctx, achievementID)
	if err != nil {
		return nil, errors.New("achievement tidak ditemukan")
	}

	// Validasi ownership
	if achievement.StudentID != student.ID.String() {
		return nil, errors.New("anda tidak memiliki akses untuk mengupdate achievement ini")
	}

	// Validasi status adalah draft
	if achievement.Status != model.StatusDraft {
		return nil, errors.New("hanya achievement dengan status draft yang dapat diupdate")
	}

	// Update achievement fields (hanya update field yang diisi)
	if req.Title != "" {
		achievement.Title = req.Title
	}
	if req.Description != "" {
		achievement.Description = req.Description
	}
	if req.AchievementType != "" {
		// Validasi achievement type
		validTypes := []model.AchievementType{
			model.AchievementTypeAcademic,
			model.AchievementTypeCompetition,
			model.AchievementTypeOrganization,
			model.AchievementTypePublication,
			model.AchievementTypeCertification,
			model.AchievementTypeOther,
		}
		validType := false
		for _, t := range validTypes {
			if req.AchievementType == t {
				validType = true
				break
			}
		}
		if !validType {
			return nil, errors.New("achievement_type tidak valid")
		}
		achievement.AchievementType = req.AchievementType
	}
	if req.Details.CustomFields != nil || req.Details.CompetitionName != nil {
		// Merge details (update field yang ada, keep yang tidak diupdate)
		if req.Details.CompetitionName != nil {
			achievement.Details.CompetitionName = req.Details.CompetitionName
		}
		if req.Details.CompetitionLevel != nil {
			achievement.Details.CompetitionLevel = req.Details.CompetitionLevel
		}
		if req.Details.Rank != nil {
			achievement.Details.Rank = req.Details.Rank
		}
		if req.Details.MedalType != nil {
			achievement.Details.MedalType = req.Details.MedalType
		}
		if req.Details.PublicationType != nil {
			achievement.Details.PublicationType = req.Details.PublicationType
		}
		if req.Details.PublicationTitle != nil {
			achievement.Details.PublicationTitle = req.Details.PublicationTitle
		}
		if req.Details.Authors != nil {
			achievement.Details.Authors = req.Details.Authors
		}
		if req.Details.Publisher != nil {
			achievement.Details.Publisher = req.Details.Publisher
		}
		if req.Details.ISSN != nil {
			achievement.Details.ISSN = req.Details.ISSN
		}
		if req.Details.OrganizationName != nil {
			achievement.Details.OrganizationName = req.Details.OrganizationName
		}
		if req.Details.Position != nil {
			achievement.Details.Position = req.Details.Position
		}
		if req.Details.Period != nil {
			achievement.Details.Period = req.Details.Period
		}
		if req.Details.CertificationName != nil {
			achievement.Details.CertificationName = req.Details.CertificationName
		}
		if req.Details.IssuedBy != nil {
			achievement.Details.IssuedBy = req.Details.IssuedBy
		}
		if req.Details.CertificationNumber != nil {
			achievement.Details.CertificationNumber = req.Details.CertificationNumber
		}
		if req.Details.ValidUntil != nil {
			achievement.Details.ValidUntil = req.Details.ValidUntil
		}
		if req.Details.EventDate != nil {
			achievement.Details.EventDate = req.Details.EventDate
		}
		if req.Details.Location != nil {
			achievement.Details.Location = req.Details.Location
		}
		if req.Details.Organizer != nil {
			achievement.Details.Organizer = req.Details.Organizer
		}
		if req.Details.Score != nil {
			achievement.Details.Score = req.Details.Score
		}
		if req.Details.CustomFields != nil {
			if achievement.Details.CustomFields == nil {
				achievement.Details.CustomFields = make(map[string]interface{})
			}
			for k, v := range req.Details.CustomFields {
				achievement.Details.CustomFields[k] = v
			}
		}
	} else if req.Details.CustomFields == nil {
		// Jika details diisi tapi kosong, replace seluruh details
		achievement.Details = *req.Details
	}
	if req.Tags != nil {
		achievement.Tags = req.Tags
	}
	if req.Points != nil {
		achievement.Points = *req.Points
	}

	if err := s.achievementRepo.UpdateAchievement(ctx, achievementID, achievement); err != nil {
		return nil, fmt.Errorf("gagal mengupdate achievement: %v", err)
	}

	// Get updated achievement
	updatedAchievement, err := s.achievementRepo.FindAchievementByID(ctx, achievementID)
	if err != nil {
		return nil, errors.New("gagal memuat achievement setelah update")
	}

	// Get reference
	reference, err := s.achievementRepo.FindReferenceByMongoID(ctx, achievementID)
	if err != nil {
		return nil, errors.New("gagal memuat reference")
	}

	result := s.mapToAchievementResponse(ctx, updatedAchievement, reference, student)
	return result, nil
}

// DeleteAchievement (FR-005)
func (s *achievementService) DeleteAchievement(ctx context.Context, userID uuid.UUID, achievementID string) error {
	// Validasi user adalah mahasiswa
	isStudent, student, err := s.isStudent(ctx, userID)
	if !isStudent || err != nil {
		return errors.New("hanya mahasiswa yang dapat menghapus prestasi")
	}

	// Get achievement
	achievement, err := s.achievementRepo.FindAchievementByID(ctx, achievementID)
	if err != nil {
		return errors.New("achievement tidak ditemukan")
	}

	// Validasi ownership
	if achievement.StudentID != student.ID.String() {
		return errors.New("anda tidak memiliki akses untuk menghapus achievement ini")
	}

	// Validasi status adalah draft
	if achievement.Status != model.StatusDraft {
		return errors.New("hanya achievement dengan status draft yang dapat dihapus")
	}

	// Soft delete di MongoDB
	if err := s.achievementRepo.SoftDeleteAchievement(ctx, achievementID); err != nil {
		return fmt.Errorf("gagal menghapus achievement: %v", err)
	}

	// Delete reference di PostgreSQL
	reference, err := s.achievementRepo.FindReferenceByMongoID(ctx, achievementID)
	if err == nil && reference != nil {
		if err := s.achievementRepo.DeleteReference(ctx, reference.ID); err != nil {
			// Log error but don't fail
			fmt.Printf("Warning: Gagal menghapus reference: %v\n", err)
		}
	}

	return nil
}

// SubmitAchievement (FR-004)
func (s *achievementService) SubmitAchievement(ctx context.Context, userID uuid.UUID, achievementID string) (*AchievementResponse, error) {
	// Validasi user adalah mahasiswa
	isStudent, student, err := s.isStudent(ctx, userID)
	if !isStudent || err != nil {
		return nil, errors.New("hanya mahasiswa yang dapat submit prestasi")
	}

	// Get achievement
	achievement, err := s.achievementRepo.FindAchievementByID(ctx, achievementID)
	if err != nil {
		return nil, errors.New("achievement tidak ditemukan")
	}

	// Validasi ownership
	if achievement.StudentID != student.ID.String() {
		return nil, errors.New("anda tidak memiliki akses untuk submit achievement ini")
	}

	// Validasi status adalah draft
	if achievement.Status != model.StatusDraft {
		return nil, errors.New("hanya achievement dengan status draft yang dapat disubmit")
	}

	// Update status ke submitted
	achievement.Status = model.StatusSubmitted
	if err := s.achievementRepo.UpdateAchievement(ctx, achievementID, achievement); err != nil {
		return nil, fmt.Errorf("gagal mengupdate status: %v", err)
	}

	// Update reference
	reference, err := s.achievementRepo.FindReferenceByMongoID(ctx, achievementID)
	if err != nil {
		return nil, errors.New("gagal memuat reference")
	}

	now := time.Now()
	reference.Status = model.StatusSubmitted
	reference.SubmittedAt = &now

	if err := s.achievementRepo.UpdateReference(ctx, reference); err != nil {
		return nil, fmt.Errorf("gagal mengupdate reference: %v", err)
	}

	// Create history
	oldStatus := model.StatusDraft
	history := &model.AchievementHistory{
		AchievementRefID:   reference.ID,
		MongoAchievementID: achievementID,
		OldStatus:          &oldStatus,
		NewStatus:          model.StatusSubmitted,
		ChangedBy:          userID,
		Notes:              "Achievement disubmit untuk verifikasi",
	}

	if err := s.historyRepo.CreateHistory(ctx, history); err != nil {
		fmt.Printf("Warning: Gagal membuat history: %v\n", err)
	}

	// Get updated achievement
	updatedAchievement, err := s.achievementRepo.FindAchievementByID(ctx, achievementID)
	if err != nil {
		return nil, errors.New("gagal memuat achievement setelah update")
	}

	result := s.mapToAchievementResponse(ctx, updatedAchievement, reference, student)
	return result, nil
}

// VerifyAchievement (FR-007)
func (s *achievementService) VerifyAchievement(ctx context.Context, userID uuid.UUID, achievementID string) (*AchievementResponse, error) {
	// Validasi user adalah dosen wali
	isLecturer, lecturer, err := s.isLecturer(ctx, userID)
	if !isLecturer || err != nil {
		return nil, errors.New("hanya dosen wali yang dapat memverifikasi prestasi")
	}

	// Get achievement
	achievement, err := s.achievementRepo.FindAchievementByID(ctx, achievementID)
	if err != nil {
		return nil, errors.New("achievement tidak ditemukan")
	}

	// Validasi status adalah submitted
	if achievement.Status != model.StatusSubmitted {
		return nil, errors.New("hanya achievement dengan status submitted yang dapat diverifikasi")
	}

	// Get student
	studentUUID, err := uuid.Parse(achievement.StudentID)
	if err != nil {
		return nil, errors.New("student ID tidak valid")
	}

		student, err := s.studentRepo.FindStudentByID(ctx, studentUUID)
	if err != nil {
		return nil, errors.New("student tidak ditemukan")
	}

	// Validasi dosen adalah advisor dari student
	if student.AdvisorID == nil || *student.AdvisorID != lecturer.ID {
		return nil, errors.New("anda bukan dosen wali dari mahasiswa ini")
	}

	// Update status ke verified
	achievement.Status = model.StatusVerified
	if err := s.achievementRepo.UpdateAchievement(ctx, achievementID, achievement); err != nil {
		return nil, fmt.Errorf("gagal mengupdate status: %v", err)
	}

	// Update reference
	reference, err := s.achievementRepo.FindReferenceByMongoID(ctx, achievementID)
	if err != nil {
		return nil, errors.New("gagal memuat reference")
	}

	now := time.Now()
	reference.Status = model.StatusVerified
	reference.VerifiedAt = &now
	reference.VerifiedBy = &userID

	if err := s.achievementRepo.UpdateReference(ctx, reference); err != nil {
		return nil, fmt.Errorf("gagal mengupdate reference: %v", err)
	}

	// Create history
	oldStatus := model.StatusSubmitted
	history := &model.AchievementHistory{
		AchievementRefID:   reference.ID,
		MongoAchievementID: achievementID,
		OldStatus:          &oldStatus,
		NewStatus:          model.StatusVerified,
		ChangedBy:          userID,
		Notes:              "Achievement diverifikasi",
	}

	if err := s.historyRepo.CreateHistory(ctx, history); err != nil {
		fmt.Printf("Warning: Gagal membuat history: %v\n", err)
	}

	// Get updated achievement
	updatedAchievement, err := s.achievementRepo.FindAchievementByID(ctx, achievementID)
	if err != nil {
		return nil, errors.New("gagal memuat achievement setelah update")
	}

	result := s.mapToAchievementResponse(ctx, updatedAchievement, reference, student)
	return result, nil
}

// RejectAchievement (FR-008)
func (s *achievementService) RejectAchievement(ctx context.Context, userID uuid.UUID, achievementID string, rejectionNote string) (*AchievementResponse, error) {
	// Validasi user adalah dosen wali
	isLecturer, lecturer, err := s.isLecturer(ctx, userID)
	if !isLecturer || err != nil {
		return nil, errors.New("hanya dosen wali yang dapat menolak prestasi")
	}

	// Validasi rejection note
	if rejectionNote == "" {
		return nil, errors.New("rejection note harus diisi")
	}

	// Get achievement
	achievement, err := s.achievementRepo.FindAchievementByID(ctx, achievementID)
	if err != nil {
		return nil, errors.New("achievement tidak ditemukan")
	}

	// Validasi status adalah submitted
	if achievement.Status != model.StatusSubmitted {
		return nil, errors.New("hanya achievement dengan status submitted yang dapat ditolak")
	}

	// Get student
	studentUUID, err := uuid.Parse(achievement.StudentID)
	if err != nil {
		return nil, errors.New("student ID tidak valid")
	}

		student, err := s.studentRepo.FindStudentByID(ctx, studentUUID)
	if err != nil {
		return nil, errors.New("student tidak ditemukan")
	}

	// Validasi dosen adalah advisor dari student
	if student.AdvisorID == nil || *student.AdvisorID != lecturer.ID {
		return nil, errors.New("anda bukan dosen wali dari mahasiswa ini")
	}

	// Update status ke rejected
	achievement.Status = model.StatusRejected
	if err := s.achievementRepo.UpdateAchievement(ctx, achievementID, achievement); err != nil {
		return nil, fmt.Errorf("gagal mengupdate status: %v", err)
	}

	// Update reference
	reference, err := s.achievementRepo.FindReferenceByMongoID(ctx, achievementID)
	if err != nil {
		return nil, errors.New("gagal memuat reference")
	}

	reference.Status = model.StatusRejected
	reference.RejectionNote = rejectionNote

	if err := s.achievementRepo.UpdateReference(ctx, reference); err != nil {
		return nil, fmt.Errorf("gagal mengupdate reference: %v", err)
	}

	// Create history
	oldStatus := model.StatusSubmitted
	history := &model.AchievementHistory{
		AchievementRefID:   reference.ID,
		MongoAchievementID: achievementID,
		OldStatus:          &oldStatus,
		NewStatus:          model.StatusRejected,
		ChangedBy:          userID,
		Notes:              fmt.Sprintf("Achievement ditolak: %s", rejectionNote),
	}

	if err := s.historyRepo.CreateHistory(ctx, history); err != nil {
		fmt.Printf("Warning: Gagal membuat history: %v\n", err)
	}

	// Get updated achievement
	updatedAchievement, err := s.achievementRepo.FindAchievementByID(ctx, achievementID)
	if err != nil {
		return nil, errors.New("gagal memuat achievement setelah update")
	}

	result := s.mapToAchievementResponse(ctx, updatedAchievement, reference, student)
	return result, nil
}

// GetAchievements (FR-006 untuk dosen, list untuk mahasiswa)
func (s *achievementService) GetAchievements(ctx context.Context, userID uuid.UUID, page, limit int, status string) (*AchievementListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	role, err := s.checkRole(ctx, userID)
	if err != nil {
		return nil, err
	}

	var studentIDs []uuid.UUID
	var achievements []model.Achievement
	var references []model.AchievementReference
	var total int64

	if role == "student" {
		// Mahasiswa melihat achievement sendiri
		isStudent, student, err := s.isStudent(ctx, userID)
		if !isStudent || err != nil {
			return nil, errors.New("user tidak ditemukan sebagai mahasiswa")
		}

		studentIDs = []uuid.UUID{student.ID}
		achievements, err = s.achievementRepo.FindAchievementsByStudentID(ctx, student.ID.String())
		if err != nil {
			return nil, fmt.Errorf("gagal memuat achievements: %v", err)
		}

		// Get references
		references, err = s.achievementRepo.FindReferencesByStudentIDs(ctx, studentIDs)
		if err != nil {
			return nil, fmt.Errorf("gagal memuat references: %v", err)
		}

		total = int64(len(achievements))
	} else if role == "lecturer" {
		// Dosen wali melihat achievement mahasiswa bimbingannya
		isLecturer, lecturer, err := s.isLecturer(ctx, userID)
		if !isLecturer || err != nil {
			return nil, errors.New("user tidak ditemukan sebagai dosen")
		}

		// Get advisees
		advisees, err := s.lecturerRepo.FindAdvisees(ctx, lecturer.ID)
		if err != nil {
			return nil, fmt.Errorf("gagal memuat mahasiswa bimbingan: %v", err)
		}

		if len(advisees) == 0 {
			return &AchievementListResponse{
				Data:       []AchievementResponse{},
				Page:       page,
				Limit:      limit,
				Total:      0,
				TotalPages: 0,
			}, nil
		}

		// Get student IDs
		for _, advisee := range advisees {
			studentIDs = append(studentIDs, advisee.ID)
		}

		// Get references dengan pagination
		references, total, err = s.achievementRepo.FindReferencesWithPagination(ctx, studentIDs, page, limit)
		if err != nil {
			return nil, fmt.Errorf("gagal memuat references: %v", err)
		}

		// Get achievements dari MongoDB
		for _, ref := range references {
			achievement, err := s.achievementRepo.FindAchievementByID(ctx, ref.MongoAchievementID)
			if err == nil {
				achievements = append(achievements, *achievement)
			}
		}
	} else {
		// Admin bisa lihat semua
		allStudents, err := s.studentRepo.FindAllStudents(ctx)
		if err != nil {
			return nil, fmt.Errorf("gagal memuat students: %v", err)
		}

		for _, student := range allStudents {
			studentIDs = append(studentIDs, student.ID)
		}

		references, total, err = s.achievementRepo.FindReferencesWithPagination(ctx, studentIDs, page, limit)
		if err != nil {
			return nil, fmt.Errorf("gagal memuat references: %v", err)
		}

		for _, ref := range references {
			achievement, err := s.achievementRepo.FindAchievementByID(ctx, ref.MongoAchievementID)
			if err == nil {
				achievements = append(achievements, *achievement)
			}
		}
	}

	// Filter by status if provided
	if status != "" {
		filteredAchievements := []model.Achievement{}
		filteredReferences := []model.AchievementReference{}

		for i, achievement := range achievements {
			if string(achievement.Status) == status {
				filteredAchievements = append(filteredAchievements, achievement)
				if i < len(references) {
					filteredReferences = append(filteredReferences, references[i])
				}
			}
		}

		achievements = filteredAchievements
		references = filteredReferences
		total = int64(len(achievements))
	}

	// Map to response
	responseData := []AchievementResponse{}
	for i, achievement := range achievements {
		var ref *model.AchievementReference
		if i < len(references) {
			ref = &references[i]
		}

		studentUUID, _ := uuid.Parse(achievement.StudentID)
		student, _ := s.studentRepo.FindStudentByID(ctx, studentUUID)

		responseData = append(responseData, *s.mapToAchievementResponse(ctx, &achievement, ref, student))
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &AchievementListResponse{
		Data:       responseData,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// GetAchievementByID
func (s *achievementService) GetAchievementByID(ctx context.Context, userID uuid.UUID, achievementID string) (*AchievementResponse, error) {
	// Get achievement
	achievement, err := s.achievementRepo.FindAchievementByID(ctx, achievementID)
	if err != nil {
		return nil, errors.New("achievement tidak ditemukan")
	}

	// Get reference
	reference, err := s.achievementRepo.FindReferenceByMongoID(ctx, achievementID)
	if err != nil {
		return nil, errors.New("gagal memuat reference")
	}

	// Check access
	role, err := s.checkRole(ctx, userID)
	if err != nil {
		return nil, err
	}

	if role == "student" {
		// Mahasiswa hanya bisa lihat achievement sendiri
		isStudent, student, err := s.isStudent(ctx, userID)
		if !isStudent || err != nil {
			return nil, errors.New("user tidak ditemukan sebagai mahasiswa")
		}

		if achievement.StudentID != student.ID.String() {
			return nil, errors.New("anda tidak memiliki akses untuk melihat achievement ini")
		}

		result := s.mapToAchievementResponse(ctx, achievement, reference, student)
		return result, nil
	} else if role == "lecturer" {
		// Dosen wali hanya bisa lihat achievement mahasiswa bimbingannya
		isLecturer, lecturer, err := s.isLecturer(ctx, userID)
		if !isLecturer || err != nil {
			return nil, errors.New("user tidak ditemukan sebagai dosen")
		}

		studentUUID, err := uuid.Parse(achievement.StudentID)
		if err != nil {
			return nil, errors.New("student ID tidak valid")
		}

		student, err := s.studentRepo.FindStudentByID(ctx, studentUUID)
		if err != nil {
			return nil, errors.New("student tidak ditemukan")
		}

		if student.AdvisorID == nil || *student.AdvisorID != lecturer.ID {
			return nil, errors.New("anda bukan dosen wali dari mahasiswa ini")
		}

		result := s.mapToAchievementResponse(ctx, achievement, reference, student)
		return result, nil
	}

	// Admin bisa lihat semua
	studentUUID, _ := uuid.Parse(achievement.StudentID)
	student, _ := s.studentRepo.FindStudentByID(ctx, studentUUID)

	result := s.mapToAchievementResponse(ctx, achievement, reference, student)
	return result, nil
}

// GetAchievementHistory
func (s *achievementService) GetAchievementHistory(ctx context.Context, userID uuid.UUID, achievementID string) ([]AchievementHistoryResponse, error) {
	// Get reference
	reference, err := s.achievementRepo.FindReferenceByMongoID(ctx, achievementID)
	if err != nil {
		return nil, errors.New("achievement tidak ditemukan")
	}

	// Check access (same as GetAchievementByID)
	role, err := s.checkRole(ctx, userID)
	if err != nil {
		return nil, err
	}

	if role == "student" {
		isStudent, student, err := s.isStudent(ctx, userID)
		if !isStudent || err != nil {
			return nil, errors.New("user tidak ditemukan sebagai mahasiswa")
		}

		achievement, err := s.achievementRepo.FindAchievementByID(ctx, achievementID)
		if err != nil {
			return nil, errors.New("achievement tidak ditemukan")
		}

		if achievement.StudentID != student.ID.String() {
			return nil, errors.New("anda tidak memiliki akses untuk melihat history achievement ini")
		}
	} else if role == "lecturer" {
		isLecturer, lecturer, err := s.isLecturer(ctx, userID)
		if !isLecturer || err != nil {
			return nil, errors.New("user tidak ditemukan sebagai dosen")
		}

		achievement, err := s.achievementRepo.FindAchievementByID(ctx, achievementID)
		if err != nil {
			return nil, errors.New("achievement tidak ditemukan")
		}

		studentUUID, err := uuid.Parse(achievement.StudentID)
		if err != nil {
			return nil, errors.New("student ID tidak valid")
		}

		student, err := s.studentRepo.FindStudentByID(ctx, studentUUID)
		if err != nil {
			return nil, errors.New("student tidak ditemukan")
		}

		if student.AdvisorID == nil || *student.AdvisorID != lecturer.ID {
			return nil, errors.New("anda bukan dosen wali dari mahasiswa ini")
		}
	}

	// Get history
	histories, err := s.historyRepo.FindHistoriesByAchievementRefID(ctx, reference.ID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat history: %v", err)
	}

	response := []AchievementHistoryResponse{}
	for _, history := range histories {
		var changedByUser *UserInfo
		if history.ChangedByUser.ID != uuid.Nil {
			changedByUser = &UserInfo{
				ID:       history.ChangedByUser.ID.String(),
				Username: history.ChangedByUser.Username,
				FullName: history.ChangedByUser.FullName,
				Email:    history.ChangedByUser.Email,
			}
		}

		response = append(response, AchievementHistoryResponse{
			ID:            history.ID.String(),
			OldStatus:     history.OldStatus,
			NewStatus:     history.NewStatus,
			ChangedBy:     history.ChangedBy.String(),
			ChangedByUser: changedByUser,
			Notes:         history.Notes,
			CreatedAt:     history.CreatedAt,
		})
	}

	return response, nil
}

// UploadAttachment
func (s *achievementService) UploadAttachment(ctx context.Context, userID uuid.UUID, achievementID string, filePath string) (string, error) {
	// Validasi user adalah mahasiswa
	isStudent, student, err := s.isStudent(ctx, userID)
	if !isStudent || err != nil {
		return "", errors.New("hanya mahasiswa yang dapat upload attachment")
	}

	// Get achievement
	achievement, err := s.achievementRepo.FindAchievementByID(ctx, achievementID)
	if err != nil {
		return "", errors.New("achievement tidak ditemukan")
	}

	// Validasi ownership
	if achievement.StudentID != student.ID.String() {
		return "", errors.New("anda tidak memiliki akses untuk upload attachment ke achievement ini")
	}

	// Validasi status adalah draft
	if achievement.Status != model.StatusDraft {
		return "", errors.New("hanya achievement dengan status draft yang dapat diupdate attachment")
	}

	// Create attachment object
	attachment := model.Attachment{
		FileName:   filePath,
		FileURL:    filePath, // In production, this should be full URL
		FileType:   getFileTypeFromPath(filePath),
		UploadedAt: time.Now(),
	}

	// Update achievement attachments
	achievement.Attachments = append(achievement.Attachments, attachment)
	if err := s.achievementRepo.UpdateAchievement(ctx, achievementID, achievement); err != nil {
		return "", fmt.Errorf("gagal mengupdate achievement: %v", err)
	}

	return filePath, nil
}

// Helper function to get file type from path
func getFileTypeFromPath(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	default:
		return "application/octet-stream"
	}
}

// Helper: mapToAchievementResponse
func (s *achievementService) mapToAchievementResponse(ctx context.Context, achievement *model.Achievement, reference *model.AchievementReference, student *model.Student) *AchievementResponse {
	var studentInfo *StudentInfo
	if student != nil {
		studentInfo = &StudentInfo{
			ID:           student.ID.String(),
			StudentID:    student.StudentID,
			FullName:     student.User.FullName,
			ProgramStudy: student.ProgramStudy,
			AcademicYear: student.AcademicYear,
		}
	}

	var refInfo *AchievementReferenceInfo
	if reference != nil {
		var verifiedBy *string
		if reference.VerifiedBy != nil {
			verifiedByStr := reference.VerifiedBy.String()
			verifiedBy = &verifiedByStr
		}

		refInfo = &AchievementReferenceInfo{
			ID:            reference.ID.String(),
			Status:        reference.Status,
			SubmittedAt:   reference.SubmittedAt,
			VerifiedAt:    reference.VerifiedAt,
			VerifiedBy:    verifiedBy,
			RejectionNote: reference.RejectionNote,
		}
	}

	return &AchievementResponse{
		ID:              achievement.ID.Hex(),
		StudentID:       achievement.StudentID,
		Student:         studentInfo,
		AchievementType: achievement.AchievementType,
		Title:           achievement.Title,
		Description:     achievement.Description,
		Details:         achievement.Details,
		Attachments:     achievement.Attachments,
		Tags:            achievement.Tags,
		Points:          achievement.Points,
		Status:          achievement.Status,
		Reference:       refInfo,
		CreatedAt:       achievement.CreatedAt,
		UpdatedAt:       achievement.UpdatedAt,
	}
}

