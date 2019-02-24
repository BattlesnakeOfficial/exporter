package engine

// type Game struct {
// 	ID           string `json:"ID"`
// 	Status       string `json:"Status"`
// 	Width        int32  `json:"Width"`
// 	Height       int32  `json:"Height"`
// 	SnakeTimeout int32  `json:"SnakeTimeout"`
// }

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

	Death Death `json:"Death"`

	Color string `json:"Color"`    // Hex Code
	Head  string `json:"HeadType"` // https://github.com/battlesnakeio/board/tree/master/public/images/snake/head
	Tail  string `json:"TailType"` // https://github.com/battlesnakeio/board/tree/master/public/images/snake/tail
}

type GameFrame struct {
	Turn   int     `json:"Turn"`
	Food   []Point `json:"Food"`
	Snakes []Snake `json:"Snakes"`
}

type Game struct {
	ID     string `json:"ID"`
	Status string `json:"Status"`
	Width  int    `json:"Width"`
	Height int    `json:"Height"`
}

type GameStatus struct {
}

// API Response Structs //

type gameStatusResponse struct {
	Game Game `json:"Game"`
}

type gameFramesResponse struct {
	Count  int          `json:"count"`
	Frames []*GameFrame `json:"frames"`
}

// type GameStatusResponse struct {
// 	Game      Game      `json:"game"`
// 	LastFrame GameFrame `json:"lastFrame"`
// }

// type DeathCause struct {
// 	// this records how the snake died, and is one of the 4 possible enum values
// 	Cause string `json:"Cause"`
// 	// this is the turn that the snake died on
// 	Turn int32 `json:"Turn"`
// }
