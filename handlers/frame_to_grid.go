package handlers

import (
	"fmt"
	"strconv"
	"strings"

	openapi "github.com/battlesnakeio/exporter/model"
)

// PixelType are the types of things on a snake board
type PixelType string

const (
	// Head head of a snake
	Head PixelType = "H"
	// Tail tail of a snake
	Tail PixelType = "T"
	// Body body of a snake
	Body PixelType = "B"
	// Space nothing here
	Space PixelType = " "
	// Food food on the board
	Food PixelType = "F"
)

// Pixel represents a graphical point on the board.
type Pixel struct {
	// ID if there is a snake this is it's id
	ID string
	// Colour if there is a snake, this is it's colour.
	Colour string
	// PixelType the type of pixel
	PixelType PixelType
	// Dead if there is a snake, is the snake dead.
	Dead bool
}

// ConvertFrameToGrid takes a frame and makes a 2d grid representatin.
func ConvertFrameToGrid(width int, height int, gameFrame *openapi.EngineGameFrame) [][]Pixel {
	response := make([][]Pixel, width)
	for x := 0; x < width; x++ {
		response[x] = make([]Pixel, height)
	}
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			response[x][y] = Pixel{PixelType: Space}
		}
	}
	for _, snake := range gameFrame.Snakes {
		for i, point := range snake.Body {
			response[point.X][point.Y] = Pixel{
				PixelType: getSegmentType(i, snake.Body),
				ID:        snake.ID,
				Colour:    snake.Color,
				Dead:      snake.Death.Cause != "",
			}
		}
	}
	for _, point := range gameFrame.Food {
		response[point.X][point.Y] = Pixel{
			PixelType: Food,
		}
	}
	return response
}

// ConvertFrameToMove takes a frame and makes a snake request.
func ConvertFrameToMove(gameFrame *openapi.EngineGameFrame, gameStatus *openapi.EngineStatusResponse, youID string) (*openapi.SnakeSnakeRequest, error) {
	deadSnakes := 0
	for _, snake := range gameFrame.Snakes {
		if snake.Death.Cause != "" {
			deadSnakes++
		}
	}
	response := &openapi.SnakeSnakeRequest{}
	response.Turn = gameFrame.Turn
	response.Game.Id = gameStatus.Game.ID
	response.Board.Width = gameStatus.Game.Width
	response.Board.Height = gameStatus.Game.Height
	response.Board.Food = make([]openapi.SnakeCoords, len(gameFrame.Food))
	response.Board.Snakes = make([]openapi.SnakeSnake, len(gameFrame.Snakes)-deadSnakes)
	for i, food := range gameFrame.Food {
		response.Board.Food[i] = openapi.SnakeCoords(food)
	}
	youIndex := -1
	k := 0
	for _, snake := range gameFrame.Snakes {
		if youID == snake.ID {
			youIndex = k
		}
		if snake.Death.Cause == "" {
			response.Board.Snakes[k] = openapi.SnakeSnake{}
			response.Board.Snakes[k].Id = snake.ID
			response.Board.Snakes[k].Health = snake.Health
			response.Board.Snakes[k].Name = snake.Name
			response.Board.Snakes[k].Body = make([]openapi.SnakeCoords, len(snake.Body))
			for j, body := range snake.Body {
				response.Board.Snakes[k].Body[j] = openapi.SnakeCoords(body)
			}
			k++
		}
	}
	if youIndex == -1 {
		return nil, fmt.Errorf("Couldn't find snake with id: %s in game: %s frame: %d.  Please set youId in the query string", youID, gameStatus.Game.ID, gameFrame.Turn)
	}
	response.You = response.Board.Snakes[youIndex]
	return response, nil
}

func getSegmentType(index int, body []openapi.EnginePoint) PixelType {
	if index == 0 {
		return Head
	}
	if index < len(body)-1 {
		return Body
	}
	return Tail
}

// ConvertGridToString takes a grid and turns it into a nice ascii picture
func ConvertGridToString(grid [][]Pixel) string {
	numberedSnakes := make(map[string]string)
	result := strings.Repeat("-", (len(grid[0])*2)+2) + "\n"
	for y := 0; y < len(grid); y++ {
		result += "|"
		for x := 0; x < len(grid[y]); x++ {
			if grid[x][y].PixelType == Food || grid[x][y].PixelType == Space {
				result += strings.Repeat(string(grid[x][y].PixelType), 2)
			} else {
				snakeIndex := convertToNumberedSnake(numberedSnakes, grid[x][y].ID)
				if grid[x][y].Dead {
					result += strings.Repeat(string(Space), 2)
				} else {
					result += string(grid[x][y].PixelType) + snakeIndex
				}
			}
		}
		result += "|\n"
	}
	result += strings.Repeat("-", (len(grid[0])*2)+2) + "\n"
	return result
}

func convertToNumberedSnake(numberedSnakes map[string]string, id string) string {
	index, exists := numberedSnakes[id]
	currentMax, maxExists := numberedSnakes["currentMax"]
	if !maxExists {
		currentMax = "0"
		numberedSnakes["currentMax"] = currentMax
	}
	if !exists {
		max, _ := strconv.Atoi(currentMax)
		numberedSnakes[id] = strconv.Itoa(max + 1)
		numberedSnakes["currentMax"] = strconv.Itoa(max + 1)
		index = numberedSnakes[id]
	}
	return index
}
