package entity

type User struct {
	Id       string
	Username string
	Name     string
	Admin    bool
}

type InsertUser struct {
	Id       string
	Username string
	Name     string
}

type Telegram struct {
	ChatId int64
	UserId int64
}

type TelegramUser struct {
	ChatId int64
	Admin  bool
}
