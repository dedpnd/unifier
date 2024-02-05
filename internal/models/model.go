package models

type User struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	Hash  string
}

type Rule struct {
	ID    int    `json:"id"`
	Rule  Config `json:"rule"`
	Owner *int   `json:"owner"`
}

type Config struct {
	TopicFrom    string         `json:"topicFrom"`
	Filter       Filter         `json:"filter"`
	EntityHash   []string       `json:"entityHash"`
	Unifier      []Unifier      `json:"unifier"`
	ExtraProcess []ExtraProcess `json:"extraProcess"`
	TopicTo      string         `json:"topicTo"`
}

type Filter struct {
	Regexp string `json:"regexp"`
}

type Unifier struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Expression string `json:"expression"`
}

type ExtraProcess struct {
	Func string `json:"func"`
	Args string `json:"args"`
	To   string `json:"to"`
}
