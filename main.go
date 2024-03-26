package main

func main() {
	var redirectsCh = make(chan []Redirect)
	var errCh = make(chan error)

	go fetchRedirectsOverChannel(redirectsCh, errCh)
	syncRedirects(redirectsCh, errCh)
}
