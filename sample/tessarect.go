package main

import (
	"fmt"

	"github.com/otiai10/gosseract/v2"
)

func main() {
	// クライアント作成
	client := gosseract.NewClient()
	defer client.Close()

	// クライアント設定
	// client.SetLanguage("eng", "jpn") //jpnを指定するにはjpn.traineddataが必要
	client.SetImage("English-Class-Memes.jpg")

	// textのみ取得
	text, _ := client.Text()
	fmt.Printf("%s\n", text)

	// Level(BLOCK, PARA, TEXTLINE, WORD, SYMBOL)を指定して詳細を取得
	boxes, _ := client.GetBoundingBoxes(gosseract.RIL_SYMBOL)
	for _, box := range boxes {
		fmt.Printf("%+v\n", box)
	}
}
