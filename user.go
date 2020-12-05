package main

type botUser struct {
	city string
	cat  string
	id   int64
}

func (u botUser) getLink() string {
	var link string

	link = hydraProxy + "catalog/" + u.cat + "?query=&region_id=" + u.city + "&subregion_id=0&price%5Bmin%5D=&price%5Bmax%5D=&unit=g&weight%5Bmin%5D=&weight%5Bmax%5D=&type=momental"

	return link
}
