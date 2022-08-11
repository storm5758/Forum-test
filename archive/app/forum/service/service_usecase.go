package service

func (s Service) Clear() error {
	return s.Repository.Clear()
}

func (s Service) Status() (Status, error) {
	return s.Repository.Status()
}
