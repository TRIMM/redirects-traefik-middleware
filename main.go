package main

func main() {
	var redirectsCh = make(chan []Redirect)
	var errCh = make(chan error)
	var redirectManager = NewRedirectManager()

	go fetchRedirectsOverChannel(redirectsCh, errCh)
	redirectManager.SyncRedirects(redirectsCh, errCh)
}
