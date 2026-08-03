package util

import "github.com/twpayne/go-geom"

func CreatePolygonFromArrays(xCoords, yCoords []int32) *geom.Polygon {
	if len(xCoords) == 0 || len(yCoords) == 0 || len(xCoords) != len(yCoords) {
		return nil
	}

	// Create a slice of []float64 for coordinates
	// Each point is represented as []float64{x, y}
	pointCount := len(xCoords)
	coords := make([][]float64, pointCount)

	for i := 0; i < pointCount; i++ {
		coords[i] = []float64{float64(xCoords[i]), float64(yCoords[i])}
	}

	return geom.NewPolygonFlat(geom.XY, FlattenCoords(coords), []int{len(coords) * 2})
}

func FlattenCoords(coords [][]float64) []float64 {
	flat := make([]float64, 0, len(coords)*2)
	for _, c := range coords {
		flat = append(flat, c...)
	}
	return flat
}

func ConvertCoordsToIntArray(coords [][]geom.Coord) (xCoords []int32, yCoords []int32) {
	for _, coord := range coords {
		for _, g := range coord {
			xCoords = append(xCoords, int32(g[0]))
			yCoords = append(yCoords, int32(g[1]))
		}
	}

	return xCoords, yCoords
}
