package auth

// Region defines supported user regions.
type Region string

const (
    RegionUzbekistan  Region = "uzbekistan"
    RegionRussia      Region = "russia"
    RegionTurkey      Region = "turkey"
    RegionSaudiArabia Region = "saudi_arabia"
    RegionUAE         Region = "uae"
    RegionOther       Region = "other"
)

// Visibility defines profile visibility settings.
type Visibility string

const (
    VisibilityPublic  Visibility = "public"
    VisibilityFriends Visibility = "friends"
    VisibilityPrivate Visibility = "private"
)

// User represents an authenticated account owner.
type User struct {
    ID                string         `json:"id"`
    Email             string         `json:"email"`
    FullName          string         `json:"fullName"`
    Username          string         `json:"username,omitempty"`
    PhoneNumber       string         `json:"phoneNumber,omitempty"`
    PasswordHash      string         `json:"passwordHash"`
    Bio               string         `json:"bio,omitempty"`
    Birthday          string         `json:"birthday,omitempty"`
    ProfileImage      string         `json:"profileImage,omitempty"`
    Region            Region         `json:"region"`
    PrimaryCurrency   string         `json:"primaryCurrency"`
    Visibility        Visibility     `json:"visibility,omitempty"`
    Preferences       UserPreferences `json:"preferences,omitempty"`
    IsEmailVerified   bool           `json:"isEmailVerified"`
    IsPhoneVerified   bool           `json:"isPhoneVerified"`
    LastLoginAt       string         `json:"lastLoginAt,omitempty"`
    CreatedAt         string         `json:"createdAt"`
    UpdatedAt         string         `json:"updatedAt"`
}

// UserPreferences captures configurable flags for a user.
type UserPreferences struct {
    ShowLevel     bool `json:"showLevel"`
    ShowAchievements bool `json:"showAchievements"`
    ShowStatistics bool `json:"showStatistics"`
    Language       string `json:"language"`
    Theme          string `json:"theme"`
    Notifications  NotificationPreferences `json:"notifications"`
}

// NotificationPreferences groups notification transport options.
type NotificationPreferences struct {
    Push      bool `json:"push"`
    Email     bool `json:"email"`
    Reminders bool `json:"reminders"`
}
