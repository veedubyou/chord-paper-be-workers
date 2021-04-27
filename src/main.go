package main

import (
	"chord-paper-be-workers/src/application/store"
	"chord-paper-be-workers/src/application/usecase/spleet"
	"context"
	"fmt"
	"os"
)

func main() {
	remoteSplitTest()
}

func remoteSplitTest() {
	workingDir := "/Users/vincent/code/go/chord-paper-be-workers/wd"
	localUsecase := spleet.NewLocalFSSplitUsecase(workingDir, "/Users/vincent/miniconda3/bin/spleeter")
	jsonKey, err := os.ReadFile("/Users/vincent/code/go/chord-paper-be-workers/cloudkey.json")
	if err != nil {
		panic(err)
	}

	fileStore, err := store.NewGoogleFileStorage(string(jsonKey), "chord-paper-tracks")
	if err != nil {
		panic(err)
	}

	remoteUsecase := spleet.NewRemoteFSSplitUsecase(workingDir, fileStore, localUsecase)
	outputs, err := remoteUsecase.SplitTrack(context.Background(), "abcdefg", "https://storage.googleapis.com/chord-paper-tracks/tanpopo-test/original.mp3", spleet.FourStemSplitType)
	if err != nil {
		panic(err)
	}

	for key, output := range outputs {
		fmt.Println(key)
		fmt.Println(string(output.Payload))
	}
}

func localSplitTest() {
	usecase := spleet.NewLocalFSSplitUsecase("./wd", "/Users/vincent/miniconda3/bin/spleeter")
	contents, err := usecase.SplitTrack(context.Background(), "test-song-id", "/Users/vincent/spleets/songs/Ocean View/original/Ocean View.mp3", spleet.TwoStemSplitType)
	if err != nil {
		panic(err)
	}

	for k, _ := range contents {
		fmt.Println(k)
	}
}

func googleFileStoreTest() {
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
