package main

import (
	"chord-paper-be-workers/src/application/store"
	"context"
	"fmt"
	"os"
)

func main() {
	ctx := context.Background()

	jsonKey, err := os.ReadFile("/Users/vincent/code/go/chord-paper-be-workers/cloudkey.json")
	if err != nil {
		panic(err)
	}

	fileStore, err := store.NewGoogleFileStorage(string(jsonKey), "chord-paper-tracks")
	if err != nil {
		panic(err)
	}

	fileContents, err := os.ReadFile("/Users/vincent/spleets/songs/yonawo - 蒲公英【OFFICIAL MUSIC VIDEO】/original/yonawo - 蒲公英【OFFICIAL MUSIC VIDEO】.mp3")
	if err != nil {
		panic(err)
	}

	fileURL, err := fileStore.WriteFile(ctx, "tanpopo-test/original.mp3", fileContents)
	if err != nil {
		panic(err)
	}

	fmt.Println(fileURL)
}
