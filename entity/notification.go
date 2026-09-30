package entity

type Notification struct {
	CartID  string
	Version int
	UserID  string
	/**
	notificationStatus                 string
	**/
	NotificationCountForCurrentVersion int
	NotificationCountTotal             int
	CommunicationChannel               CommunicationChannel
	// An reference to an identifier into content management system which can
	// store plain string, email or any other content.
	ContentId  string
	RetryCount int
	MaxRetries int
}
