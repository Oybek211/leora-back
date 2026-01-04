package users

// Profile mirrors user profile data required by docs.
type Profile struct {
	ID              string `json:"id"`
	FullName        string `json:"fullName"`
	Email           string `json:"email"`
	Region          string `json:"region"`
	PrimaryCurrency string `json:"primaryCurrency"`
}
