package stores

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/datarohit/gopher-social-backend/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileStore struct {
	dbPool *pgxpool.Pool
}

// NewProfileStore creates a new ProfileStore.
//
// Parameters:
//   - dbPool (*pgxpool.Pool): Pgx connection pool.
//
// Returns:
//   - *ProfileStore: ProfileStore instance.
func NewProfileStore(dbPool *pgxpool.Pool) *ProfileStore {
	return &ProfileStore{
		dbPool: dbPool,
	}
}

// ErrProfileNotFound is returned when a profile is not found.
var ErrProfileNotFound = errors.New("profile not found")

// UpdateProfile updates an existing user profile in the database or creates a new one if it doesn't exist.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - profile (*models.Profile): Profile object containing updated profile information.
//
// Returns:
//   - *models.Profile: The updated or created profile if successful.
//   - error: Errors during database query.
func (ps *ProfileStore) UpdateProfile(ctx context.Context, profile *models.Profile) (*models.Profile, error) {
	var updatedProfile models.Profile
	err := ps.dbPool.QueryRow(ctx, `
		INSERT INTO profiles (
			user_id,
			first_name,
			last_name,
			website,
			github,
			linkedin,
			twitter,
			google_scholar,
			created_at,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()
		) ON CONFLICT (user_id) DO UPDATE SET
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			website = EXCLUDED.website,
			github = EXCLUDED.github,
			linkedin = EXCLUDED.linkedin,
			twitter = EXCLUDED.twitter,
			google_scholar = EXCLUDED.google_scholar,
			updated_at = NOW()
		RETURNING id, user_id, first_name, last_name, website, github, linkedin, twitter, google_scholar, created_at, updated_at
	`, profile.UserID, profile.FirstName, profile.LastName, profile.Website, profile.Github, profile.LinkedIn, profile.Twitter, profile.GoogleScholar).Scan(
		&updatedProfile.ID, &updatedProfile.UserID, &updatedProfile.FirstName, &updatedProfile.LastName, &updatedProfile.Website, &updatedProfile.Github, &updatedProfile.LinkedIn, &updatedProfile.Twitter, &updatedProfile.GoogleScholar, &updatedProfile.CreatedAt, &updatedProfile.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update or create profile: %w", err)
	}

	return &updatedProfile, nil
}

// CreateProfile creates a new user profile in the database.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - profile (*models.Profile): Profile object containing profile information.
//
// Returns:
//   - *models.Profile: The created profile if successful.
//   - error: ErrProfileNotFound if profile not found or other errors during database query.
func (ps *ProfileStore) CreateProfile(ctx context.Context, profile *models.Profile) (*models.Profile, error) {
	var createdProfile models.Profile
	profile.ID = uuid.New()
	err := ps.dbPool.QueryRow(ctx, `
		INSERT INTO profiles (
			id,
			user_id,
			first_name,
			last_name,
			website,
			github,
			linkedin,
			twitter,
			google_scholar,
			created_at,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()
		)
		RETURNING id, user_id, first_name, last_name, website, github, linkedin, twitter, google_scholar, created_at, updated_at
	`, profile.ID, profile.UserID, profile.FirstName, profile.LastName, profile.Website, profile.Github, profile.LinkedIn, profile.Twitter, profile.GoogleScholar).Scan(
		&createdProfile.ID, &createdProfile.UserID, &createdProfile.FirstName, &createdProfile.LastName, &createdProfile.Website, &createdProfile.Github, &createdProfile.LinkedIn, &createdProfile.Twitter, &createdProfile.GoogleScholar, &createdProfile.CreatedAt, &createdProfile.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	return &createdProfile, nil
}

// GetProfileByUserID retrieves a user profile by user ID from the database and joins with user data.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - userID (uuid.UUID): The ID of the user whose profile is to be retrieved.
//
// Returns:
//   - *models.Profile: The retrieved profile with associated user data if found.
//   - error: ErrProfileNotFound if profile not found or other errors during database query.
func (ps *ProfileStore) GetProfileByUserID(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	var profile models.Profile
	profile.User = &models.User{}
	profile.User.Role = &models.Role{}

	err := ps.dbPool.QueryRow(ctx, `
		SELECT
			p.id, p.user_id, p.first_name, p.last_name, p.website, p.github, p.linkedin, p.twitter, p.google_scholar, p.created_at, p.updated_at,
			u.id, u.username, u.email, u.timeout_until, u.banned, u.is_active, u.created_at, u.updated_at,
			r.level, r.description,
			(SELECT COUNT(*) FROM follows WHERE followee_id = u.id) as followers_count,
			(SELECT COUNT(*) FROM follows WHERE follower_id = u.id) as following_count
		FROM profiles p
		INNER JOIN users u ON p.user_id = u.id
		INNER JOIN roles r ON u.role_id = r.id
		WHERE p.user_id = $1
	`, userID).Scan(
		&profile.ID, &profile.UserID, &profile.FirstName, &profile.LastName, &profile.Website, &profile.Github, &profile.LinkedIn, &profile.Twitter, &profile.GoogleScholar, &profile.CreatedAt, &profile.UpdatedAt,
		&profile.User.ID, &profile.User.Username, &profile.User.Email, &profile.User.TimeoutUntil, &profile.User.Banned, &profile.User.IsActive, &profile.User.CreatedAt, &profile.User.UpdatedAt,
		&profile.User.Role.Level, &profile.User.Role.Description,
		&profile.User.Followers, &profile.User.Following,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProfileNotFound
		}
		return nil, fmt.Errorf("failed to get profile by user ID: %w", err)
	}

	return &profile, nil
}

