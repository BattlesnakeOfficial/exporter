package engine

type Point struct {
	X int `json:"X"`
	Y int `json:"Y"`
}

type Death struct {
	Cause string `json:"Cause"`
	Turn  int    `json:"Turn"`
}

type Snake struct {
	ID     string  `json:"ID"`
	Name   string  `json:"Name"`
	Body   []Point `json:"Body"`
	Health int     `json:"Health"`

	Death *Death `json:"Death"`

	Color string `json:"Color"`    // Hex Code
	Head  string `json:"HeadType"` // https://github.com/BattlesnakeOfficial/board/tree/master/public/images/snake/head
	Tail  string `json:"TailType"` // https://github.com/BattlesnakeOfficial/board/tree/master/public/images/snake/tail
}

type GameFrame struct {
	Turn    int     `json:"Turn"`
	Food    []Point `json:"Food"`
	Snakes  []Snake `json:"Snakes"`
	Hazards []Point `json:"Hazards"`
}

type Game struct {
	ID     string `json:"ID"`
	Status string `json:"Status"`
	Width  int    `json:"Width"`
	Height int    `json:"Height"`
}

// API Response Structs //

type gameStatusResponse struct {
	Game Game `json:"Game"`
}

type gameFramesResponse struct {
	Count  int          `json:"count"`
	Frames []*GameFrame `json:"frames"`
}
