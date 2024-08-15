package listcountry

type Country struct {
	Name  string `bson:"name" json:"name"`
}

type ListCountryResponse struct {
	Reply string `bson:"reply" json:"reply"`
}