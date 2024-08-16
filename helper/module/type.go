package module

type Module struct {
	Name         string   `json:"name,omitempty" bson:"name,omitempty"`
	Keyword      []string `json:"keyword,omitempty" bson:"keyword,omitempty"`
	Phonenumbers []string `json:"phonenumbers,omitempty" bson:"phonenumbers,omitempty"`
	Group        bool     `json:"group,omitempty" bson:"group,omitempty"`
	Personal     bool     `json:"personal,omitempty" bson:"personal,omitempty"`
}

type Typo struct {
	From string `json:"from,omitempty" bson:"from,omitempty"`
	To   string `json:"to,omitempty" bson:"to,omitempty"`
}