// GetProfileForLoggedInUser retrieves a user profile for the logged-in user by user ID from the database.
// It's similar to GetProfileByUserID but specifically named to reflect its use case.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - userID (uuid.UUID): The ID of the logged-in user whose profile is to be retrieved.
//
// Returns:
//   - *models.Profile: The retrieved profile if found.
//   - error: ErrProfileNotFound if profile not found or other errors during database query.
func (ps *ProfileStore) GetProfileForLoggedInUser(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	profile, err := ps.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrProfileNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, fmt.Errorf("failed to get profile for logged-in user: %w", err)
	}
	return profile, nil
}

// GetProfileByIdentifier retrieves a user profile by username, email, or user ID.
// It attempts to identify the identifier type and queries the database accordingly.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - identifier (string): Username, email, or user ID of the user profile to retrieve.
//
// Returns:
//   - *models.Profile: The retrieved profile if found.
//   - error: ErrProfileNotFound if profile not found or other errors during database query.
func (ps *ProfileStore) GetProfileByIdentifier(ctx context.Context, identifier string) (*models.Profile, error) {
	var profile *models.Profile
	var err error

	userID, uuidErr := uuid.Parse(identifier)
	if uuidErr == nil {
		profile, err = ps.GetProfileByUserID(ctx, userID)
		if err == nil {
			return profile, nil
		} else if !errors.Is(err, ErrProfileNotFound) {
			return nil, fmt.Errorf("failed to get profile by user ID: %w", err)
		}
	}

	profile, err = ps.getProfileByUsernameOrEmail(ctx, identifier)
	if err == nil {
		return profile, nil
	} else if !errors.Is(err, ErrProfileNotFound) {
		return nil, fmt.Errorf("failed to get profile by username or email: %w", err)
	}

	return nil, ErrProfileNotFound
}

// getProfileByUsernameOrEmail retrieves a user profile by either username or email.
// It queries the database to find a user matching the given username or email.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - identifier (string): Username or email to search for.
//
// Returns:
//   - *models.Profile: The retrieved profile if found.
//   - error: ErrProfileNotFound if profile not found or other errors during database query.
func (ps *ProfileStore) getProfileByUsernameOrEmail(ctx context.Context, identifier string) (*models.Profile, error) {
	var profile models.Profile
	profile.User = &models.User{}
	profile.User.Role = &models.Role{}

	var query string
	if strings.Contains(identifier, "@") {
		query = `
			SELECT
				p.id, p.user_id, p.first_name, p.last_name, p.website, p.github, p.linkedin, p.twitter, p.google_scholar, p.created_at, p.updated_at,
				u.id, u.username, u.email, u.timeout_until, u.banned, u.is_active, u.created_at, u.updated_at,
				r.level, r.description,
				(SELECT COUNT(*) FROM follows WHERE followee_id = u.id) as followers_count,
				(SELECT COUNT(*) FROM follows WHERE follower_id = u.id) as following_count
			FROM profiles p
			INNER JOIN users u ON p.user_id = u.id
			INNER JOIN roles r ON u.role_id = r.id
			WHERE u.email = $1
		`
	} else {
		query = `
			SELECT
				p.id, p.user_id, p.first_name, p.last_name, p.website, p.github, p.linkedin, p.twitter, p.google_scholar, p.created_at, p.updated_at,
				u.id, u.username, u.email, u.timeout_until, u.banned, u.is_active, u.created_at, u.updated_at,
				r.level, r.description,
				(SELECT COUNT(*) FROM follows WHERE followee_id = u.id) as followers_count,
				(SELECT COUNT(*) FROM follows WHERE follower_id = u.id) as following_count
			FROM profiles p
			INNER JOIN users u ON p.user_id = u.id
			INNER JOIN roles r ON u.role_id = r.id
			WHERE u.username = $1
		`
	}

	err := ps.dbPool.QueryRow(ctx, query, identifier).Scan(
		&profile.ID, &profile.UserID, &profile.FirstName, &profile.LastName, &profile.Website, &profile.Github, &profile.LinkedIn, &profile.Twitter, &profile.GoogleScholar, &profile.CreatedAt, &profile.UpdatedAt,
		&profile.User.ID, &profile.User.Username, &profile.User.Email, &profile.User.TimeoutUntil, &profile.User.Banned, &profile.User.IsActive, &profile.User.CreatedAt, &profile.User.UpdatedAt,
		&profile.User.Role.Level, &profile.User.Role.Description,
		&profile.User.Followers, &profile.User.Following,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProfileNotFound
		}
		return nil, fmt.Errorf("failed to get profile by username or email: %w", err)
	}

	return &profile, nil
}
