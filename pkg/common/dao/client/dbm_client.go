package client

type CoreInterface interface {
	ListenerGetter
}

type dbmClient struct {
}

func (m *dbmClient) Listener() ListenInterface {
	return newListener()
}

func NewDBMClient() CoreInterface {
	return &dbmClient{}
}
