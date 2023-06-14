package main

func (a *application) newWalServer() error {
	cfg := &walConfig{}
	walsvr, err := newWalServer(cfg)
	if err != nil {
		return err
	}
	a.walServer = walsvr
	return nil
}
