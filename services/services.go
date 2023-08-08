package services

type service struct{}

// Service is to provide settings configuration
var Service *service

// Init is initialization when the service started
func (service *service) Init() {
	// cfg := config.GetApplication()
}
