package devicesessions

// DeviceType enumerates supported device platforms.
type DeviceType string

const (
    DeviceTypeIOS     DeviceType = "ios"
    DeviceTypeAndroid DeviceType = "android"
    DeviceTypeWeb     DeviceType = "web"
)

// UserDevice tracks registered devices.
type UserDevice struct {
    ID             string     `json:"id"`
    UserID         string     `json:"userId"`
    DeviceID       string     `json:"deviceId"`
    DeviceName     string     `json:"deviceName,omitempty"`
    DeviceType     DeviceType `json:"deviceType"`
    OSVersion      string     `json:"osVersion,omitempty"`
    AppVersion     string     `json:"appVersion,omitempty"`
    PushToken      string     `json:"pushToken,omitempty"`
    LastActiveAt   string     `json:"lastActiveAt"`
    IsTrusted      bool       `json:"isTrusted"`
    CreatedAt      string     `json:"createdAt"`
}

// UserSession describes an active authentication session.
type UserSession struct {
    ID         string `json:"id"`
    UserID     string `json:"userId"`
    DeviceID   string `json:"deviceId"`
    TokenHash  string `json:"tokenHash"`
    IsActive   bool   `json:"isActive"`
    ExpiresAt  string `json:"expiresAt"`
    LastUsedAt string `json:"lastUsedAt"`
    CreatedAt  string `json:"createdAt"`
    RevokedAt  string `json:"revokedAt,omitempty"`
}
