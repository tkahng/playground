package apis

func bindGameApi(api *Api) {
	bindGetUserPlayerApi(api)
	bindPutUserPlayerApi(api)
}
