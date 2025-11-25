package entity

type SystemConfig struct {
	ID                   uint `gorm:"primaryKey"`
	AIObjective          string
	FollowupIntervalMins int
	AutoReplyEnabled     bool
}
