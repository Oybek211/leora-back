package achievements

// Achievement models a predefined goal.
type Achievement struct {
    ID          string `json:"id"`
    Key         string `json:"key"`
    Name        string `json:"name"`
    Description string `json:"description"`
    Category    string `json:"category"`
    Icon        string `json:"icon"`
    Color       string `json:"color"`
    XPReward    int    `json:"xpReward"`
    Requirement map[string]interface{} `json:"requirement"`
    IsSecret    bool   `json:"isSecret"`
    Tier        string `json:"tier"`
    Order       int    `json:"order"`
}

// UserAchievement tracks progress towards an achievement.
type UserAchievement struct {
    ID            string `json:"id"`
    UserID        string `json:"userId"`
    AchievementID string `json:"achievementId"`
    Progress      float64 `json:"progress"`
    UnlockedAt    string  `json:"unlockedAt,omitempty"`
    NotifiedAt    string  `json:"notifiedAt,omitempty"`
}

// UserLevel holds XP-specific data.
type UserLevel struct {
    ID         string `json:"id"`
    UserID     string `json:"userId"`
    Level      int    `json:"level"`
    CurrentXP  int    `json:"currentXP"`
    TotalXP    int    `json:"totalXP"`
    Title      string `json:"title,omitempty"`
    CreatedAt  string `json:"createdAt"`
    UpdatedAt  string `json:"updatedAt"`
}
