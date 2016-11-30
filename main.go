package main

/*
JSeed is a struct that copies the json AES-512 encrypted file contents that describes the operations required to get links from.
*/
type JSeed struct {
	SiteLink string `json:"siteLink"`
	SiteProtocol string `json:"siteProtocol"`
	SiteSignature string `json:"siteSignature"`
	FullListLinks []string `json:"fullListLinks"`
	FullListAcquire struct {
		         Under string `json:"under"`
		         LookFor []string `json:"lookFor"`
	         } `json:"fullListAcquire"`
	EpisodeListAcquire struct {
		         Under string `json:"under"`
		         LookFor []string `json:"lookFor"`
	         } `json:"episodeListAcquire"`
	VideoListAcquire struct {
		         Under string `json:"under"`
		         LookFor []string `json:"lookFor"`
	         } `json:"videoListAcquire"`
}

func main(){

}
