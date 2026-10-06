package main

import (
	"errors"
	"fmt"
	"os"

	"charm.land/huh/v2"
)

func main() {
	var selectLanguage string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Language.").
				Options(
					huh.NewOption("日本語", "ja"),
					huh.NewOption("English", "en"),
				).
				Value(&selectLanguage),
		),
	)

	err := form.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			fmt.Fprint(os.Stderr, "入力を中断しました。")
			return
		}
		fmt.Fprint(os.Stderr, "フォームの実行中にエラーが発生しました: ", err)
		return
	}

	fmt.Print(selectLanguage)
}
