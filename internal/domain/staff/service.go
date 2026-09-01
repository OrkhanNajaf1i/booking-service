// File: internal/domain/staff/service.go
package staff

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type StaffService struct {
	repo        Repository
	userService UserService
	owners      BusinessOwnerLookup
}

func NewService(
	repo Repository,
	userService UserService,
	owners BusinessOwnerLookup,
) *StaffService {
	return &StaffService{
		repo:        repo,
		userService: userService,
		owners:      owners,
	}
}

// CreateStaffProfile - Yeni işçi profili (owner tərəfindən manual)
func (s *StaffService) CreateStaffProfile(
	ctx context.Context,
	businessID uuid.UUID,
	req *CreateStaffRequest,
) (*StaffProfile, error) {
	if businessID == uuid.Nil {
		return nil, &StaffError{Code: "INVALID_BUSINESS", Message: "Business ID cannot be empty"}
	}
	if req == nil {
		return nil, &StaffError{Code: "INVALID_REQUEST", Message: "Request cannot be nil"}
	}

	// Devet yalniz komanda rejiminde.
	if s.owners != nil {
		team, err := s.owners.IsTeamMode(ctx, businessID)
		if err != nil {
			return nil, fmt.Errorf("failed to read business mode: %w", err)
		}
		if !team {
			return nil, &StaffError{
				Code:    "TEAM_MODE_REQUIRED",
				Message: "Bu hesab tek isci kimi qurulub. Evvelce komanda rejimine kecin.",
			}
		}
	}

	// Validation
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	profile := NewStaffProfile(req.UserID, businessID, req.Role, req.Title)
	profile.LocationID = req.LocationID

	if err := s.validateStaffProfile(profile); err != nil {
		return nil, err
	}

	if err := s.repo.CreateStaffProfile(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to create staff profile: %w", err)
	}

	return profile, nil
}

// GetStaff - Konkret işçi tapmaq
func (s *StaffService) GetStaff(
	ctx context.Context,
	staffID, businessID uuid.UUID,
) (*StaffProfile, error) {
	if staffID == uuid.Nil || businessID == uuid.Nil {
		return nil, &StaffError{Code: "INVALID_ID", Message: "Staff ID and Business ID are required"}
	}

	staff, err := s.repo.GetStaffByID(ctx, staffID, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to get staff: %w", err)
	}
	if staff == nil {
		return nil, &StaffError{Code: "NOT_FOUND", Message: "Staff not found"}
	}

	return staff, nil
}

// ListStaff - Business-ə aid bütün işçilər (JOIN ilə User detalları)
func (s *StaffService) ListStaff(
	ctx context.Context,
	businessID uuid.UUID,
) ([]*StaffWithUser, error) {
	if businessID == uuid.Nil {
		return nil, &StaffError{Code: "INVALID_BUSINESS", Message: "Business ID cannot be empty"}
	}

	staff, err := s.repo.ListByBusiness(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to list staff: %w", err)
	}

	// Sahib bir defe tapilir, her setir ucun ayrica sorgu getmir.
	// Axtaris alinmasa siyahi yene qaytarilir — is_owner sadece false
	// qalir ve server terefindeki qadaga onsuz da yerindedir.
	if s.owners != nil {
		if ownerID, ownerErr := s.owners.OwnerUserID(ctx, businessID); ownerErr == nil &&
			ownerID != uuid.Nil {
			for _, member := range staff {
				member.IsOwner = member.UserID == ownerID
			}
		}
	}

	return staff, nil
}

// UpdateStaff - İşçi məlumatlarını yeniləmək
func (s *StaffService) UpdateStaff(
	ctx context.Context,
	staffID, businessID uuid.UUID,
	req *UpdateStaffRequest,
) error {
	if staffID == uuid.Nil || businessID == uuid.Nil {
		return &StaffError{Code: "INVALID_ID", Message: "Staff ID and Business ID are required"}
	}

	staff, err := s.repo.GetStaffByID(ctx, staffID, businessID)
	if err != nil {
		return fmt.Errorf("failed to get staff: %w", err)
	}
	if staff == nil {
		return &StaffError{Code: "NOT_FOUND", Message: "Staff not found"}
	}

	// Update fields
	staff.Role = req.Role
	staff.Title = req.Title
	staff.Department = req.Department
	staff.Bio = req.Bio
	staff.HourlyRate = req.HourlyRate
	staff.LocationID = req.LocationID
	staff.UpdatedAt = time.Now()

	if err := s.validateStaffProfile(staff); err != nil {
		return err
	}

	if err := s.repo.UpdateStaffProfile(ctx, staff); err != nil {
		return fmt.Errorf("failed to update staff: %w", err)
	}

	return nil
}

// DeactivateStaff - Soft delete (status = inactive)
func (s *StaffService) DeactivateStaff(
	ctx context.Context,
	staffID, businessID uuid.UUID,
) error {
	if staffID == uuid.Nil || businessID == uuid.Nil {
		return &StaffError{Code: "INVALID_ID", Message: "Staff ID and Business ID are required"}
	}

	// Sahib ozunu silmemelidir.
	//
	// Tek isleyen hekim/berber ucun bu, yegane iscini silmek demekdir:
	// biznes kesf ekranindan itir, movcud randevular sahibsiz qalir.
	// Komandali biznesde de sahib eyni zamanda adminlerdir — o gedende
	// biznesi idare edecek adam qalmir. Isci cixarmaq lazimdirsa
	// DIGER iscilər silinir.
	owner, err := s.isOwner(ctx, staffID, businessID)
	if err != nil {
		return err
	}
	if owner {
		return &StaffError{
			Code:    "OWNER_CANNOT_BE_REMOVED",
			Message: "Biznes sahibi işçi siyahısından silinə bilməz",
		}
	}

	if err := s.repo.DeactivateStaff(ctx, staffID, businessID); err != nil {
		return fmt.Errorf("failed to deactivate staff: %w", err)
	}

	return nil
}

