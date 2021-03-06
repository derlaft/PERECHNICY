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

func MinMaxPoint(p1, p2 Point) (min, max Point) {
	return Point{Min64(p1.X, p2.X), Min64(p1.Y, p2.Y)},
		Point{Max64(p1.X, p2.X), Max64(p1.Y, p2.Y)}
}

func EachPoint_(p1, p2 Point) []*Point {
	from, to := MinMaxPoint(p1, p2)
	ret := make([]*Point, 0, (to.X-from.X)*(to.Y-from.Y))

	for i := from.Y; i <= to.Y; i++ {
		for j := from.X; j <= to.X; j++ {
			ret = append(ret, &Point{j, i})
		}
	}
	return ret
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
