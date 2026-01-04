package settings

// UserSettings groups the configurable domains for a user.
type UserSettings struct {
    ID            string                 `json:"id"`
    UserID        string                 `json:"userId"`
    Theme         string                 `json:"theme"`
    Language      string                 `json:"language"`
    Notifications map[string]interface{} `json:"notifications"`
    Security      map[string]interface{} `json:"security"`
    AI            map[string]interface{} `json:"ai"`
    Focus         map[string]interface{} `json:"focus"`
    Privacy       map[string]interface{} `json:"privacy"`
    CreatedAt     string                 `json:"createdAt"`
    UpdatedAt     string                 `json:"updatedAt"`
}
