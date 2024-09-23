# Github Uploader

using net http
```go
func PostUploadGithub(w http.ResponseWriter, r *http.Request) {
	var respn itmodel.Response
	// _, err := watoken.Decode(config.PublicKeyWhatsAuth, helper.GetLoginFromHeader(req))
	// if err != nil {
	// 	respn.Info = helper.GetSecretFromHeader(req)
	// 	respn.Response = err.Error()
	// 	helper.WriteJSON(respw, http.StatusForbidden, respn)
	// 	return
	// }
	// Parse the form file
	_, header, err := r.FormFile("image")
	if err != nil {

		respn.Response = err.Error()
		helper.WriteJSON(w, http.StatusBadRequest, respn)
		return
	}

	//folder := ctx.Params("folder")
	folder := helper.GetParam(r)
	var pathFile string
	if folder != "" {
		pathFile = folder + "/" + header.Filename
	} else {
		pathFile = header.Filename
	}

	// save to github
	gh, err := atdb.GetOneDoc[model.Ghcreates](config.Mongoconn, "github", bson.M{})
	if err != nil {
		respn.Info = helper.GetSecretFromHeader(r)
		respn.Response = err.Error()
		helper.WriteJSON(w, http.StatusConflict, respn)
		return
	}

	content, _, err := ghupload.GithubUpload(gh.GitHubAccessToken, gh.GitHubAuthorName, gh.GitHubAuthorEmail, header, "alittifaq", "cdn", pathFile, false)
	if err != nil {
		respn.Info = "gagal upload github"
		respn.Response = err.Error()
		helper.WriteJSON(w, http.StatusEarlyHints, content)
		return
	}
	respn.Info = *content.Content.Name
	respn.Response = *content.Content.Path
	helper.WriteJSON(w, http.StatusOK, respn)

}
```