package ds

type User struct {
	UserID      int    `gorm:"primaryKey;column:user_id"`
	Login       string `gorm:"type:varchar(150);unique;not null;column:login"`
	Password    string `gorm:"type:varchar(128);not null;column:password"`
	IsModerator bool   `gorm:"type:boolean;default:false;column:is_moderator"`
}
