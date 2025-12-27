package database

type Status string

const (
	StatusUnread Status = "unread"
	StatusRead   Status = "read"
	StatusAny    Status = "" // don't use this in queries
)

func AllStatusValues() []Status {
	return []Status{
		StatusRead,
		StatusUnread,
	}
}
