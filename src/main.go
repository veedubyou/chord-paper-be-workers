package main

import (
	"chord-paper-be-workers/src/application/store"
	"chord-paper-be-workers/src/application/usecase/split"
	"context"
	"fmt"
	"os"
)

func main() {
	remoteSplitTest()
}

func remoteSplitTest() {
	workingDir := "/Users/vincent/code/go/chord-paper-be-workers/wd"
	localUsecase, err := split.NewLocalFileSplitter(workingDir, "/Users/vincent/miniconda3/bin/spleeter")
	if err != nil {
		panic(err)
	}
	jsonKey, err := os.ReadFile("/Users/vincent/code/go/chord-paper-be-workers/cloudkey.json")
	if err != nil {
		panic(err)
	}

	fileStore, err := store.NewGoogleFileStorage(string(jsonKey))
	if err != nil {
		panic(err)
	}

	remoteUsecase, err := split.NewRemoteFileSplitter(workingDir, fileStore, localUsecase)
	if err != nil {
		panic(err)
	}

	songSplitUsecase := split.NewSongSplitter(remoteUsecase, "chord-paper-tracks")

	outputs, err := songSplitUsecase.HandleSplitTrack(context.Background(), "https://storage.googleapis.com/chord-paper-tracks/tanpopo-test/original.mp3", "tanpopo-2-test", "track-id", split.TwoStemSplitType)
	if err != nil {
		panic(err)
	}

	for key, output := range outputs {
		fmt.Println(key)
		fmt.Println(output)
	}
}
