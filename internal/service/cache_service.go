package service

type CacheService struct{}

func NewCacheService() *CacheService {
	return &CacheService{}
}

func (s *CacheService) InvalidateUserNotifications(userID int64) error {
	return nil
}
