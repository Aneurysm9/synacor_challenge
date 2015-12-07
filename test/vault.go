package main

import (
	"fmt"
	"math/rand"
	"errors"
)

var graph [4][4]int = [4][4]int{
	{0, 1, 9, 3},
	{2, 4, 1, 18},
	{4, 3, 11, 3},
	{3, 8, 1, 1},
}

func walk(distance int) ([][]int, error) {
	var pos [2]int = [2]int{0,0}
	var score int = 22
	var op int = 0
	moves := make([][]int, 0)

	for pos[0] != 3 || pos[1] != 3 {
		xory := getXorY()
		dir := getMove()
		newval := pos[xory] + dir
		if newval < 0 || newval > 3 || (newval == 0 && pos[1-xory] == 0) {
			continue
		}
		moves = append(moves, []int{xory, dir})
		if len(moves) > distance {
			return nil, errors.New("Too many moves")
		}
		pos[xory] = newval
		curval := graph[pos[0]][pos[1]]
		
		if op == 0 {
			op = curval
		} else {
			if op == 1 {
				score -= curval
			} else if op == 2 {
				score += curval
			} else if op == 3 {
				score *= curval
			}
			op = 0
		}
	}

	if op == 1 {
		score -= 1
	}

	if score == 30 {
		return moves, nil
	}

	return nil, errors.New("Wrong score")
}

func getXorY() int {
	if r := rand.Float32(); r > .5 {
		return 1
	}
	return 0
}

func getMove() int {
	if r := rand.Float32(); r > .5 {
		return 1
	}
	return -1
}

func main() {
	distance := 9999999
	var moves [][]int
	var err error
	for distance > 12 {
		moves, err = walk(distance)
		if err == nil {
			distance = len(moves)
		}
	}
	fmt.Println(moves)
}
