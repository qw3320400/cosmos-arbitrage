package graph

type Pool interface {
	GetID() uint64
	GetFirstDenom() string
	GetSecondDenom() string
}
type Circle struct {
	Path []uint64
}

func findPath(pools []Pool, inDenom, outDenom string, maxHops int, path []Pool, circles []*Circle) []*Circle {
	for i, pool := range pools {
		if pool.GetFirstDenom() != inDenom && pool.GetSecondDenom() != inDenom {
			continue
		}
		var tmpOutDenom string
		if pool.GetFirstDenom() == inDenom {
			tmpOutDenom = pool.GetSecondDenom()
		} else {
			tmpOutDenom = pool.GetFirstDenom()
		}
		pathCopy := make([]Pool, len(path))
		copy(pathCopy, path)
		pathCopy = append(pathCopy, pool)
		if tmpOutDenom == outDenom && len(pathCopy) > 2 {
			circles = append(circles, pathToCircle(pathCopy))
		} else if maxHops > 1 && len(pools) > 1 {
			excludePool := make([]Pool, len(pools)-1)
			copy(excludePool, pools[:i])
			copy(excludePool[i:], pools[i+1:])
			circles = findPath(excludePool, tmpOutDenom, outDenom, maxHops-1, pathCopy, circles)
		}
	}
	return circles
}

func pathToCircle(paths []Pool) *Circle {
	circle := &Circle{
		Path: []uint64{},
	}
	for _, path := range paths {
		circle.Path = append(circle.Path, path.GetID())
	}
	return circle
}

func FindCircle(pools []Pool, denom string, maxHops int) []*Circle {
	var (
		circle = []*Circle{}
		path   = []Pool{}
	)
	return findPath(pools, denom, denom, maxHops, path, circle)
}
