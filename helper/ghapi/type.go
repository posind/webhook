package ghapi

type CommitDetails struct {
	Files []struct {
		Filename  string `json:"filename"`
		Additions int    `json:"additions"`
		Deletions int    `json:"deletions"`
	} `json:"files"`
}

type FileChangeInfo struct {
	Filename  string
	Additions int
	Deletions int
}
