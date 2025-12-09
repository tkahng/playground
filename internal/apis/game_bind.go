package apis

func bindGameApi(api *Api) {
	bindGameGetUserPlayerApi(api)
	bindGamePutUserPlayerApi(api)
}
