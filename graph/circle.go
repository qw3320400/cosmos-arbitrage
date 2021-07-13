package graph

import "fmt"

type Pool struct {
	ID     uint64
	Denoms []string
}

func (p *Pool) String() string {
	if p == nil || len(p.Denoms) <= 0 {
		return "pool empty or nil"
	}
	return "[" + fmt.Sprintf("%d", p.ID) + ":(" + p.Denoms[0] + "|" + p.Denoms[1] + ")]"
}

type Circle struct {
	Path []*Pool
}

func findPath(pools []*Pool, inDenom, outDenom string, maxHops int, path []*Pool, circles []*Circle) []*Circle {
	for i, pool := range pools {
		if pool.Denoms[0] != inDenom && pool.Denoms[1] != inDenom {
			continue
		}
		var tmpOutDenom string
		if pool.Denoms[0] == inDenom {
			tmpOutDenom = pool.Denoms[1]
		} else {
			tmpOutDenom = pool.Denoms[0]
		}
		pathCopy := make([]*Pool, len(path))
		copy(pathCopy, path)
		pathCopy = append(pathCopy, pool)
		if tmpOutDenom == outDenom && len(pathCopy) > 2 {
			circles = append(circles, &Circle{Path: pathCopy})
		} else if maxHops > 1 && len(pools) > 1 {
			excludePool := make([]*Pool, len(pools)-1)
			copy(excludePool, pools[:i])
			copy(excludePool[i:], pools[i+1:])
			circles = findPath(excludePool, tmpOutDenom, outDenom, maxHops-1, pathCopy, circles)
		}
	}
	return circles
}

func FindCircle(pools []*Pool, denom string, maxHops int) []*Circle {
	var (
		circle = []*Circle{}
		path   = []*Pool{}
	)
	return findPath(pools, denom, denom, maxHops, path, circle)
}