// IsOwner – hemin isci biznesin sahibidirmi. Panel duymeni bunun uzre
// gizledir; qerar yene serverde verilir.
func (s *StaffService) IsOwner(
	ctx context.Context,
	staffID, businessID uuid.UUID,
) (bool, error) {
	if staffID == uuid.Nil || businessID == uuid.Nil {
		return false, &StaffError{Code: "INVALID_ID", Message: "Staff ID and Business ID are required"}
	}
	return s.isOwner(ctx, staffID, businessID)
}

func (s *StaffService) isOwner(
	ctx context.Context,
	staffID, businessID uuid.UUID,
) (bool, error) {
	// Sahib axtarisi qurulmayibsa qadaga tetbiq edile bilmez.
	// Sessizce kecmek daha pisdir — sahib silinerdi.
	if s.owners == nil {
		return false, &StaffError{
			Code:    "OWNER_LOOKUP_UNAVAILABLE",
			Message: "Sahib yoxlanışı mümkün olmadı",
		}
	}

	profile, err := s.repo.GetStaffByID(ctx, staffID, businessID)
	if err != nil {
		return false, fmt.Errorf("failed to get staff: %w", err)
	}
	if profile == nil {
		return false, &StaffError{Code: "NOT_FOUND", Message: "İşçi tapılmadı"}
	}

	ownerID, err := s.owners.OwnerUserID(ctx, businessID)
	if err != nil {
		return false, fmt.Errorf("failed to get business owner: %w", err)
	}

	return ownerID != uuid.Nil && ownerID == profile.UserID, nil
}

// InviteStaff - Yeni işçi dəvət etmək
func (s *StaffService) InviteStaff(
	ctx context.Context,
	businessID uuid.UUID,
	req *InviteStaffRequest,
) (string, error) {
	if businessID == uuid.Nil {
		return "", &StaffError{Code: "INVALID_BUSINESS", Message: "Business ID cannot be empty"}
	}
	if req == nil {
		return "", &StaffError{Code: "INVALID_REQUEST", Message: "Request cannot be nil"}
	}

	// Devet yalniz komanda rejiminde.
	if s.owners != nil {
		team, err := s.owners.IsTeamMode(ctx, businessID)
		if err != nil {
			return "", fmt.Errorf("failed to read business mode: %w", err)
		}
		if !team {
			return "", &StaffError{
				Code:    "TEAM_MODE_REQUIRED",
				Message: "Bu hesab tek isci kimi qurulub. Evvelce komanda rejimine kecin.",
			}
		}
	}

	// Validation
	if err := s.validateInviteRequest(req); err != nil {
		return "", err
	}

	// Generate secure token (32 bytes)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	invite := &BusinessInvite{
		ID:           uuid.New(),
		BusinessID:   businessID,
		InvitedEmail: req.Email,
		InvitedPhone: req.Phone,
		Role:         req.Role,
		LocationID:   req.LocationID,
		Token:        token,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
		Used:         false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.CreateInvite(ctx, invite); err != nil {
		return "", fmt.Errorf("failed to create invite: %w", err)
	}

	return token, nil
}

func (s *StaffService) ValidateInviteToken(
	ctx context.Context,
	token string,
) (*BusinessInvite, error) {
	if token == "" {
		return nil, &StaffError{Code: "INVALID_TOKEN", Message: "Invite token is required"}
	}

	invite, err := s.repo.GetInviteByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get invite: %w", err)
	}
	if invite == nil {
		return nil, &StaffError{Code: "INVALID_TOKEN", Message: "Invalid invite token"}
	}

	if invite.Used {
		return nil, &StaffError{Code: "TOKEN_USED", Message: "Invite token has already been used"}
	}

	if time.Now().After(invite.ExpiresAt) {
		return nil, &StaffError{Code: "TOKEN_EXPIRED", Message: "Invite token has expired"}
	}

	return invite, nil
}

func (s *StaffService) AcceptInvite(
	ctx context.Context,
	userID uuid.UUID,
	token, password string,
) error {
	if userID == uuid.Nil {
		return &StaffError{Code: "INVALID_USER", Message: "User ID cannot be empty"}
	}
	if token == "" {
		return &StaffError{Code: "INVALID_TOKEN", Message: "Invite token is required"}
	}

	invite, err := s.ValidateInviteToken(ctx, token)
	if err != nil {
		return err
	}

	if err := s.userService.UpdateUserBusinessID(ctx, userID, invite.BusinessID, false); err != nil {
		return fmt.Errorf("failed to link user to business: %w", err)
	}

	profile := NewStaffProfile(userID, invite.BusinessID, invite.Role, "Staff Member")
	profile.LocationID = invite.LocationID

	if err := s.repo.CreateStaffProfile(ctx, profile); err != nil {
		return fmt.Errorf("failed to create staff profile: %w", err)
	}

	if err := s.repo.MarkInviteAsUsed(ctx, invite.ID); err != nil {
		return fmt.Errorf("failed to mark invite as used: %w", err)
	}

	return nil
}
