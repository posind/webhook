package listnegara

type Negara struct {
	Nama string `bson:"nama" json:"nama"`
}

type ListNegaraResponse struct {
	Reply string `bson:"reply" json:"reply"`
}
