package code

const (
	Success       = 0
	InvalidParam  = 10001
	UserNotFound  = 20001
	UserExists    = 20002
	DatabaseError = 30001
	RedisError    = 30002
	RateLimited   = 40001
	InternalError = 50000
)