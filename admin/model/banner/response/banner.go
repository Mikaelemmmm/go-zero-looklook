package response

type GetBanner struct{
	Id      int64  `json:"id"`
	Title   string `son:"title"`
	Forward string `json:"forward"`
	Img     string `json:"img"`
}
