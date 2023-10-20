package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"

	haversine "github.com/fcassin/gohaversine/haversine"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: haversine [uniform/cluster] [seed] [number of coordinate pairs to generate]")
		os.Exit(1)
	}

	var average float64
	var cluster bool
	var method string
	var number, seed int
	var err error

	method = os.Args[1]
	if method == "cluster" {
		cluster = true
	}

	seed, err = strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	number, err = strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	average, err = generate(cluster, int64(seed), number)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("          Method: %s\n", method)
	fmt.Printf("     Random seed: %d\n", seed)
	fmt.Printf("      Pair count: %d\n", number)
	fmt.Printf("Expected average: %f\n", average)
}

func generate(cluster bool, seed int64, number int) (average float64, err error) {
	var count int
	var sum float64
	var coordinateFile, answerFile *os.File

	coordinateFile, err = os.Create(fmt.Sprintf("coordinates_%d.json", number))
	if err != nil {
		return
	}
	defer coordinateFile.Close()

	answerFile, err = os.Create(fmt.Sprintf("answers_%d.f64", number))
	if err != nil {
		return
	}
	defer answerFile.Close()

	var source rand.Source = rand.NewSource(seed)
	var rng *rand.Rand = rand.New(source)

	var firstX, firstY, secondX, secondY float64
	var radiusX, radiusY float64 = 360, 180
	if cluster {
		firstX = (rng.Float64() - 0.5) * 360
		secondX = (rng.Float64() - 0.5) * 360
		firstY = (rng.Float64() - 0.5) * 180
		secondY = (rng.Float64() - 0.5) * 180

		radiusX = 10
		radiusY = 20
	}

	var bytes [8]byte
	var distance float64
	var coordinates haversine.Pairs
	coordinates.Pairs = make([]haversine.Pair, number)
	for count < number {
		coordinates.Pairs[count].X0 = (rng.Float64()-0.5)*radiusX + firstX
		coordinates.Pairs[count].X1 = (rng.Float64()-0.5)*radiusX + secondX
		coordinates.Pairs[count].Y0 = (rng.Float64()-0.5)*radiusY + firstY
		coordinates.Pairs[count].Y1 = (rng.Float64()-0.5)*radiusY + secondY

		distance = haversine.ReferenceHaversine(
			coordinates.Pairs[count].X0,
			coordinates.Pairs[count].X1,
			coordinates.Pairs[count].Y0,
			coordinates.Pairs[count].Y1,
			6372.8)

		// Convert distance to byte array
		binary.LittleEndian.PutUint64(bytes[:], math.Float64bits(distance))
		answerFile.Write(bytes[:])

		sum = sum + distance

		count++
	}

	average = sum / float64(number)

	var marshalledBytes []byte
	marshalledBytes, err = json.MarshalIndent(coordinates, "", "  ")
	coordinateFile.Write(marshalledBytes)

	return
}
