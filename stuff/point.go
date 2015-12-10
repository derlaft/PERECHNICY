package stuff

import (
	"math"
)

type Point struct {
	X, Y int64
}

func Abs(a int64) int64 {
	if a < 0 {
		return -a
	}
	return a
}

func (p Point) Add(add Point) Point {
	return Point{p.X + add.X, p.Y + add.Y}
}

func sqrt64(a int64) int {
	return int(math.Sqrt(float64(a)))
}

func (p Point) Dist(q Point) int {
	return sqrt64((p.X-q.X)*(p.X-q.X) + (p.Y-q.Y)*(p.Y-q.Y))
}

func Min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func MinMaxPoint(p1, p2 Point) (min, max Point) {
	return Point{Min(p1.X, p2.X), Min(p1.Y, p2.Y)},
		Point{Max(p1.X, p2.X), Max(p1.Y, p2.Y)}
}

func EachPoint(p1, p2 Point) chan *Point {
	c := make(chan *Point)

	go func() {
		from, to := MinMaxPoint(p1, p2)

		for i := from.Y; i <= to.Y; i++ {
			for j := from.X; j <= to.X; j++ {
				c <- &Point{j, i}
			}
		}
		close(c)
	}()

	return c
}
